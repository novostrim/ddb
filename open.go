// Модуль ddb - универсальная библиотека для работы с разными sql серверами. Описывается структура базы, независимая от sql сервера,
// происходит автоматическая верификация структуры базы данных, прием и передача данных ведется через структуры пригодные для преобразования в json формат.
package ddb

import (
	"database/sql"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

func GetDriver(driver string) TDriver {
	adrivers := []string{"sqlite3", "mssql"}
	var i int
	for i = 0; i < len(adrivers); i++ {
		if adrivers[i] == driver {
			break
		}
	}
	return TDriver(i)
}

type TBase struct {
	BaseType     string
	Language     string
	LangNative   string
	Version      int
	Name         string
	BaseDescr    string
	Tables       []*Table
	Modules      []*TabItem
	MTables      map[string]*Table
	MModules     map[string]*TabItem
	Lang         map[string]string
	LangOriginal map[string]string
	LangCommon   map[string]string
	LangUndef    map[string]*Field
}

type Table struct {
	Name    string
	Descr   string
	Fields  []*Field
	Groups  []*Group
	Label   string
	TType   string `yaml:"type"`
	Fill    string
	MFields map[string]*Field
}

type TabItem struct {
	Table    string
	MainCard CardItem
}

type Field struct {
	Name      string `yaml:"name"`
	Type      string `yaml:"type"`
	Size      int    `yaml:"size"`
	Label     string `yaml:"label"`
	Clabel    string `yaml:"clabel"`
	Descr     string `yaml:"descr"`
	Req       int    `yaml:"req"`
	Link      string `yaml:"link"`
	LinkField string `yaml:"linkfield"`
	Disable   int    `yaml:"disable"`
}

type Group struct {
	Name   string `yaml:"name"`
	Type   string `yaml:"type"`
	Fields []*Field
}

type Card struct {
	Xxx string
}

type CardItem struct {
	Childs   []*CardItem
	Itemtype string
	//Type   TItemCard
	//Column string
}

var (
	Bases map[string]*TBase = make(map[string]*TBase)
)
var (
	pathdb    string
	pathdescr string
	basetype  = "dokido"
)

func init() {
	//pathdb = "db" //filepath.Join(GetAppPath(), "db")
}

//func Init(pathdescr,)

func (cn *Cnct) GetSqlType(f *Field) string {
	res := ""
	switch cn.Driver {
	case TD_Mssql:
		switch f.Type {
		case "char":
			res = fmt.Sprintf("VARCHAR(%d)", f.Size)
		case "int", "bool", "link":
			res = "INT"
		case "id":
			res = "INT IDENTITY(1,1) PRIMARY KEY"
		case "bookid":
			res = "INT PRIMARY KEY"
		case "num":
			res = "DECIMAL(15,3)"
		case "file":
			res = "VARBINARY"
		}
	case TD_Sqlite3:
		switch f.Type {
		case "char":
			res = fmt.Sprintf("CHAR(%d)", f.Size)
		case "int", "bool", "link":
			res = "INTEGER"
		case "id":
			res = "INTEGER PRIMARY KEY AUTOINCREMENT"
		case "bookid":
			res = "INTEGER PRIMARY KEY"
		case "num":
			res = "DECIMAL(15,3)"
		case "file":
			res = "BLOB"
		}
	}
	return res
	//if db.driver
}

var (
	Cns   map[string]*Cnct = make(map[string]*Cnct)
	curcn *Cnct
)

type TDriver int

const (
	TD_Sqlite3 TDriver = iota
	TD_Mssql
	TD_Unknown

	S_serv string = "serv"
)

type Cnct struct {
	//Db *db
	T  time.Time
	Id string

	Driver   TDriver
	Db       *sql.DB
	ConnStr  string
	BaseName string
	B        *TBase
	FLogin   bool
	IdUser   int
}

// open( driver string, connectionstring string, )
//OpenDB Открыть/создать базу данных
func OpenDB(driver string, basetype string, constring string, basename string) (*Cnct, error) {
	var err error
	var cn = Cnct{}
	Base, ok := Bases[basetype]
	if !ok {
		return nil, errors.New("Undefined base description")
	}

	Cns["0"] = &cn
	curcn = &cn

	cn.B = Base
	cn.BaseName = basename
	cn.Driver = GetDriver(driver)

	switch cn.Driver {
	case TD_Mssql:
		//cn.Db, err = sql.Open("mssql", "server=NOVOSTRIM\\SQLEXPRESS;user id=us1;password=123;database=avto")
		//fmt.Println("Open 1")
		cn.Db, err = sql.Open("mssql", constring) //"server=NOVOSTRIM;Trusted_Connection=Yes;database=x")
		if err != nil {
			//fmt.Println("Errror")
		}

		//fmt.Println("Open 2 ")
	case TD_Sqlite3:
		cn.Db, err = sql.Open("sqlite3", filepath.Join(constring, basename+".db"))
	default:
		err = errors.New("Unsupported SQL driver:" + driver)
	}
	if err == nil {
		err = cn.VeriBase(false)
	}

	/*if err != nil {
		golog.Fatal(err)
	}*/
	//defer db.DB.Close()
	return &cn, err
}

// LoadDescr начальная загрузка описания базы данных
func LoadDescr(basedescr io.Reader, basetype string) (*TBase, error) {
	var err error
	xmlBase := &TBase{}
	xmlBase.BaseType = basetype
	Bases[basetype] = xmlBase
	decoder := yaml.NewDecoder(basedescr)
	if err := decoder.Decode(xmlBase); err != nil {
		return nil, err
	}
	xmlBase.MTables = make(map[string]*Table)
	xmlBase.MModules = make(map[string]*TabItem)
	xmlBase.LangOriginal = make(map[string]string)
	/*xmlBase.LangUndef = make(map[string]*pb.PField)
	xmlBase.LangCommon = make(map[string]string)*/
	for _, tbl := range xmlBase.Tables {
		xmlBase.MTables[tbl.Name] = tbl

		//Добавление стандартных полей
		/*if tbl.TType == "book" {
			tbl.Fields = append(xd.Tables[1].Fields, tbl.Fields...)
		} else {
			tbl.Fields = append(xd.Tables[0].Fields, tbl.Fields...)
		}*/

		tbl.MFields = make(map[string]*Field)

		//xmlBase.AddRes("", "t_"+tbl.Name, tbl.Label)

		for _, fld := range tbl.Fields {
			tbl.MFields[fld.Name] = fld
		}

		for _, grp := range tbl.Groups {
			for _, fld := range grp.Fields {
				tbl.MFields[fld.Name] = fld
			}
		}
	}

	for _, mdl := range xmlBase.Modules {
		xmlBase.MModules[mdl.Table] = mdl
		//fmt.Println("Childs count: ", len(mdl.MainCard.Childs))
		//for _, it := range mdl.MainCard.Childs {
		//fmt.Println("It count: ", len(it.Childs))
		//}
	}
	//_, _, err = Normalize(xmlBase, xmlBase.Language)

	return xmlBase, err
}
