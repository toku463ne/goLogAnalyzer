package analyzer

import (
	"fmt"
	"testing"
)

func TestMain_RunStats(t *testing.T) {
	testName := "TestMain_RunStats"
	testDir, err := ensureTestDir(testName + "/data")
	if err != nil {
		t.Errorf("%v", err)
		return
	}

	logPathRegex := fmt.Sprintf("%s/maintest.log*", testDir)
	minGapToRecord := 0.1
	linesInBlock := 5
	maxBlocks := 3

	if err := CleanupDB(testDir, false); err != nil {
		t.Errorf("%v", err)
		return
	}

	if _, err := copyFile("inputs/maintest.log",
		fmt.Sprintf("%s/maintest.log", testDir)); err != nil {
		t.Errorf("%v", err)
		return
	}

	RunRar(logPathRegex,
		testDir,
		"", "",
		minGapToRecord,
		0, linesInBlock, maxBlocks,
		false, false, true, false)

	a, err := loadAnalyzer(testDir)
	if err != nil {
		t.Errorf("%v", err)
		return
	}

	table := a.db.tables["logRecords"]
	cnt, err := table.count(nil, "*")
	if err != nil {
		t.Errorf("%v", err)
		return
	}
	if cnt != 11 {
		t.Errorf("logRecords incorrect")
		return
	}

	table = a.db.tables["countPerScore"]
	countPerScoreH := getCountPerScoreH()
	total := 0.0
	for _, colname := range countPerScoreH {
		s, err := table.sum(colname, nil, "*")
		if err != nil {
			t.Errorf("%v", err)
			return
		}
		total += s
	}
	if total != 12 {
		t.Errorf("countPerScore incorrect")
		return
	}

	a.close()

	if err := UpdateStats(testDir); err != nil {
		t.Errorf("%v", err)
		return
	}

	a, err = loadAnalyzer(testDir)
	if err != nil {
		t.Errorf("%v", err)
		return
	}

	table = a.db.tables["countPerScore"]
	total = 0.0
	for _, colname := range countPerScoreH {
		s, err := table.sum(colname, nil, "*")
		if err != nil {
			t.Errorf("%v", err)
			return
		}
		total += s
	}

	if total != 12 {
		t.Errorf("countPerScore incorrect")
		return
	}

	a.close()

}
