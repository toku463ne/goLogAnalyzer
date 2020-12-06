package analyzer

import (
	"fmt"
	"math"
	"strconv"
	"testing"
	"time"
)

func TestFileRarityAnalyzer_run1(t *testing.T) {
	testDir, err := ensureTestDir("TestFileRarityAnalyzer_run1")
	if err != nil {
		t.Errorf("%v", err)
		return
	}

	iniV := new(fileRarityAnalyzerVars)
	iniV.name = "TestFileRarityAnalyzer_run1"
	iniV.rootDir = fmt.Sprintf("%s/db", testDir)
	iniV.logPathRegex = fmt.Sprintf("%s/sample3.log*", testDir)
	iniV.linesInBlock = 5
	iniV.maxBlocks = 3
	iniV.rarityThreshold = 0.5
	iniV.useDB = true
	verbose = false

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

	a, err := newFileRarityAnalyzerByVars(iniV)
	if err != nil {
		t.Errorf("%v", err)
		return
	}
	defer a.close()

	if err := a.destroy(); err != nil {
		t.Errorf("%v", err)
		return
	}

	cnt, err := a.countTargetLines()
	if err != nil {
		t.Errorf("%v", err)
		return
	}
	if cnt != 6 {
		t.Errorf("countTargetLines does not match!")
		return
	}
	if a.rowID != 0 {
		t.Errorf("rowID does not match!")
		return
	}

	if _, err := a.run(0, -1); err != nil {
		println(err)
		t.Errorf("%v", err)
		return
	}

	if a.rowNum != 6 {
		t.Errorf("rowNum is wrong")
		return
	}

	if a.rowID != 6 {
		t.Errorf("rowID does not match!")
		return
	}

	db := a.db

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

	a, err = newFileRarityAnalyzerByVars(iniV)
	if err != nil {
		t.Errorf("%v", err)
		return
	}

	defer a.close()

	if err := a.loadDB(); err != nil {
		t.Errorf("%v", err)
		return
	}

	cnt, err = a.countTargetLines()
	if err != nil {
		t.Errorf("%v", err)
		return
	}
	if cnt != 6 {
		t.Errorf("countTargetLines does not match!")
		return
	}

	if a.rowNum != 0 {
		t.Errorf("rowNum is wrong")
		return
	}

	if a.rowID != 5 {
		t.Errorf("rowID does not match!")
		return
	}

	if _, err := a.run(0, -1); err != nil {
		println(err)
		t.Errorf("%v", err)
		return
	}

	if a.rowNum != 6 {
		t.Errorf("currCount is wrong")
		return
	}
	if a.rowID != 11 {
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

	scores := make([]float64, 11)
	s := 0.0
	ss := 0.0
	blockID := 0
	for i := range scores {
		score := math.Log(float64(i+1)) + 1
		scores[i] = score
		s += score
		ss += score * score
		if i%5 == 4 {
			v, err = db.tables["logBlocks"].select1rec(nil, fmt.Sprint(blockID))
			if err != nil {
				t.Errorf("%v", err)
				return
			}
			s = Round(s, 4)
			s1, _ := (strconv.ParseFloat(v[3], 64))
			s1 = Round(s1, 4)
			if s1 != s {
				t.Errorf("scoreSum don't match")
				return
			}
			ss = Round(ss, 4)
			ss1, _ := strconv.ParseFloat(v[4], 64)
			ss1 = Round(ss1, 4)
			if ss1 != ss {
				t.Errorf("scoreSqrSum don't match")
				return
			}
			s = 0
			ss = 0
			blockID++
		}

		a.close()
	}

}

// block test
func TestFileRarityAnalyzer_run2_blocks(t *testing.T) {
	testItemsCount := func(testname string, items1 items,
		words []string, okIfExist bool) bool {
		for _, word := range words {
			itemID, ok := items1.getItemID(word)
			if ok == false {
				if okIfExist {
					t.Errorf("%s testItemsCount reason: %s", testname, word)
				} else {
					continue
				}
			}
			cnt := items1.getCount(itemID)
			if okIfExist && cnt == 0 || !okIfExist && cnt > 0 {
				t.Errorf("%s testItemsCount reason: %s", testname, word)
				return false
			}
		}
		return true
	}

	testDir, err := ensureTestDir("TestFileRarityAnalyzer_run2_blocks")
	if err != nil {
		t.Errorf("%v", err)
		return
	}

	iniV := new(fileRarityAnalyzerVars)
	iniV.name = "TestFileRarityAnalyzer_run2_blocks"
	iniV.rootDir = fmt.Sprintf("%s/db", testDir)
	iniV.logPathRegex = fmt.Sprintf("%s/sample4.log", testDir)
	iniV.linesInBlock = 5
	iniV.maxBlocks = 3
	iniV.rarityThreshold = 0.5
	iniV.useDB = true
	verbose = false

	if _, err := copyFile("inputs/sample4.log",
		fmt.Sprintf("%s/sample4.log", testDir)); err != nil {
		t.Errorf("%v", err)
		return
	}

	a, err := newFileRarityAnalyzerByVars(iniV)
	if err != nil {
		t.Errorf("%v", err)
		return
	}
	defer a.close()

	if err := a.destroy(); err != nil {
		t.Errorf("%v", err)
		return
	}

	cnt, err := a.countTargetLines()
	if err != nil {
		t.Errorf("%v", err)
		return
	}
	if cnt != 35 {
		t.Errorf("countTargetLines does not match!")
		return
	}

	if _, err := a.run(15, -1); err != nil {
		println(err)
		t.Errorf("%v", err)
		return
	}
	if a.rowNum != 15 {
		t.Errorf("rowNum is wrong")
		return
	}
	if a.scoreCount != 15 {
		t.Errorf("scoreCount is wrong")
		return
	}

	if testItemsCount("test rowNum=15", *a.items,
		[]string{"a001", "a015", "a016"}, true) == false {
		return
	}

	if _, err := a.run(1, -1); err != nil {
		println(err)
		t.Errorf("%v", err)
		return
	}
	if a.rowNum != 16 {
		t.Errorf("rowNum is wrong")
		return
	}

	if a.scoreCount != 16 {
		t.Errorf("scoreCount is wrong")
		return
	}

	if testItemsCount("test rowNum=16", *a.items,
		[]string{"a001", "a006", "a016", "a017"}, true) == false {
		return
	}
	if _, err := a.run(4, -1); err != nil {
		println(err)
		t.Errorf("%v", err)
		return
	}
	if a.rowNum != 20 {
		t.Errorf("rowNum is wrong")
		return
	}

	if a.scoreCount != 15 {
		t.Errorf("scoreCount is wrong")
		return
	}

	if testItemsCount("false rowNum=20", *a.items,
		[]string{"a001", "a005"}, false) == false {
		return
	}
	if testItemsCount("true rowNum=20", *a.items,
		[]string{"a006", "a021"}, true) == false {
		return
	}
	if _, err := a.run(0, -1); err != nil {
		println(err)
		t.Errorf("%v", err)
		return
	}
	if a.rowNum != 35 {
		t.Errorf("rowNum is wrong")
		return
	}
	if a.scoreCount != 15 {
		t.Errorf("scoreCount is wrong")
		return
	}

	a.close()

	befScoreCount := a.scoreCount
	befScoreSum := a.scoreSum
	befScoreSqrSum := a.scoreSqrSum

	// restart
	a, err = newFileRarityAnalyzerByVars(iniV)
	if err != nil {
		t.Errorf("%v", err)
		return
	}
	defer a.close()

	if err = a.destroy(); err != nil {
		t.Errorf("%v", err)
		return
	}

	procCnt, err := a.run(0, -1)
	if err != nil {
		println(err)
		t.Errorf("%v", err)
		return
	}
	if procCnt != 35 {
		t.Errorf("Count is wrong")
		return
	}
	if a.rowNum != 35 {
		t.Errorf("rowNum is wrong")
		return
	}

	if befScoreCount != a.scoreCount {
		t.Errorf("scoreCount is wrong")
		return
	}
	if befScoreSum != a.scoreSum {
		t.Errorf("scoreSum is wrong")
		return
	}
	if befScoreSqrSum != a.scoreSqrSum {
		t.Errorf("scoreSqrSum is wrong")
		return
	}
	a.close()

}
