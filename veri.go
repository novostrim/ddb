package ddb

import (
	"errors"
)

//IsEmptyBase проверяет на существование базы данных
func (cn *Cnct) IsEmptyBase() (bool, error) {
	var err error
	count := 0
	switch cn.Driver {
	case TD_Mssql:
		//DB.DB, err = sql.Open("mssql", "server=NOVOSTRIM\\SQLEXPRESS;user id=us1;password=123;database=avto")
		//fmt.Println("IsEmptyBase2")
		err = cn.Db.QueryRow("SELECT count(*) FROM INFORMATION_SCHEMA.COLUMNS").Scan(&count)
		if err != nil {
			//fmt.Println(err.Error())
		}
		//fmt.Println("IsEmptyBase3")
	case TD_Sqlite3:
		err = cn.Db.QueryRow("SELECT count(*) FROM sqlite_master").Scan(&count)
	}
	if err == nil && count == 0 {
		return true, err
	}
	//fmt.Println("IsEmptyBase4")
	return false, err
}

//VeriBase верифицирует базу данных, а также опционально может создать базу данных, если она пустая
func (cn *Cnct) VeriBase(bcreate bool) error {
	var version int = 0
	var err error
	var sq string
	switch cn.Driver {
	case TD_Mssql:
		sq = "SELECT COUNT(*) FROM information_schema.tables WHERE TABLE_NAME=?"
	case TD_Sqlite3:
		sq = "SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name=?"
	default:
	}

	bempty, err := cn.IsEmptyBase()

	if err != nil {
		return err
	}
	if bcreate {
		if !bempty {
			return errors.New("Невозможно создать структуру в не пустой базе данных")
		}
	} else {
		if bempty {
			bcreate = true
		} else {
			count := 0
			err = cn.Db.QueryRow(sq, S_serv).Scan(&count)
			if err != nil {
				return err
			}
			if count == 0 {
				return errors.New("Неподдерживаемая структура базы данных")
			}
			err = cn.Db.QueryRow("SELECT version FROM "+S_serv+" WHERE basename=?", cn.B.Name).Scan(&version)
			if err != nil {
				return err
			}
		}
	}

	if !bcreate {
		err = cn.Db.QueryRow("SELECT version FROM "+S_serv+" WHERE basename=?", cn.B.Name).Scan(&version)
		if err != nil {
			return err
		}
	}

	if version < cn.B.Version {
		tr, err := cn.Db.Begin()
		if err != nil {
			return err
		}

		for _, tbl := range cn.B.Tables {

			btblcreate := false
			count := 0

			err = tr.QueryRow(sq, tbl.Name).Scan(&count)
			if err != nil {
				break
			}

			if count == 0 {
				btblcreate = true
			}

			if btblcreate {
				query := ""
				for _, fld := range tbl.Fields {
					if len(query) > 0 {
						query += ", "
					}
					query += fld.Name + " " + cn.GetSqlType(fld)
				}
				query = "CREATE TABLE " + tbl.Name + " ( " + query + ")"

				_, err = tr.Exec(query)
				if err != nil {
					break
				}

			} else {
				query := "SELECT * FROM " + tbl.Name + " WHERE id=0"
				rows, err := tr.Query(query)
				if err != nil {
					break
				}
				cols, err := rows.Columns()
				if err != nil {
					break
				}
				mcols := make(map[string]bool)
				for _, col := range cols {
					mcols[col] = true
				}
				for _, fld := range tbl.Fields {
					_, ok := mcols[fld.Name]
					if !ok {
						_, err = tr.Exec("ALTER TABLE  " + tbl.Name + " ADD COLUMN " + fld.Name + " " + cn.GetSqlType(fld))
						if err != nil {
							break
						}
					}
				}

			}

			if tbl.TType == "book" && len(tbl.Fill) > 0 {

				_, err = tr.Exec("DELETE FROM " + tbl.Name)
				if err != nil {
					break
				}
				_, err = tr.Exec(tbl.Fill)
				if err != nil {
					break
				}
				tbl.Fill = ""
			}
		}
		if err == nil {
			if bcreate {
				sq = "INSERT INTO " + S_serv + "(version,basename) VALUES ( ?, ?)"
			} else {
				sq = "UPDATE " + S_serv + " SET version=? WHERE basename=?"
			}
			_, err = tr.Exec(sq, cn.B.Version, cn.B.Name)
		}

		if err != nil {
			tr.Rollback()
			return err
		}

		err = tr.Commit()
	}
	return err
}
