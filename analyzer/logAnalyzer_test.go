package analyzer

import (
	"fmt"
	"testing"
	"time"
)

func TestLogAnalyzer_run1(t *testing.T) {

	iniV := new(logAnalyzerVars)
	iniV.name = "TestLogAnalyzer_run1"
	testDir, err := ensureTestDir(iniV.name)
	if err != nil {
		t.Errorf("%v", err)
		return
	}
	iniV.rootDir = fmt.Sprintf("%s/db", testDir)
	iniV.logPathRegex = fmt.Sprintf("%s/sample3.log*", testDir)
	iniV.linesInBlock = 5
	iniV.maxBlocks = 3
	iniV.useDB = true
	iniV.frequencyCheck = true
	iniV.minSupportPerBlock = 0.1
	verbose = true

	if _, err := copyFile("inputs/sample3.log.1",
		fmt.Sprintf("%s/sample3.log.1", testDir)); err != nil {
		t.Errorf("%v", err)
		return
	}
	time.Sleep(time.Second * 2)
	if _, err := copyFile("inputs/sample3.log",
		fmt.Sprintf("%s/sample3.log", testDir)); err != nil {
		t.Errorf("%v", err)
		return
	}

	a, err := newLogAnalyzerByVars(iniV)
	if err != nil {
		t.Errorf("%v", err)
		return
	}
	defer a.close()

	err = a.destroy()
	if err != nil {
		t.Errorf("%v", err)
		return
	}

	err = a.loadDB()
	if err != nil {
		t.Errorf("%v", err)
		return
	}

	cnt, err := a.rarAnal.countTargetLines()
	if err != nil {
		t.Errorf("%v", err)
		return
	}
	if cnt != 6 {
		t.Errorf("countTargetLines does not match!")
		return
	}
	if a.rarAnal.rowID != 0 {
		t.Errorf("rowID does not match!")
		return
	}

	if err := a.run(); err != nil {
		println(err)
		t.Errorf("%v", err)
		return
	}

	if a.rarAnal.rowNum != 5 {
		t.Errorf("rowNum is wrong")
		return
	}

	if a.rarAnal.rowID != 5 {
		t.Errorf("rowID does not match!")
		return
	}

	db := a.rarAnal.db

	table := db.tables["lastStatus"]
	v, err := table.select1rec(nil, "")
	if err != nil {
		t.Errorf("%v", err)
		return
	}
	if v[0] != "5" {
		t.Errorf("lastRowID is incorrect")
		return
	}
	if v[1] != "0" {
		t.Errorf("lastBlockID is incorrect")
		return
	}
	if v[3] != "2" {
		t.Errorf("lastRow is incorrect")
		return
	}

	table = db.tables["logBlocks"]
	blockIDf, v, err := table.max("blockID", nil, "*")
	if err != nil {
		t.Errorf("%v", err)
		return
	}

	if blockIDf != 0.0 {
		t.Errorf("blockID is incorrect")
		return
	}

	if v[2] != "5" {
		t.Errorf("blockCnt is incorrect")
		return
	}

	a.close()

	if _, err := copyFile("inputs/sample3_plus.log",
		fmt.Sprintf("%s/sample3.log", testDir)); err != nil {
		t.Errorf("%v", err)
		return
	}

	a, err = newLogAnalyzerByVars(iniV)
	if err != nil {
		t.Errorf("%v", err)
		return
	}
	defer a.close()

	err = a.loadDB()
	if err != nil {
		t.Errorf("%v", err)
		return
	}

	cnt, err = a.rarAnal.countTargetLines()
	if err != nil {
		t.Errorf("%v", err)
		return
	}
	if cnt != 6 {
		t.Errorf("countTargetLines does not match!")
		return
	}

	if a.rarAnal.rowNum != 0 {
		t.Errorf("rowNum is wrong")
		return
	}

	if a.rarAnal.rowID != 5 {
		t.Errorf("rowID does not match!")
		return
	}

	if err := a.run(); err != nil {
		println(err)
		t.Errorf("%v", err)
		return
	}

	if a.rarAnal.rowNum != 5 {
		t.Errorf("currCount is wrong")
		return
	}
	if a.rarAnal.rowID != 10 {
		t.Errorf("rowID does not match!")
		return
	}

	if ma, _, err := db.tables["items"].max("name", nil, "*"); err != nil {
		t.Errorf("currCount is wrong")
		return
	} else if ma != 10 {
		t.Errorf("Item is wrong")
		return
	}

	a.close()
}

/*
	db, err = newDB()
	if err != nil {
		t.Errorf("%v", err)
		return
	}

	v, err = db.select1row("logBlocks",
		[]string{"blockID", "lastRow", "lastEpoch"},
		[]string{fmt.Sprintf("logID=%d", a.logID),
			fmt.Sprintf("blockID = (select max(blockID) from logBlocks where logID=%d)", a.logID)})
	if v == nil || len(v) != 3 {
		t.Errorf("no rows")
		return
	}
	if v == nil || len(v) != 3 {
		t.Errorf("logBlocks table must have a row")
		return
	}

	if v[0] != "2" {
		t.Errorf("lastBlockID is incorrect")
		return
	}
	if v[1] != "7" {
		t.Errorf("lastRow is incorrect")
		return
	}

	file, _ := os.Stat("tmp/sample3.log")
	filesEpoch := file.ModTime().Unix()
	if v[2] != fmt.Sprintf("%d", filesEpoch) {
		t.Errorf("timestamps don't match")
		return
	}

	v, err = db.select1row("items",
		[]string{"cnt"},
		[]string{"logID=1", "blockID=2", "word='006'"})
	if err != nil {
		println(err)
		t.Errorf("%v", err)
		return
	}

	if v == nil || len(v) != 1 {
		t.Errorf("items table must have a row")
		return
	}
	if v[0] != "1" {
		t.Errorf("count is not correct")
		return
	}

	scores := make([]float64, 11)
	s := 0.0
	ss := 0.0
	blockID := 1
	for i := range scores {
		score := math.Log(float64(i+1)) + 1
		scores[i] = score
		s += score
		ss += score * score
		if i%5 == 4 {
			v, err = db.select1row("logBlocks",
				[]string{"scoreSum", "scoreSqrSum"},
				[]string{"logID=1", fmt.Sprintf("blockID=%d", blockID)})
			if err != nil {
				println(err)
				t.Errorf("%v", err)
				return
			}
			s = Round(s, 4)
			s1, _ := (strconv.ParseFloat(v[0], 64))
			s1 = Round(s1, 4)
			if s1 != s {
				t.Errorf("scoreSum don't match")
				return
			}
			ss = Round(ss, 4)
			ss1, _ := strconv.ParseFloat(v[1], 64)
			ss1 = Round(ss1, 4)
			if ss1 != ss {
				t.Errorf("scoreSqrSum don't match")
				return
			}
			s = 0
			ss = 0
			blockID++
		}
	}

	db.close()

	if err := a.tokenize(false); err != nil {
		t.Errorf("%v", err)
		return
	}
	a.close()
}
*/
