package ddb

import (
	"database/sql"
	"errors"
)

/*
func Q(val string) string {
	return "'" + strings.ReplaceAll(val, "'", "''") + "'"
}*/
/*
func getBaseCnct() (*Cnct, error) {
	var (
		err error
		cn  *Cnct
	)
	cn, ok := Cns["0"]
	if !ok {
		err = errors.New("Base is't connected!")
	}
	return cn, err
}*/

func (cn *Cnct) getTable(tablename string) (*Table, error) {
	var err error
	table, ok := cn.B.MTables[tablename]
	if !ok {
		err = errors.New("Осутствует таблица " + tablename)
	}
	return table, err
}

func (cn *Cnct) getField(tablename string, fieldname string) (*Field, error) {
	var err error
	table, err := cn.getTable(tablename)
	if err != nil {
		return nil, err
	}
	field, ok := (*table).MFields[fieldname]
	if !ok {
		err = errors.New("Осутствует поле " + fieldname)
	}
	return field, err
}

func (cn *Cnct) getColumns(tablename string) (*[]TableColumn, string, error) {
	var (
		columns []TableColumn
		err     error
		scol    string
		table   *Table
	)
	table, err = cn.getTable(tablename)
	fields := table.Fields
	columns = make([]TableColumn, 0) //len(fields))

	for _, f := range fields {
		if f.Disable == 1 {
			continue
		}
		if len(scol) > 0 {
			scol += ", "
		}
		var column string
		if f.Type == "link" && len(f.Link) > 0 {
			column = "(SELECT name FROM " + f.Link + " WHERE id=" + tablename + "." + f.Name + ") "
		} else {
			column = f.Name
		}

		/*columns[j].Title = f.Label
		columns[j].Name = f.Name*/
		var t ColumnType
		s := 0
		switch f.Type {
		case "char":
			t = CTString
			s = f.Size
			break
		case "int", "bool", "link", "id":
			t = CTInteger
			break
		case "file":
			t = CTFile
			column = "length(" + column + ") as " + f.Name
			s = 10000
			break
		}

		scol += column
		/*columns[j].Type = t
		columns[j].Sortable = true
		columns[j].Filter = true*/
		columns = append(columns, TableColumn{Title: f.Label, Length: s, Name: f.Name, Type: t, Sortable: true, Filter: true})
	}

	return &columns, scol, err
}

func (cn *Cnct) getRow(tablename string, id string, bcard bool) (*[]TableColumn, *[]string, error) {
	var (
		err      error
		scol     string
		columns  *[]TableColumn
		table    *Table
		colcount int
	)

	columns, scol, err = cn.getColumns(tablename)
	colcount = len(*columns)
	if err != nil {
		return nil, nil, err
	}

	table, err = cn.getTable(tablename)
	if err != nil {
		return nil, nil, err
	}
	mfields := table.MFields

	rs := make([]string, colcount+2)
	if len(id) > 0 {
		rows, err := cn.Db.Query("SELECT  "+scol+", id, '' FROM "+tablename+" WHERE id=?", id)
		if err != nil {
			return nil, nil, err
		}

		defer rows.Close()
		vals := make([]interface{}, colcount+2)
		for i := range vals {
			vals[i] = new(sql.RawBytes)
		}
		if rows.Next() {
			err := rows.Scan(vals...)
			if err == nil {
				for i := 0; i < len(vals); i++ {
					if bcard && i < len(vals)-2 && len((*columns)[i].Name) > 0 && mfields[(*columns)[i].Name].Type == "file" {
						if len(string(*vals[i].(*sql.RawBytes))) > 0 {
							/*list := []FileInfo{FileInfo{File: "test", ID: "1", Size: 1000}}
							b, err := json.Marshal(list)
							if err == nil {
								rs[i] = string(b)
							}*/
						}
					} else {
						rs[i] = string(*vals[i].(*sql.RawBytes))
					}

				}
			}
		} else {
			err = errors.New("Запись " + id + " не существует или удалена")
		}
	} else {
		for i := 0; i < colcount+2; i++ {
			rs[i] = ""
		}
	}
	return columns, &rs, err
}
