package analyzer

import (
	"fmt"
	"os"

	"github.com/pkg/errors"
	csvdb "github.com/toku463ne/goCsvDb"
)

func getDBPath(rootDir, dbName string) string {
	return fmt.Sprintf("%s/%s", rootDir, dbName)
}

func newCsvdbObj(rootDir, dbName string) (*csvdbObj, error) {
	if err := ensureDir(rootDir); err != nil {
		return nil, err
	}
	dataDir := getDBPath(rootDir, dbName)
	cdb, err := csvdb.NewCsvDB(dataDir)
	if err != nil {
		return nil, err
	}
	db := new(csvdbObj)
	db.dataDir = dataDir
	db.CsvDB = cdb
	db.dbName = dbName
	return db, nil
}

func dropCsvDB(dataDir, dbName string) error {
	dbPath := getDBPath(dataDir, dbName)
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		return nil
	} else if err != nil {
		return errors.WithStack(err)
	}
	if err := os.RemoveAll(dbPath); err != nil {
		return errors.WithStack(err)
	}
	return nil
}

func (d *csvdbObj) count(tableName string, conditionCheckFunc func([]string) bool) int {
	t, err := d.GetTable(tableName)
	if err != nil {
		return -1
	}
	return t.Count(conditionCheckFunc)
}

func (d *csvdbObj) close() {
	d.CsvDB = nil
}
