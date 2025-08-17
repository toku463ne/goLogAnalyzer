package logan

import (
	"goLogAnalyzer/pkg/csvdb"
)

type patternTags struct {
	db       *csvdb.CircuitDB
	DataDir  string // directory for the pattern tags
	testMode bool
	tags     map[string]map[string]string // patternKeyId -> tagName -> tagValue
}

func newPatternTags(dataDir string, useGzip, testMode bool) (*patternTags, error) {
	pt := &patternTags{
		tags: make(map[string]map[string]string, 1000),
	}
	ptdb, err := csvdb.NewCircuitDB(dataDir, "patternTags",
		tableDefs["patternTags"], 0, 0, 0, 0, useGzip)
	if err != nil {
		return nil, err
	}
	pt.db = ptdb
	pt.testMode = testMode
	pt.DataDir = dataDir

	return pt, nil
}

func (pt *patternTags) get(patternKeyId string) map[string]string {
	if tags, ok := pt.tags[patternKeyId]; ok {
		return tags
	}
	return nil
}

func (pt *patternTags) set(patternKeyId string, tagName, tagValue string) {
	if tags, ok := pt.tags[patternKeyId]; ok {
		tags[tagName] = tagValue
	} else {
		pt.tags[patternKeyId] = map[string]string{tagName: tagValue}
	}
}

func (pt *patternTags) reset() {
	// Reset the tags map
	pt.tags = make(map[string]map[string]string, 1000)
}

// return len of tags for the given patternKeyId
func (pt *patternTags) len(patternKeyId string) int {
	return len(pt.tags[patternKeyId])
}

func (pt *patternTags) flush() error {
	if pt.DataDir == "" || pt.testMode {
		return nil
	}

	// insert the patternkeys into the CSV DB
	for patternKeyId, tags := range pt.tags {
		for tagName, tagValue := range tags {
			if err := pt.db.InsertRow(tableDefs["patternTags"],
				patternKeyId, tagName, tagValue); err != nil {
				return err
			}
		}
	}

	// flush buffer to the block table file with WRITE mode (not APPEND)
	if err := pt.db.FlushOverwriteCurrentTable(); err != nil {
		return err
	}

	return nil
}

func (pt *patternTags) commit(completed bool) error {
	if pt.DataDir == "" {
		return nil
	}
	if err := pt.flush(); err != nil {
		return err
	}

	if err := pt.db.UpdateBlockStatus(completed); err != nil {
		return err
	}
	return nil
}

func (pt *patternTags) next(updated int64) error {
	if err := pt.flush(); err != nil {
		return err
	}
	if err := pt.db.NextBlock(updated); err != nil {
		return err
	}

	pt.tags = make(map[string]map[string]string, 1000) // reset tags for the next block
	return nil
}

func (pt *patternTags) loadAll() error {
	if pt.DataDir == "" || pt.testMode {
		return nil
	}
	pt.reset()
	rows, err := pt.db.SelectRows(nil, nil, tableDefs["patternTags"])
	if err != nil {
		return err
	}
	for rows.Next() {
		var patternKeyId, tagName, tagValue string
		if err := rows.Scan(&patternKeyId, &tagName, &tagValue); err != nil {
			return err
		}
		pt.set(patternKeyId, tagName, tagValue)
	}

	return nil
}
