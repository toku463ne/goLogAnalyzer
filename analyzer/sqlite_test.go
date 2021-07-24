package analyzer

import (
	"database/sql"
	"fmt"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

func Test_sqlite(t *testing.T) {
	rootDir, err := ensureTestDir("sqlite")
	if err != nil {
		t.Errorf("%v", err)
		return
	}
	db, err := sql.Open("sqlite3", rootDir+"./foo.db")
	if err != nil {
		t.Errorf("%v", err)
		return
	}

	stmt, err := db.Prepare(`DROP TABLE IF EXISTS userinfo;`)
	if err != nil {
		t.Errorf("%v", err)
		return
	}
	_, err = stmt.Exec()
	if err != nil {
		t.Errorf("%v", err)
		return
	}

	stmt, err = db.Prepare(`CREATE TABLE userinfo (
		uid INTEGER PRIMARY KEY AUTOINCREMENT,
		username VARCHAR(64) NULL,
		departname VARCHAR(64) NULL,
		created DATE NULL
	);`)
	if err != nil {
		t.Errorf("%v", err)
		return
	}

	_, err = stmt.Exec()
	if err != nil {
		t.Errorf("%v", err)
		return
	}

	stmt, err = db.Prepare("INSERT INTO userinfo(username, departname, created) values(?,?,?)")
	if err != nil {
		t.Errorf("%v", err)
		return
	}

	res, err := stmt.Exec("astaxie", "kaihatu", "2012-12-09")
	if err != nil {
		t.Errorf("%v", err)
		return
	}

	id, err := res.LastInsertId()
	if err != nil {
		t.Errorf("%v", err)
		return
	}
	fmt.Println(id)

	stmt, err = db.Prepare("update userinfo set username=? where uid=?")
	if err != nil {
		t.Errorf("%v", err)
		return
	}

	res, err = stmt.Exec("astaxieupdate", id)
	if err != nil {
		t.Errorf("%v", err)
		return
	}
	affect, err := res.RowsAffected()
	if err != nil {
		t.Errorf("%v", err)
		return
	}

	fmt.Println(affect)

	rows, err := db.Query("SELECT * FROM userinfo")
	if err != nil {
		t.Errorf("%v", err)
		return
	}

	for rows.Next() {
		var uid int
		var username string
		var department string
		var created time.Time
		err = rows.Scan(&uid, &username, &department, &created)
		if err != nil {
			t.Errorf("%v", err)
			return
		}
		fmt.Println(uid)
		fmt.Println(username)
		fmt.Println(department)
		fmt.Println(created)
	}

	stmt, err = db.Prepare("delete from userinfo where uid=?")
	if err != nil {
		t.Errorf("%v", err)
		return
	}

	res, err = stmt.Exec(id)
	if err != nil {
		t.Errorf("%v", err)
		return
	}

	affect, err = res.RowsAffected()
	if err != nil {
		t.Errorf("%v", err)
		return
	}

	fmt.Println(affect)

	db.Close()
}
