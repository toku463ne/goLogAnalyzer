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
	iniV.minSupportPerBlock = 0.1
	iniV.absenceCheck = true
	iniV.rarityThreshold = 0.5
	iniV.absenceThreshold = 0.5
	iniV.logLevel = "debug"

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

	if err := a.run(0); err != nil {
		println(err)
		t.Errorf("%v", err)
		return
	}

	if a.rarAnal.rowNum != 6 {
		t.Errorf("rowNum is wrong")
		return
	}

	if a.rarAnal.rowID != 6 {
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

	db2 := a.dciDB
	table = db2.tables["closedItemSets"]
	cnt, err = table.count(nil, "*")
	if err != nil {
		t.Errorf("%v", err)
		return
	}
	if cnt != 5 {
		t.Errorf("count of closedItemSets is not correct")
		return
	}
	table = db2.tables["closedItemKeys"]
	cnt, err = table.count(nil, "*")
	if err != nil {
		t.Errorf("%v", err)
		return
	}
	if cnt != 5 {
		t.Errorf("count of closedItemKeys is not correct")
		return
	}

	db3 := a.absAnal.db
	table = db3.tables["lastStatus"]
	v, err = table.select1rec(nil, "")
	if err != nil {
		t.Errorf("%v", err)
		return
	}
	if v[0] != "0" {
		t.Errorf("lastRowID is incorrect")
		return
	}
	if v[1] != "0" {
		t.Errorf("lastBlockID is incorrect")
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

	if err := a.run(0); err != nil {
		println(err)
		t.Errorf("%v", err)
		return
	}

	if a.rarAnal.rowNum != 6 {
		t.Errorf("currCount is wrong")
		return
	}
	if a.rarAnal.rowID != 11 {
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

	db2 = a.dciDB
	table = db2.tables["closedItemSets"]
	cnt, err = table.count(nil, "*")
	if err != nil {
		t.Errorf("%v", err)
		return
	}
	if cnt != 10 {
		t.Errorf("count of closedItemSets is not correct")
		return
	}
	table = db2.tables["closedItemKeys"]
	cnt, err = table.count(nil, "*")
	if err != nil {
		t.Errorf("%v", err)
		return
	}
	if cnt != 10 {
		t.Errorf("count of closedItemKeys is not correct")
		return
	}

	db3 = a.absAnal.db
	table = db3.tables["lastStatus"]
	v, err = table.select1rec(nil, "")
	if err != nil {
		t.Errorf("%v", err)
		return
	}
	if v[0] != "5" {
		t.Errorf("lastRowID is incorrect")
		return
	}
	if v[1] != "1" {
		t.Errorf("lastBlockID is incorrect")
		return
	}

	a.close()
}

func TestLogAnalyzer_run2_blocks(t *testing.T) {
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

	iniV := new(logAnalyzerVars)
	iniV.name = "TestLogAnalyzer_run2_blocks"
	testDir, err := ensureTestDir(iniV.name)
	if err != nil {
		t.Errorf("%v", err)
		return
	}
	iniV.rootDir = fmt.Sprintf("%s/db", testDir)
	iniV.logPathRegex = fmt.Sprintf("%s/sample4.log*", testDir)
	iniV.linesInBlock = 5
	iniV.maxBlocks = 3
	iniV.useDB = true
	iniV.minSupportPerBlock = 0.1
	iniV.absenceCheck = true
	iniV.rarityThreshold = 0.5
	iniV.absenceThreshold = 0.5

	verbose = false

	if _, err := copyFile("inputs/sample4.log",
		fmt.Sprintf("%s/sample4.log", testDir)); err != nil {
		t.Errorf("%v", err)
		return
	}

	a, err := newLogAnalyzerByVars(iniV)
	if err != nil {
		t.Errorf("%v", err)
		return
	}
	defer a.close()

	if err := a.destroy(); err != nil {
		t.Errorf("%v", err)
		return
	}

	if err := a.run(15); err != nil {
		println(err)
		t.Errorf("%v", err)
		return
	}

	if a.rarAnal.rowNum != 15 {
		t.Errorf("currCount is wrong")
		return
	}

	if testItemsCount("frq true rowNum=15", *a.frqAnal.items,
		[]string{"a005", "a010", "a015"}, true) == false {
		return
	}

	if testItemsCount("frq false rowNum=15", *a.frqAnal.items,
		[]string{"a001", "a016"}, false) == false {
		return
	}

	if testItemsCount("abs true rowNum=15", *a.absAnal.items,
		[]string{"a002", "a005", "a010"}, true) == false {
		return
	}

	if testItemsCount("abs false rowNum=15", *a.absAnal.items,
		[]string{"a001", "a011", "a015"}, false) == false {
		return
	}

	if err := a.run(1); err != nil {
		println(err)
		t.Errorf("%v", err)
		return
	}

	if a.rarAnal.rowNum != 16 {
		t.Errorf("currCount is wrong")
		return
	}

	if testItemsCount("frq true rowNum=16", *a.frqAnal.items,
		[]string{"a002", "a005", "a010", "a015"}, true) == false {
		return
	}
	if testItemsCount("frq false rowNum=16", *a.frqAnal.items,
		[]string{"a001", "a016"}, false) == false {
		return
	}

	if testItemsCount("abs true rowNum=16", *a.absAnal.items,
		[]string{"a002", "a005", "a010"}, true) == false {
		return
	}
	if testItemsCount("abs false rowNum=16", *a.absAnal.items,
		[]string{"a001", "a011", "a016", "a017"}, false) == false {
		return
	}

	if err := a.run(4); err != nil {
		println(err)
		t.Errorf("%v", err)
		return
	}

	if a.rarAnal.rowNum != 20 {
		t.Errorf("currCount is wrong")
		return
	}

	if a.rarAnal.scoreCount != 15 {
		t.Errorf("scoreCount is wrong")
		return
	}

	if testItemsCount("frq true rowNum=20", *a.frqAnal.items,
		[]string{"a007", "a015", "a020"}, true) == false {
		return
	}
	if testItemsCount("frq false rowNum=20", *a.frqAnal.items,
		[]string{"a006", "a021"}, false) == false {
		return
	}
	if testItemsCount("abs true rowNum=20", *a.absAnal.items,
		[]string{"a005", "a010"}, true) == false {
		return
	}
	if testItemsCount("abs false rowNum=20", *a.absAnal.items,
		[]string{"a001", "a011", "a016", "a017"}, false) == false {
		return
	}

	if err := a.run(5); err != nil {
		println(err)
		t.Errorf("%v", err)
		return
	}

	if a.rarAnal.rowNum != 25 {
		t.Errorf("currCount is wrong")
		return
	}

	if a.rarAnal.scoreCount != 15 {
		t.Errorf("scoreCount is wrong")
		return
	}

	if testItemsCount("frq true rowNum=25", *a.frqAnal.items,
		[]string{"a012", "a025"}, true) == false {
		return
	}
	if testItemsCount("frq false rowNum=25", *a.frqAnal.items,
		[]string{"a010", "a011", "a009"}, false) == false {
		return
	}
	if testItemsCount("abs true rowNum=25", *a.absAnal.items,
		[]string{"a007", "a020"}, true) == false {
		return
	}
	if testItemsCount("abs false rowNum=25", *a.absAnal.items,
		[]string{"a001", "a006", "a021"}, false) == false {
		return
	}

	a.close()
}

func TestLogAnalyzer_run3_middle(t *testing.T) {

	iniV := new(logAnalyzerVars)
	iniV.name = "TestLogAnalyzer_run3_middle"
	testDir, err := ensureTestDir(iniV.name)
	if err != nil {
		t.Errorf("%v", err)
		return
	}
	iniV.rootDir = fmt.Sprintf("%s/db", testDir)
	iniV.logPathRegex = fmt.Sprintf("%s/sample5.log", testDir)
	iniV.linesInBlock = 10
	iniV.maxBlocks = 20
	iniV.useDB = true
	iniV.minSupportPerBlock = 0.1
	iniV.absenceCheck = true
	iniV.rarityThreshold = 0.5
	iniV.absenceThreshold = 0.0

	verbose = false

	if _, err := copyFile("inputs/sample5.log",
		fmt.Sprintf("%s/sample5.log", testDir)); err != nil {
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
	if cnt != 200 {
		t.Errorf("countTargetLines does not match!")
		return
	}

	if err := a.run(0); err != nil {
		println(err)
		t.Errorf("%v", err)
		return
	}

	db := a.dciDB
	table := db.tables["closedItemSets"]
	cnt, err = table.count(nil, "*")
	if err != nil {
		t.Errorf("%v", err)
		return
	}
	if cnt != 200 {
		t.Errorf("count of closedItemSets is not correct")
		return
	}
	db = a.absAnal.db
	table = db.tables["items"]
	cnt, err = table.count(nil, "*")
	if err != nil {
		t.Errorf("%v", err)
		return
	}
	if cnt != 12 {
		t.Errorf("count of closedItemSets is not correct")
		return
	}
}

func TestLogAnalyzer_run4_nodb(t *testing.T) {
	//name := "TestLogAnalyzer_run4_nodb"
	logPathRegex := "inputs/sample5.log"
	iniV := new(logAnalyzerVars)
	iniV.name = "TestLogAnalyzer_run4_nodb"
	iniV.rootDir = ""
	iniV.logPathRegex = logPathRegex
	iniV.linesInBlock = 10
	iniV.maxBlocks = 20
	iniV.useDB = false
	iniV.minSupportPerBlock = 0.1
	iniV.absenceCheck = false
	iniV.rarityThreshold = 0.5
	iniV.absenceThreshold = 0.0

	verbose = false

	a, err := newLogAnalyzerByVars(iniV)
	if err != nil {
		t.Errorf("%v", err)
		return
	}
	defer a.close()

	cnt, err := a.rarAnal.countTargetLines()
	if err != nil {
		t.Errorf("%v", err)
		return
	}
	if cnt != 200 {
		t.Errorf("countTargetLines does not match!")
		return
	}

	if err := a.run(0); err != nil {
		println(err)
		t.Errorf("%v", err)
		return
	}

	blocks := a.rarAnal.blocks
	if len(blocks) != 20 {
		t.Errorf("Number of blocks does not match!")
		return
	}
	for i, block := range blocks {
		if block == nil {
			t.Errorf("A block is not expected to be nil!")
			return
		}
		if i != block.blockID {
			t.Errorf("A block must have blockID!")
			return
		}
		if block.blockCnt != 10 {
			t.Errorf("BlockCnt is not as expected!")
			return
		}
	}

}

func TestLogAnalyzer_run5_noabsence(t *testing.T) {
	iniV := new(logAnalyzerVars)
	iniV.name = "TestLogAnalyzer_run5_noabsence"
	testDir, err := ensureTestDir(iniV.name)
	if err != nil {
		t.Errorf("%v", err)
		return
	}
	iniV.rootDir = fmt.Sprintf("%s/db", testDir)
	iniV.logPathRegex = fmt.Sprintf("%s/sample4.log*", testDir)
	iniV.linesInBlock = 5
	iniV.maxBlocks = 3
	iniV.useDB = true
	iniV.minSupportPerBlock = 0.1
	iniV.absenceCheck = false
	iniV.rarityThreshold = 0.5
	iniV.absenceThreshold = 0.5

	verbose = false

	if _, err := copyFile("inputs/sample4.log",
		fmt.Sprintf("%s/sample4.log", testDir)); err != nil {
		t.Errorf("%v", err)
		return
	}

	a, err := newLogAnalyzerByVars(iniV)
	if err != nil {
		t.Errorf("%v", err)
		return
	}
	defer a.close()

	if err := a.destroy(); err != nil {
		t.Errorf("%v", err)
		return
	}

	if err := a.run(5); err != nil {
		println(err)
		t.Errorf("%v", err)
		return
	}

	if a.rarAnal.rowID != 5 {
		t.Errorf("currCount is wrong")
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

	a.close()

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

	if a.rarAnal.rowID != 5 {
		t.Errorf("currCount is wrong")
		return
	}

	if err := a.run(5); err != nil {
		println(err)
		t.Errorf("%v", err)
		return
	}
	if a.rarAnal.rowID != 10 {
		t.Errorf("currCount is wrong")
		return
	}
}
