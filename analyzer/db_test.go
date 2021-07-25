package analyzer

import (
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func Test_db01(t *testing.T) {
	dataDir, err := ensureTestDir("db")
	if err != nil {
		t.Errorf("%v", err)
		return
	}
	dbName := "testdb"

	err = dropDB(dataDir, dbName)
	if err != nil {
		t.Errorf("%v", err)
		return
	}

	d, err := newDB(dataDir, dbName)
	if err != nil {
		t.Errorf("%v", err)
		return
	}

	defer d.close()

	err = d.createTable("customer")
	if err != nil {
		t.Errorf("%v", err)
		return
	}

	tables, err := d.tables()
	if err != nil {
		t.Errorf("%v", err)
		return
	}

	if len(tables) != 1 {
		t.Errorf("number of tables. got=%d expected=%d", len(tables), 1)
	}

	cnt := d.count("customer", "")
	if cnt != 0 {
		t.Errorf("Count is not correct")
		return
	}

	_, err = d.exec("insert into customer(id, comment) values(1, 'comment1')")
	if err != nil {
		t.Errorf("%v", err)
		return
	}

	cnt = d.count("customer", "")
	if cnt != 1 {
		t.Errorf("Count is not correct")
		return
	}

	sqlstr := "select * from customer;"
	id := -1
	comment := ""
	err = d.select1rec(sqlstr, &id, &comment)
	if err != nil {
		t.Errorf("%v", err)
		return
	}

	if id != 1 {
		t.Errorf("Id is not correct")
		return
	}
	if comment != "comment1" {
		t.Errorf("comment is not correct")
		return
	}
}
