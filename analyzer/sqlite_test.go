package analyzer

import (
	"testing"
)

func Test_sqlite_execSQL(t *testing.T) {
	rootDir, err := ensureTestDir("sqlitest")
	if err != nil {
		t.Errorf("%+v", err)
	}

	db, err := newSqliteDB("test", rootDir)
	if err != nil {
		t.Errorf("%+v", err)
	}

	if err := db.open(); err != nil {
		t.Errorf("%+v", err)
	}

	defer db.close()

	if err := db.createTable("test"); err != nil {
		t.Errorf("Error creating table test : %v", err)
		return
	}

	if err := db.update("delete from test;"); err != nil {
		t.Errorf("%v", err)
		return
	}

	count, err := db.countTable("test", nil)
	if err != nil {
		t.Errorf("%v", err)
		return
	}

	if count != 0 {
		t.Errorf("Count is not as expected")
		return
	}

	err = db.update("insert into test(score) values(10.6);")
	if err != nil {
		t.Errorf("%v", err)
		return
	}

	count, err = db.countTable("test", nil)
	if err != nil {
		t.Errorf("%v", err)
		return
	}

	if count != 1 {
		t.Errorf("Count is not as expected")
		return
	}

	v, err := db.select1row("test", []string{"blockId"}, []string{"blockId = 3"})
	if err != nil {
		t.Errorf("%v", err)
		return
	}
	if v != nil {
		t.Errorf("There must be no data.")
		return
	}

	if err := db.update("drop table test;"); err != nil {
		t.Errorf("%v", err)
		return
	}

}
