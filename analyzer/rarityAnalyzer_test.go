package analyzer

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

func Test_rarityAnalyzerInit(t *testing.T) {
	testDir, err := ensureTestDir("rarityAnalyzerInit")
	if err != nil {
		t.Errorf("%v", err)
		return
	}

	rootDir := testDir + "/data"
	logPathRegex := fmt.Sprintf("%s/sample1.log*", testDir)
	filterStr := ""
	xFilterStr := ""
	minGapToRecord := 0.1
	maxBlocks := 3
	maxItemBlocks := 6
	linesInBlock := 5

	a := newRarityAnalyzer(rootDir)

	if err := a.clean(); err != nil {
		t.Errorf("%v", err)
		return
	}

	if err := a.init(logPathRegex,
		filterStr, xFilterStr,
		minGapToRecord, maxBlocks, maxItemBlocks, linesInBlock, 0); err != nil {
		t.Errorf("%v", err)
		return
	}

	if a.logPathRegex != logPathRegex || a.minGapToRecord != minGapToRecord || a.maxBlocks != maxBlocks || a.maxItemBlocks != maxItemBlocks || a.linesInBlock != linesInBlock {
		t.Errorf("Not properly initialized")
		return
	}

	if a.stats.maxBlocks != maxBlocks || a.stats.maxRowsInBlock != linesInBlock {
		t.Errorf("stats params Not properly set")
		return
	}
	if a.trans.items.maxBlocks != maxItemBlocks || a.trans.maxRowsInBlock != linesInBlock {
		t.Errorf("stats params Not properly set")
		return
	}
	if a.logRecs.maxBlocks != maxBlocks || a.logRecs.maxRowsInBlock != linesInBlock {
		t.Errorf("logRecs params Not properly set")
		return
	}

	a.close()

	a = newRarityAnalyzer(rootDir)
	if err := a.load(); err != nil {
		t.Errorf("%v", err)
		return
	}

	if a.logPathRegex != logPathRegex || a.minGapToRecord != minGapToRecord || a.maxBlocks != maxBlocks || a.maxItemBlocks != maxItemBlocks || a.linesInBlock != linesInBlock {
		t.Errorf("Not properly loaded")
		return
	}

	a.close()

	a = newRarityAnalyzer(rootDir)
	a.clean()

	if err := a.init(logPathRegex,
		"", "",
		-1.0, -1, -1, -1, 0); err != nil {
		t.Errorf("%v", err)
		return
	}

	a.close()

	a = newRarityAnalyzer(rootDir)
	if err := a.load(); err != nil {
		t.Errorf("%v", err)
		return
	}

	if a.lastFileEpoch != 0 || a.lastFileRow != 0 {
		t.Errorf("Not properly loaded lastStatus")
		return
	}
}

func Test_rarityAnalyzerRun(t *testing.T) {
	testDir, err := ensureTestDir("rarityAnalyzerRun")
	if err != nil {
		t.Errorf("%v", err)
		return
	}

	rootDir := testDir + "/data"
	logPathRegex := fmt.Sprintf("%s/sample.log*", testDir)
	filterStr := ""
	xFilterStr := ""
	minGapToRecord := -100.0
	maxBlocks := 3
	maxItemBlocks := 6
	linesInBlock := 5

	if err := removePath(fmt.Sprintf("%s/sample.log*", testDir)); err != nil {
		t.Errorf("%v", err)
		return
	}

	if _, err := copyFile("testdata/rarityAnalizer/001/sample.log.1",
		fmt.Sprintf("%s/sample.log.1", testDir)); err != nil {
		t.Errorf("%v", err)
		return
	}

	a := newRarityAnalyzer(rootDir)

	if err := a.clean(); err != nil {
		t.Errorf("%v", err)
		return
	}

	if err := a.init(logPathRegex,
		filterStr, xFilterStr,
		minGapToRecord, maxBlocks, maxItemBlocks, linesInBlock, 0); err != nil {
		t.Errorf("%v", err)
		return
	}

	if lines, err := a.analyze(0); err != nil {
		t.Errorf("%v", err)
		return
	} else {
		if err := getGotExpErr("lines processed", lines, 31); err != nil {
			t.Errorf("%v", err)
			return
		}
	}

	if err := getGotExpErr("rowNo", a.rowID, int64(31)); err != nil {
		t.Errorf("%v", err)
		return
	}

	if err := getGotExpErr("logRecords count", a.logRecs.countAll(nil), 11); err != nil {
		t.Errorf("%v", err)
		return
	}
	if err := getGotExpErr("items count", a.trans.items.countAll(nil), 38); err != nil {
		t.Errorf("%v", err)
		return
	}

	itemIdx := getColIdx("items", "item")
	it := "test100"
	fu := func(v []string) bool {
		return strings.Contains(v[itemIdx], it)
	}

	if err := getGotExpErr("items count test100",
		a.trans.items.countAll(fu), 2); err != nil {
		t.Errorf("%v", err)
		return
	}
	it = "test102"
	if err := getGotExpErr("items count test102",
		a.trans.items.countAll(fu), 1); err != nil {
		t.Errorf("%v", err)
		return
	}
	it = "test3"
	if err := getGotExpErr("items count test3*",
		a.trans.items.countAll(fu), 26); err != nil {
		t.Errorf("%v", err)
		return
	}

	lastFileEpoch := 0
	lastFileRow := 0

	if err := a.lastStatusTable.Select1Row(nil,
		[]string{"lastFileEpoch", "lastFileRow"},
		&lastFileEpoch, &lastFileRow); err != nil {
		t.Errorf("%v", err)
		return
	} else {
		if lastFileEpoch == 0 || lastFileRow == 0 {
			t.Errorf("lastStatus is not properly configured")
			return
		}
	}

	a.close()

	time.Sleep(time.Second * 2)

	a = newRarityAnalyzer(rootDir)
	if err := a.load(); err != nil {
		t.Errorf("%v", err)
		return
	}

	if _, err := copyFile("testdata/rarityAnalizer/001/sample.log",
		fmt.Sprintf("%s/sample.log", testDir)); err != nil {
		t.Errorf("%v", err)
		return
	}

	if lines, err := a.analyze(0); err != nil {
		t.Errorf("%v", err)
		return
	} else {
		if err := getGotExpErr("lines processed", lines, 4); err != nil {
			t.Errorf("%v", err)
			return
		}
	}

	if err := getGotExpErr("rowNo", a.rowID, int64(35)); err != nil {
		t.Errorf("%v", err)
		return
	}

	it = "test3"
	if err := getGotExpErr("items count test3*",
		a.trans.items.countAll(fu), 30); err != nil {
		t.Errorf("%v", err)
		return
	}

	topN, err := a.scanAndGetNTops(5, 0, 0, "", "")
	if err != nil {
		t.Errorf("%v", err)
		return
	}
	if err := getGotExpErr("topN len", len(topN), 5); err != nil {
		t.Errorf("%v", err)
		return
	}
	if err := getGotExpErr("topN top1", topN[0].rowid, int64(30)); err != nil {
		t.Errorf("%v", err)
		return
	}

}

func Test_rarityAnalyzerRun2(t *testing.T) {
	testDir, err := ensureTestDir("rarityAnalyzerRun2")
	if err != nil {
		t.Errorf("%v", err)
		return
	}

	rootDir := testDir + "/data"
	logPathRegex := fmt.Sprintf("%s/sample.log*", testDir)
	filterStr := ""
	xFilterStr := ""
	minGapToRecord := 0.0
	maxBlocks := 3
	maxItemBlocks := 6
	linesInBlock := 5
	useGzipInCircuitTables = false

	if err := removePath(fmt.Sprintf("%s/sample.log*", testDir)); err != nil {
		t.Errorf("%v", err)
		return
	}

	if _, err := copyFile("testdata/rarityAnalizer/001/sample.log.1",
		fmt.Sprintf("%s/sample.log.1", testDir)); err != nil {
		t.Errorf("%v", err)
		return
	}

	a := newRarityAnalyzer(rootDir)

	if err := a.clean(); err != nil {
		t.Errorf("%v", err)
		return
	}

	if err := a.init(logPathRegex,
		filterStr, xFilterStr,
		minGapToRecord, maxBlocks, maxItemBlocks, linesInBlock, 0); err != nil {
		t.Errorf("%v", err)
		return
	}

	if lines, err := a.analyze(6); err != nil {
		t.Errorf("%v", err)
		return
	} else {
		if err := getGotExpErr("lines processed", lines, 6); err != nil {
			t.Errorf("%v", err)
			return
		}
	}

	if err := getGotExpErr("rowNo", a.rowID, int64(6)); err != nil {
		t.Errorf("%v", err)
		return
	}

	itemIdx := getColIdx("items", "item")
	it := "test3"
	fu := func(v []string) bool {
		return strings.Contains(v[itemIdx], it)
	}
	if err := getGotExpErr("items count test3*",
		a.trans.items.countAll(fu), 6); err != nil {
		t.Errorf("%v", err)
		return
	}

	a.close()

	a = newRarityAnalyzer(rootDir)
	if err := a.load(); err != nil {
		t.Errorf("%v", err)
		return
	}

	if lines, err := a.analyze(5); err != nil {
		t.Errorf("%v", err)
		return
	} else {
		if err := getGotExpErr("lines processed", lines, 5); err != nil {
			t.Errorf("%v", err)
			return
		}
	}

	if err := getGotExpErr("rowNo", a.rowID, int64(11)); err != nil {
		t.Errorf("%v", err)
		return
	}

	it = "test3"
	if err := getGotExpErr("items count test3*",
		a.trans.items.countAll(fu), 11); err != nil {
		t.Errorf("%v", err)
		return
	}
	a.close()

	a = newRarityAnalyzer(rootDir)
	if err := a.load(); err != nil {
		t.Errorf("%v", err)
		return
	}

	if lines, err := a.analyze(100); err != nil {
		t.Errorf("%v", err)
		return
	} else {
		if err := getGotExpErr("lines processed", lines, 20); err != nil {
			t.Errorf("%v", err)
			return
		}
	}

	if err := getGotExpErr("rowNo", a.rowID, int64(31)); err != nil {
		t.Errorf("%v", err)
		return
	}

	it = "test3"
	if err := getGotExpErr("items count test3*",
		a.trans.items.countAll(fu), 26); err != nil {
		t.Errorf("%v", err)
		return
	}
	a.close()

}

func Test_rarityAnalyzerNodb(t *testing.T) {
	testDir, err := ensureTestDir("rarityAnalyzerRun2")
	if err != nil {
		t.Errorf("%v", err)
		return
	}

	logPathRegex := fmt.Sprintf("%s/sample.log*", testDir)
	filterStr := ""
	xFilterStr := ""
	minGapToRecord := 0.0
	maxBlocks := 3
	maxItemBlocks := 6
	linesInBlock := 5
	useGzipInCircuitTables = false

	if err := removePath(fmt.Sprintf("%s/sample.log*", testDir)); err != nil {
		t.Errorf("%v", err)
		return
	}

	if _, err := copyFile("testdata/rarityAnalizer/001/sample.log.1",
		fmt.Sprintf("%s/sample.log.1", testDir)); err != nil {
		t.Errorf("%v", err)
		return
	}

	a := newRarityAnalyzer("")
	if err := a.init(logPathRegex,
		filterStr, xFilterStr,
		minGapToRecord, maxBlocks, maxItemBlocks, linesInBlock, 5); err != nil {
		t.Errorf("%v", err)
		return
	}

	if lines, err := a.analyze(0); err != nil {
		t.Errorf("%v", err)
		return
	} else {
		if err := getGotExpErr("lines processed", lines, 31); err != nil {
			t.Errorf("%v", err)
			return
		}
	}

	if err := getGotExpErr("rowID", a.rowID, int64(31)); err != nil {
		t.Errorf("%v", err)
		return
	}
	if err := getGotExpErr("items.totalCount", a.trans.items.totalCount, 93); err != nil {
		t.Errorf("%v", err)
		return
	}

}
