package logan

import (
	"goLogAnalyzer/pkg/csvdb"
)

type patternRelations struct {
	db        *csvdb.CircuitDB
	DataDir   string // directory for the pattern relations
	testMode  bool
	relations map[string][]string // patternKeyId -> relationKey
}

func newPatternRelations(dataDir string, useGzip, testMode bool) (*patternRelations, error) {
	pr := &patternRelations{
		relations: make(map[string][]string, 1000),
	}
	prdb, err := csvdb.NewCircuitDB(dataDir, "patternRelations",
		tableDefs["patternRelations"], 0, 0, 0, 0, useGzip)
	if err != nil {
		return nil, err
	}
	pr.db = prdb
	pr.testMode = testMode
	pr.DataDir = dataDir

	return pr, nil
}

func (pr *patternRelations) get(patternKeyId string) []string {
	if relations, ok := pr.relations[patternKeyId]; ok {
		return relations
	}
	return nil
}

func (pr *patternRelations) set(patternKeyId string, relationKey string) {
	if relations, ok := pr.relations[patternKeyId]; ok {
		pr.relations[patternKeyId] = append(relations, relationKey)
	} else {
		pr.relations[patternKeyId] = []string{relationKey}
	}
}

func (pr *patternRelations) reset() {
	// Reset the relations map
	pr.relations = make(map[string][]string, 1000)
}

// return len of relations for the given patternKeyId
func (pr *patternRelations) len(patternKeyId string) int {
	return len(pr.relations[patternKeyId])
}

func (pr *patternRelations) flush() error {
	if pr.DataDir == "" || pr.testMode {
		return nil
	}

	// insert the patternkeys into the CSV DB
	for patternKeyId, relations := range pr.relations {
		for _, relationKey := range relations {
			if err := pr.db.InsertRow(tableDefs["patternRelations"],
				patternKeyId, relationKey); err != nil {
				return err
			}
		}
	}

	// flush buffer to the block table file with WRITE mode (not APPEND)
	if err := pr.db.FlushOverwriteCurrentTable(); err != nil {
		return err
	}

	return nil
}

func (pr *patternRelations) commit(completed bool) error {
	if pr.DataDir == "" {
		return nil
	}
	if err := pr.flush(); err != nil {
		return err
	}

	if err := pr.db.UpdateBlockStatus(completed); err != nil {
		return err
	}
	return nil
}

func (pr *patternRelations) next(updated int64) error {
	if err := pr.flush(); err != nil {
		return err
	}
	if err := pr.db.NextBlock(updated); err != nil {
		return err
	}

	pr.relations = make(map[string][]string, 1000) // reset relations for the next block
	return nil
}

func (pr *patternRelations) loadAll() error {
	if pr.DataDir == "" || pr.testMode {
		return nil
	}
	pr.reset()
	rows, err := pr.db.SelectRows(nil, nil, tableDefs["patternRelations"])
	if err != nil {
		return err
	}
	for rows.Next() {
		var patternKeyId, relationKey string
		if err := rows.Scan(&patternKeyId, &relationKey); err != nil {
			return err
		}
		pr.set(patternKeyId, relationKey)
	}

	return nil
}
