package analyzer

import "os"

type csvDB struct {
	tables  map[string]*csvTable
	baseDir string
}

func newCsvDB(baseDir string, tableDefs map[string]csvTableDef) (*csvDB, error) {
	cdb := new(csvDB)
	cdb.tables = map[string]*csvTable{}
	for name, tableDef := range tableDefs {
		cdb.tables[name] = newCsvTable(name,
			tableDef.columns,
			baseDir,
			tableDef.maxPartitions)
	}
	cdb.baseDir = baseDir
	if _, err := os.Stat(baseDir); os.IsNotExist(err) {
		os.Mkdir(baseDir, 0755)
	} else if err != nil {
		return nil, err
	}
	return cdb, nil
}

func (db *csvDB) dropAllTables() error {
	for _, table := range db.tables {
		partitionID := "*"
		if table.maxPartitions == 0 {
			partitionID = ""
		}
		if err := table.drop(partitionID); err != nil {
			return err
		}
	}
	return nil
}
