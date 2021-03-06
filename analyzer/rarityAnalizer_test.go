package analyzer

import (
	"fmt"
	"math"
	"strconv"
	"testing"
	"time"
)

func getTestRarityAnalyzer(
	testDir, logPathRegex string,
	gapThreshold float64) (*rarityAnalyzer, error) {
	rootDir := fmt.Sprintf("%s/data", testDir)
	linesInBlock := 5
	maxBlocks := 3

	a, err := newRarityAnalyzer(logPathRegex,
		rootDir,
		"", "",
		gapThreshold,
		linesInBlock, maxBlocks, 3)

	return a, err
}

func TestRarityAnalyzer_run1(t *testing.T) {
	testName := "TestRarityAnalyzer_run1"
	testDir, err := ensureTestDir(testName)
	if err != nil {
		t.Errorf("%v", err)
		return
	}

	logPathRegex := fmt.Sprintf("%s/sample3.log*", testDir)
	gapThreshold := 0.5

	a, err := getTestRarityAnalyzer(testDir, logPathRegex, gapThreshold)
	if err != nil {
		t.Errorf("%v", err)
		return
	}
	defer a.close()
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

	if err := a.clean(); err != nil {
		t.Errorf("%v", err)
		return
	}

	if _, err := a.run(0); err != nil {
		println(err)
		t.Errorf("%v", err)
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
	if v[0] != "6" {
		t.Errorf("lastRowID is incorrect")
		return
	}
	if v[1] != "1" {
		t.Errorf("lastBlockID is incorrect")
		return
	}
	if v[3] != "3" {
		t.Errorf("lastRow is incorrect")
		return
	}

	table = db.tables["logBlocks"]
	blockIDf, v, err := table.max("blockID", nil, "*")
	if err != nil {
		t.Errorf("%v", err)
		return
	}

	if blockIDf != 1.0 {
		t.Errorf("blockID is incorrect")
		return
	}

	if v[2] != "1" {
		t.Errorf("blockCnt is incorrect")
		return
	}

	table = db.tables["logRecords"]
	cnt, err := table.count(nil, "*")
	if err != nil {
		t.Errorf("%v", err)
		return
	}
	if cnt != 6 {
		t.Errorf("logRecords incorrect")
		return
	}

	/*
		table = db.tables["nTopRareLogs"]
		cnt, err = table.count(nil, "*")
		if err != nil {
			t.Errorf("%v", err)
			return
		}
		if cnt != 6 {
			t.Errorf("nTopRareLogs incorrect")
			return
		}
	*/

	a.close()

	a.printCountPerGap(a.countPerGap, "Count per score")

	if a.countPerGap[0] != 1 {
		t.Errorf("countPerGap is incorrect")
		return
	}
	if a.countPerGap[1] != 5 {
		t.Errorf("countPerGap is incorrect")
		return
	}

	if _, err := copyFile("inputs/sample3_plus.log",
		fmt.Sprintf("%s/sample3.log", testDir)); err != nil {
		t.Errorf("%v", err)
		return
	}

	a, err = getTestRarityAnalyzer(testDir, logPathRegex, gapThreshold)
	if err != nil {
		t.Errorf("%v", err)
		return
	}
	defer a.close()

	if err := a.loadDB(); err != nil {
		t.Errorf("%v", err)
		return
	}

	if a.rowID != 6 {
		t.Errorf("rowID does not match!")
		return
	}

	if _, err := a.run(0); err != nil {
		println(err)
		t.Errorf("%v", err)
		return
	}

	if a.rowID != 11 {
		t.Errorf("rowID does not match!")
		return
	}

	if ma, _, err := db.tables["items"].max("name", nil, "*"); err != nil {
		t.Errorf("currCount is wrong")
		return
	} else if ma != 11 {
		t.Errorf("Item is wrong")
		return
	}

	table = db.tables["logRecords"]
	cnt, err = table.count(nil, "*")
	if err != nil {
		t.Errorf("%v", err)
		return
	}
	if cnt != 11 {
		t.Errorf("logRecords incorrect")
		return
	}

	/*
		table = db.tables["nTopRareLogs"]
		cnt, err = table.count(nil, "*")
		if err != nil {
			t.Errorf("%v", err)
			return
		}
		if cnt != 9 {
			t.Errorf("nTopRareLogs incorrect")
			return
		}
	*/

	a.printCountPerGap(a.countPerGap, "Count per score")

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

	a, err = getTestRarityAnalyzer(testDir, "", 0.0)
	if err != nil {
		t.Errorf("%v", err)
		return
	}
	defer a.close()
	a.run(0)

}

func TestRarityAnalyzer_run2_blocks(t *testing.T) {
	testItemsCount := func(testname string, items1 items,
		word string, expectedCnt int) bool {
		itemID := items1.getItemID(word)
		if itemID < 0 {
			if expectedCnt > 0 {
				t.Errorf("%s testItemsCount reason: %s", testname, word)
				return false
			}
			return true
		}
		cnt := items1.counts[itemID]
		if cnt != expectedCnt {
			t.Errorf("%s testItemsCount reason: %s", testname, word)
			return false
		}
		return true
	}

	testItemsExistance := func(testname string, items1 items,
		words []string, okIfExist bool) bool {
		for _, word := range words {
			itemID := items1.getItemID(word)
			if itemID < 0 {
				if okIfExist {
					t.Errorf("%s testItemsExistance reason: %s", testname, word)
				} else {
					continue
				}
			}
			cnt := items1.getCount(itemID)
			if okIfExist && cnt == 0 || !okIfExist && cnt > 0 {
				t.Errorf("%s testItemsExistance reason: %s", testname, word)
				return false
			}
		}
		return true
	}

	testName := "TestRarityAnalyzer_run2_blocks"
	testDir, err := ensureTestDir(testName)
	if err != nil {
		t.Errorf("%v", err)
		return
	}

	logPathRegex := fmt.Sprintf("%s/sample4.log*", testDir)
	gapThreshold := 0.0
	verbose = true

	a, err := getTestRarityAnalyzer(testDir, logPathRegex, gapThreshold)
	if err != nil {
		t.Errorf("%v", err)
		return
	}
	defer a.close()

	if _, err := copyFile("inputs/sample4.log",
		fmt.Sprintf("%s/sample4.log", testDir)); err != nil {
		t.Errorf("%v", err)
		return
	}

	if err := a.clean(); err != nil {
		t.Errorf("%v", err)
		return
	}

	if _, err := a.run(15); err != nil {
		println(err)
		t.Errorf("%v", err)
		return
	}
	if a.rowID != 15 {
		t.Errorf("rowNum is wrong")
		return
	}
	if a.scoreCount != 15 {
		t.Errorf("scoreCount is wrong")
		return
	}

	if !testItemsCount("test rowNum=15", *a.trans.items, "a001", 1) {
		return
	}
	if !testItemsCount("test rowNum=15", *a.trans.items, "a002", 2) {
		return
	}
	if !testItemsCount("test rowNum=15", *a.trans.items, "a005", 2) {
		return
	}
	if !testItemsCount("test rowNum=15", *a.trans.items, "a015", 2) {
		return
	}
	if !testItemsCount("test rowNum=15", *a.trans.items, "a016", 1) {
		return
	}

	if _, err := a.run(1); err != nil {
		println(err)
		t.Errorf("%v", err)
		return
	}
	if a.rowID != 16 {
		t.Errorf("rowNum is wrong")
		return
	}

	if a.scoreCount != 16 {
		t.Errorf("scoreCount is wrong")
		return
	}

	if testItemsExistance("test rowNum=16", *a.trans.items,
		[]string{"a001", "a006", "a016", "a017"}, true) == false {
		return
	}

	if _, err := a.run(4); err != nil {
		println(err)
		t.Errorf("%v", err)
		return
	}
	if a.rowID != 20 {
		t.Errorf("rowNum is wrong")
		return
	}

	if a.scoreCount != 15 {
		t.Errorf("scoreCount is wrong")
		return
	}

	if testItemsExistance("false rowNum=20", *a.trans.items,
		[]string{"a001", "a005"}, false) == false {
		return
	}
	if testItemsExistance("true rowNum=20", *a.trans.items,
		[]string{"a006", "a021"}, true) == false {
		return
	}
	if _, err := a.run(0); err != nil {
		println(err)
		t.Errorf("%v", err)
		return
	}
	if a.rowID != 35 {
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

	a, err = getTestRarityAnalyzer(testDir, logPathRegex, gapThreshold)
	if err != nil {
		t.Errorf("%v", err)
		return
	}
	defer a.close()

	if err = a.clean(); err != nil {
		t.Errorf("%v", err)
		return
	}

	procCnt, err := a.run(0)
	if err != nil {
		println(err)
		t.Errorf("%v", err)
		return
	}
	if procCnt != 35 {
		t.Errorf("Count is wrong")
		return
	}
	if a.rowID != 35 {
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

func TestRarityAnalyzer_run3_dontsave(t *testing.T) {

	testName := "TestRarityAnalyzer_run4_dontsave"
	testDir, err := ensureTestDir(testName)
	if err != nil {
		t.Errorf("%v", err)
		return
	}

	logPathRegex := fmt.Sprintf("%s/sample.txt", testDir)
	gapThreshold := 0.5

	a, err := getTestRarityAnalyzer(testDir, logPathRegex, gapThreshold)
	if err != nil {
		t.Errorf("%v", err)
		return
	}
	defer a.close()

	verbose = false

	a.xFilterRe = "melon"

	if _, err := copyFile("inputs/sample.txt",
		fmt.Sprintf("%s/sample.txt", testDir)); err != nil {
		t.Errorf("%v", err)
		return
	}

	if err := a.clean(); err != nil {
		t.Errorf("%v", err)
		return
	}

	cnt, err := a.run(0)
	if err != nil {
		println(err)
		t.Errorf("%v", err)
		return
	}

	if a.rowID != 6 {
		t.Errorf("rowID is wrong")
		return
	}

	if a.blocks[0] != nil {
		t.Errorf("blockCnt is wrong")
		return
	}

	if cnt != 1 {
		t.Errorf("processed rows is wrong")
		return
	}
}

func TestRarityAnalyzer_scanAndGetNTops(t *testing.T) {
	testName := "TestRarityAnalyzer_scanAndGetNTops"
	testDir, err := ensureTestDir(testName)
	if err != nil {
		t.Errorf("%v", err)
		return
	}

	logPathRegex := fmt.Sprintf("%s/sample5.log", testDir)
	gapThreshold := 0.5
	verbose = true

	a, err := getTestRarityAnalyzer(testDir, logPathRegex, gapThreshold)
	if err != nil {
		t.Errorf("%v", err)
		return
	}
	defer a.close()

	if _, err := copyFile("inputs/sample5.log",
		fmt.Sprintf("%s/sample5.log", testDir)); err != nil {
		t.Errorf("%v", err)
		return
	}

	if err := a.clean(); err != nil {
		t.Errorf("%v", err)
		return
	}

	_, err = a.run(0)
	if err != nil {
		println(err)
		t.Errorf("%v", err)
		return
	}

	nTop, err := a.scanAndGetNTops(2, "", "", nil)
	if len(nTop) != 2 {
		t.Errorf("nTop is wrong")
		return
	}

	nTop, err = a.scanAndGetNTops(5, "", "", nil)
	if err != nil {
		t.Errorf("%v", err)
		return
	}
	if len(nTop) != 5 {
		t.Errorf("nTop is wrong")
		return
	}

	nTop, err = a.scanAndGetNTops(0, "", "", nil)
	if err != nil {
		t.Errorf("%v", err)
		return
	}
	if len(nTop) != 3 {
		t.Errorf("nTop is wrong")
		return
	}

	nTop, err = a.scanAndGetNTops(5, "a006", "", nil)
	if err != nil {
		t.Errorf("%v", err)
		return
	}
	if nTop[0] == nil || nTop[1] == nil || nTop[2] != nil {
		t.Errorf("nTop is wrong")
		return
	}

	nTop, err = a.scanAndGetNTops(5, "", "a006", nil)
	if err != nil {
		t.Errorf("%v", err)
		return
	}
	if nTop[0] == nil || nTop[4] == nil {
		t.Errorf("nTop is wrong")
		return
	}
}
