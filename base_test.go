package ddb

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func TestLoadDescr(t *testing.T) {
	path, err := os.Getwd()
	if err != nil {
		t.Error("Getwd error", err)
		return
	}
	path = filepath.Join(path, "dtest")
	reader, err := os.Open(filepath.Join(path, "base.yaml"))
	if err != nil {
		t.Error("No description", err)
		return
	}
	defer reader.Close()

	sum, err := LoadDescr(reader, "dokido")
	if err != nil {
		t.Error("LoadDescr error", err)
	}
	fmt.Println(sum)
	cn, err := OpenDB("sqlite3", "dokido", path, "test")
	if err != nil {
		t.Error("OpenDB error", err)
	}
	fmt.Println(cn)
	tblresponse, err := cn.GetTable(&TableRequest{Table: "clients"})
	if err != nil {
		t.Error("GetTable error", err)
	}
	fmt.Println(tblresponse)

	editresponse, err := cn.Edit(&EditRequest{Table: "clients", Values: map[string]string{"name": "test"}})
	if err != nil {
		t.Error("Edit error", err)
	}
	fmt.Println(editresponse)

	rowresponse, err := cn.GetRow(&RowRequest{Table: "clients", ID: editresponse.ID})
	if err != nil {
		t.Error("Row error", err)
	}
	fmt.Println(rowresponse)

	deleterowresponse, err := cn.DeleteRow(&DeleteRowRequest{Table: "clients", ID: editresponse.ID})
	if err != nil {
		t.Error("Row error", err)
	}
	fmt.Println(deleterowresponse)

}
