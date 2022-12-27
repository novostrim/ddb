package ddb

import (
	"database/sql"
	"strconv"
)

const ( //	ColumnType
	CTString = iota
	CTText
	CTLogic
	CTInteger
	CTDecimal
	CTMoney
	CTDate
	CTFile
)

type ColumnType int

const ( //	FilterType
	FTEqual = iota
	FTLess
	FTGreater
	FTEmpty
	FTBegin
	FTContain
	FTWeek
	FTMonth
	FTDays
	FTCheck
)

type FilterType int

type TableSettings struct {
	PerPage int
}

type TableRequest struct {
	Table   string       `json:"table"`
	Page    int          `json:"page,omitempty"`
	PerPage int          `json:"perpage,omitempty"`
	Sort    string       `json:"sort,omitempty"` // имя колонки сортировки
	Asc     bool         `json:"asc,omitempty"`
	Filter  []FilterInfo `json:"filter,omitempty"`
}

type FilterInfo struct {
	Column string     `json:"column"`
	Not    bool       `json:"not,omitempty"`
	Type   FilterType `json:"type"`
	Value  string     `json:"value,omitempty"`
	Prev   int        `json:"prev,omitempty"` // 0 - AND 1 - OR
}

type TableColumn struct {
	Title    string     `json:"title"`
	Name     string     `json:"name"`
	Type     ColumnType `json:"type"`
	Sortable bool       `json:"sortable,omitempty"`
	Sort     int        `json:"sort,omitempty"` // -1 или 1. 0 - нет сортировки по колонке
	Length   int        `json:"length,omitempty"`
	Hidden   bool       `json:"hidden,omitempty"`
	Align    int        `json:"align,omitempty"` // 0-left 1-center 2-right
	Filter   bool       `json:"filter,omitempty"`
}

type TableResponse struct {
	Title   string        `json:"title"`
	Columns []TableColumn `json:"columns"`
	Values  [][]string    `json:"values"`
	Count   int           `json:"count"`   // Общее количество записей
	PerPage int           `json:"perpage"` // Количество записей на страницу
	Page    int           `json:"page"`    // Текущая страница с 1
}

const Count = 1025

var (
	defaultSettings = TableSettings{
		PerPage: 25,
	}
	curSettings = defaultSettings
)

func (cn *Cnct) GetTable(request *TableRequest) (*TableResponse, error) {
	var (
		response TableResponse
		err      error
		scol     string
		colcount int
		columns  *[]TableColumn
	)

	page := 1
	if request.Page > 1 {
		page = request.Page
	}
	if request.PerPage > 0 {
		curSettings.PerPage = request.PerPage
	}

	response.PerPage = curSettings.PerPage
	response.Page = page
	limstart := (page - 1) * response.PerPage
	limcount := response.PerPage

	columns, scol, err = cn.getColumns(request.Table)
	if err == nil {
		colcount = len(*columns)
		for i := range *columns {
			(*columns)[i].Sort = 0
		}

		where := ""
		order := ""

		if len(request.Sort) > 0 {
			for i, v := range *columns {
				if v.Name == request.Sort {
					if len(order) > 0 {
						order += ", "
					}
					order += v.Name
					if request.Asc {
						(*columns)[i].Sort = -1
						order += " DESC "
					} else {
						(*columns)[i].Sort = 1
					}
					break
				}
			}
		}
		values := make([]interface{}, 0)

		for _, f := range request.Filter {
			field, err := cn.getField(request.Table, f.Column)
			if err == nil /*&& field.Name == f.Column*/ {
				wh := ""
				switch f.Type {
				case FTEqual:
					wh = field.Name + " = ?"
					values = append(values, f.Value)
					break
				case FTLess:
					wh = field.Name + " < ?"
					values = append(values, f.Value)
					break
				case FTGreater:
					wh = field.Name + " > ?"
					values = append(values, f.Value)
					break
				case FTEmpty:
					wh = field.Name + " IS NULL OR " + field.Name + " = ''"
					break
				case FTBegin:
					wh = field.Name + " LIKE( ? )"
					values = append(values, f.Value+"%")
					break
				case FTContain:
					wh = field.Name + " LIKE( ? )"
					values = append(values, "%"+f.Value+"%")
					break
				case FTWeek:
					break
				case FTMonth:
					break
				case FTDays:
					break
				case FTCheck:
					break
				}
				if len(wh) > 0 {
					if len(where) > 0 {
						where += " AND "
					}
					where += "(" + wh + ")"
				}
			}
		}

		response.Columns = *columns

		if len(where) > 0 {
			where = " WHERE " + where
		}
		if len(order) > 0 {
			order = " ORDER BY " + order
		}
		count := 0
		err := cn.Db.QueryRow("SELECT COUNT( * ) FROM "+request.Table+where, values...).Scan(&count)
		if err != nil {
			return nil, err
		}
		response.Count = count

		query := "SELECT  " + scol + ", id, '' FROM " + request.Table + where + order

		var (
			rows *sql.Rows
		)
		if limstart > 0 || limcount != 0 {
			if cn.Driver == TD_Mssql {
				query = "WITH num_row AS ( SELECT row_number() OVER (ORDER BY id) as num, " + scol + " FROM " + request.Table + ") " +
					"SELECT " + scol + ", id, '' FROM num_row WHERE num >= " + strconv.Itoa(limstart) + " AND num < " + strconv.Itoa(limstart+limcount)
			} else {
				query += " LIMIT " + strconv.Itoa(limstart) + "," + strconv.Itoa(limcount)
			}
		}
		//query := "SELECT  " + scol + ", id, '' FROM " + request.Table + where + order + " LIMIT " + strconv.Itoa(limstart) + "," + strconv.Itoa(limcount)
		rows, err = cn.Db.Query(query, values...)
		if err == nil {
			defer rows.Close()
			vals := make([]interface{}, colcount+2)
			for i := range vals {
				vals[i] = new(sql.RawBytes)
			}
			rs := make([][]string, 0)
			j := 0
			for rows.Next() {
				rs = append(rs, make([]string, colcount+2))
				err := rows.Scan(vals...)
				if err != nil {
					break
				}
				for i := 0; i < len(vals); i++ {
					rs[j][i] = string(*vals[i].(*sql.RawBytes))
				}
				j++
			}
			response.Values = rs
		}
	}
	return &response, nil
}
