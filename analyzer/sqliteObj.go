package analyzer

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"os"

	_ "github.com/mattn/go-sqlite3"
	"github.com/pkg/errors"
)

func getSqlitePath(dataDir, dbName string) string {
	return fmt.Sprintf("%s/%s.db", dataDir, dbName)
}

func newSqliteObj(dataDir, dbName string) (*sqliteObj, error) {
	if err := ensureDir(dataDir); err != nil {
		return nil, err
	}
	dbPath := getSqlitePath(dataDir, dbName)
	conn, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	db := new(sqliteObj)
	db.dataDir = dataDir
	db.conn = conn
	db.dbName = dbName
	return db, nil
}

func dropSqliteDB(dataDir, dbName string) error {
	dbPath := getSqlitePath(dataDir, dbName)
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		return nil
	} else if err != nil {
		return errors.WithStack(err)
	}
	if err := os.Remove(dbPath); err != nil {
		return errors.WithStack(err)
	}
	return nil
}

func (d *sqliteObj) close() error {
	err := d.conn.Close()
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}

func (d *sqliteObj) exec(sqlstr string) (sql.Result, error) {
	stmt, err := d.conn.Prepare(sqlstr)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return stmt.Exec()
}

func (d *sqliteObj) query(sqlstr string) (*sql.Rows, error) {
	stmt, err := d.conn.Prepare(sqlstr)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return stmt.Query()
}

func (d *sqliteObj) getSqlFileContents(sqlFile string) (string, error) {
	f, err := os.Open(fmt.Sprintf("%s", sqlFile))
	if err != nil {
		return "", errors.WithStack(err)
	}
	defer f.Close()
	b, err := ioutil.ReadAll(f)
	if err != nil {
		return "", errors.WithStack(err)
	}
	return string(b), nil
}

func (d *sqliteObj) execFromFile(sqlFile string) (sql.Result, error) {
	sqlstr, err := d.getSqlFileContents(sqlFile)
	if err != nil {
		return nil, err
	}
	return d.exec(sqlstr)
}

/*
func (d *sqliteObj) createTable(tableName string) error {
	sqlFile := fmt.Sprintf("%s/create_table_%s.sql", d.dbName, tableName)
	_, err := d.execFromFile(sqlFile)
	return err
}*/

func (d *sqliteObj) createTable(tableName string) error {
	_, err := d.exec(dbDefVar[d.dbName][tableName])
	if err != nil {
		return err
	}

	return err
}

func (d *sqliteObj) dropTable(tableName string) error {
	sqlstr := fmt.Sprintf("DROP TABLE IF EXISTS %s;", tableName)
	_, err := d.exec(sqlstr)
	return err
}

func (d *sqliteObj) recreateTable(tableName string) error {
	if err := d.dropTable(tableName); err != nil {
		return err
	}
	if err := d.createTable(tableName); err != nil {
		return err
	}
	return nil
}

func (d *sqliteObj) sum(tableName, colName, conds string) (float64, error) {
	sqlstr := fmt.Sprintf("SELECT SUM(%s) FROM %s ", colName, tableName)
	if conds != "" {
		sqlstr += fmt.Sprintf("WHERE %s", conds)
	}
	rows, err := d.query(sqlstr)
	if err != nil {
		return -1, err
	}
	defer rows.Close()
	if !rows.Next() {
		return -1, err
	}
	var s float64
	err = rows.Scan(&s)
	if err != nil {
		return -1, err
	}
	return s, nil
}

func (d *sqliteObj) count(tableName, conds string) int {
	sqlstr := fmt.Sprintf("SELECT COUNT(*) FROM %s ", tableName)
	if conds != "" {
		sqlstr += fmt.Sprintf("WHERE %s", conds)
	}
	rows, err := d.query(sqlstr)
	if err != nil {
		return -1
	}
	defer rows.Close()
	if !rows.Next() {
		return -1
	}
	cnt := -1
	err = rows.Scan(&cnt)
	if err != nil {
		return -1
	}
	return cnt
}

func (d *sqliteObj) select1rec(sqlstr string, a ...interface{}) error {
	rows, err := d.query(sqlstr)
	if err != nil {
		return err
	}
	defer rows.Close()
	if !rows.Next() {
		return errors.New("something went wrong")
	}
	err = rows.Scan(a...)
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}

func (d *sqliteObj) tables() ([]string, error) {
	var name string
	var tableNames []string
	rows, err := d.query("select name from sqlite_master where type = 'table';")
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		err = rows.Scan(&name)
		if err != nil {
			return nil, err
		}
		tableNames = append(tableNames, name)
	}
	return tableNames, nil
}
