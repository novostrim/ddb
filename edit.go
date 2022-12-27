package ddb

import (
	"database/sql"
	"strconv"
	"time"
)

type FileInfo struct {
	ID      string    `json:"id"`
	File    string    `json:"file"`
	Size    int       `json:"size"`
	Changed time.Time `json:"changed"`
	Comment string    `json:"comment,omitempty"`
	Bytes   []byte
}

type EditRequest struct {
	Table  string              `json:"table" form:"table"`
	ID     string              `json:"id" form:"id"`
	Form   string              `json:"form" form:"form"`
	Values map[string]string   `json:"values" form:"values"`
	Files  map[string]FileInfo `json:"files" form:"files"`
}

type EditResponse struct {
	ID          string            `json:"id"`
	Columns     []TableColumn     `json:"columns"`
	Values      []string          `json:"values"`
	TableValues map[string]string `json:"tablevalues,omitempty"`
}

func (cn *Cnct) Edit(request *EditRequest) (*EditResponse, error) {
	var (
		response EditResponse
		err      error
		id       string
	)

	bnew := false
	col1 := ""
	col2 := ""

	if len(request.ID) > 0 {
		id = request.ID
	} else {
		bnew = true
	}
	vallen := len(request.Values)
	if !bnew {
		vallen++
	}
	values := make([]interface{}, vallen)
	mfields := cn.B.MTables[request.Table].MFields
	i := 0
	for key, svalue := range request.Values {
		f, ok := mfields[key]
		if ok {
			if f.Type == "file" {
				continue
			}
			if len(col1) > 0 {
				col1 += ", "
				col2 += ", "
			}
			col1 += key
			if bnew {
				col2 += " ? "
			} else {
				col1 += " = ?"
			}
			values[i] = svalue
			i++
		}
	}
	for key, fvalue := range request.Files {
		mfields := cn.B.MTables[request.Table].MFields
		f, ok := mfields[key]
		if ok {
			if f.Type != "file" {
				continue
			}
			if len(col1) > 0 {
				col1 += ", "
				col2 += ", "
			}
			col1 += key
			if bnew {
				col2 += "?"
			} else {
				col1 += " = ?"
			}
			values[i] = fvalue.Bytes
			i++
		}
	}

	var res sql.Result
	if len(col1) > 0 {
		if bnew {
			res, err = cn.Db.Exec("INSERT INTO "+request.Table+"( "+col1+") VALUES ("+col2+")", values...)
			if err != nil {
				return nil, err
			}
			var newid64 int64
			newid64, err = res.LastInsertId()
			id = strconv.FormatInt(newid64, 10)
		} else {
			values[vallen-1] = id
			_, err = cn.Db.Exec("UPDATE "+request.Table+" SET "+col1+" WHERE id=?", values...)
			if err != nil {
				return nil, err
			}
		}
	}
	response.ID = id
	if err != nil {
		return nil, err
	}
	columns, rs, err := cn.getRow(request.Table, request.ID, true)
	_, rtable, err := cn.getRow(request.Table, request.ID, false)
	if err == nil {
		response.Values = *rs
		response.Columns = *columns
		if !bnew {
			tablevalues := make(map[string]string)
			for i, col := range *columns {
				tablevalues[col.Name] = (*rtable)[i]
			}
			response.TableValues = tablevalues
		}
		return &response, err
	}
	return nil, err
}
