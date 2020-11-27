package analyzer

import (
	"database/sql"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3" // blank import here because there is no main
	"github.com/pkg/errors"
)

var (
	sqliteErrCodeDBisLocket = 5
)

type sqliteDB struct {
	name    string
	baseDir string
	db      *sql.DB
	path    string
}

func newSqliteDB(name, baseDir string) (*sqliteDB, error) {
	s := new(sqliteDB)
	s.name = name
	s.baseDir = baseDir
	if _, err := os.Stat(baseDir); os.IsNotExist(err) {
		if err := os.Mkdir(baseDir, 0755); err != nil {
			return nil, errors.WithStack(err)
		}
	}
	s.path = fmt.Sprintf("%s/%s.db", baseDir, name)

	return s, nil
}

func (s *sqliteDB) open() error {
	db, err := sql.Open("sqlite3", s.path)
	if err != nil {
		return errors.WithStack(err)
	}
	s.db = db
	return nil
}

func (s *sqliteDB) close() error {
	return s.db.Close()
}

func (s *sqliteDB) reopen() error {
	s.close()
	/*
		// Workaround for "database is locked error"
		if err := os.Rename(s.path, s.path+".tmp"); err != nil {
				return errors.WithStack(err)
			} else {
				if err := os.Rename(s.path+".tmp", s.path); err != nil {
					return errors.WithStack(err)
				}
			}
	*/
	return s.open()
}

func (s *sqliteDB) update(queryStr string) error {
	err := s.reopen()
	if err != nil {
		return err
	}
	cnt := 0
	err = errors.New("dummy")
	for err != nil {
		_, err = s.db.Exec(queryStr)
		if err == nil {
			break
		}
		if cnt > sqliteDBLockWaitCnt {
			break
		}
		cnt++
		time.Sleep(1)
	}
	if err != nil {
		errors.WithStack(err)
	}
	return nil
}

func (s *sqliteDB) transaction(queryStr string) error {
	err := s.reopen()
	if err != nil {
		return err
	}
	cnt := 0
	err = errors.New("dummy")

	for err != nil {
		tx, err := s.db.Begin()
		if err != nil {
			return errors.WithStack(err)
		}
		stmt, err := tx.Prepare(queryStr)
		if err != nil {
			return extError(err, queryStr)
		}
		defer stmt.Close()
		if _, err := stmt.Exec(); err != nil {
			return extError(err, queryStr)
		}
		if cnt > sqliteDBLockWaitCnt {
			break
		}
		err = tx.Commit()
		if err == nil {
			break
		}
		stmt.Close()
		time.Sleep(1)
		cnt++
	}
	return err
}

func (s *sqliteDB) query(queryStr string) (*cursor, error) {
	rows, err := s.db.Query(queryStr)
	if err != nil {
		return nil, extError(err, queryStr)
	}

	cols, err := rows.Columns()
	if err != nil {
		return nil, extError(err, queryStr)
	}
	cur := new(cursor)
	v := make([]sql.RawBytes, len(cols))
	cur.scanArgs = make([]interface{}, len(v))
	for i := range v {
		cur.scanArgs[i] = &v[i]
	}
	cur.v = v
	cur.cols = cols
	cur.rows = rows
	return cur, nil
}

func (s *sqliteDB) selectTable(tableName string,
	cols []string,
	conditions []string) (*cursor, error) {
	colstr := strings.Join(cols, ",")
	whereStr := cond2WhereStr(conditions)
	q := fmt.Sprintf("select %s from %s %s;",
		colstr,
		tableName,
		whereStr)
	cur, err := s.query(q)
	return cur, errors.WithStack(err)
}

func (s *sqliteDB) select1row(tableName string,
	cols []string,
	conditions []string) ([]string, error) {

	cur, err := s.selectTable(tableName, cols, conditions)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	if cur.next() {
		if cur.err() == nil {
			v := cur.values()
			return v, nil
		}
		err = cur.err()
	}
	err = cur.err()
	if err != nil {
		err = errors.WithStack(err)
	}
	cur.close()
	return nil, err
}

func (s *sqliteDB) createTable(tableName string) error {
	q, ok := dbTablesDefs[tableName]
	if ok == false {
		return fmt.Errorf("The table %s is not defined in dbTables", tableName)
	}
	err := s.update(q)
	return err
}

func (s *sqliteDB) countTable(tableName string, conditions []string) (int, error) {
	v, err := s.select1row(tableName, []string{"count(*)"}, conditions)
	if err != nil {
		return -1, err
	}
	if len(v) != 1 {
		return -1, fmt.Errorf("Unexpected value length expected=1 got=%d", len(v))
	}
	n, err := strconv.Atoi(v[0])
	if err != nil {
		return -1, fmt.Errorf("%v", err)
	}
	return n, nil
}

func (s *sqliteDB) dropTable(tableName string) error {
	q := fmt.Sprintf("drop table if exists %s;", tableName)
	if err := s.update(q); err != nil {
		return err
	}
	return nil
}

func (s *sqliteDB) dropAllTables() error {
	cnt := len(dbTables)
	delTables := make([]string, cnt)

	i := cnt - 1
	for _, tableName := range dbTables {
		delTables[i] = tableName
		i--
	}

	for _, tableName := range delTables {
		if err := s.dropTable(tableName); err != nil {
			return err
		}
	}
	return nil
}

func (s *sqliteDB) tables() ([]string, error) {
	var t []string
	cur, err := s.query(".tables")
	if err != nil {
		return nil, err
	}
	for cur.next() {
		v := cur.values()
		t = append(t, v[0])
	}
	cur.close()

	return t, nil
}
