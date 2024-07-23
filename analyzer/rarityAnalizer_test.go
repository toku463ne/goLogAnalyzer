package analyzer

import (
	"fmt"
	"math"
	"strings"
	"testing"
	"time"
)

func getTestConfNodate(rootDir, logPath string) *AnalConf {
	c := NewAnalConf(rootDir)
	c.LogPathRegex = logPath
	c.BlockSize = 5
	c.DatetimeLayout = ""
	c.DatetimeStartPos = 0
	c.MaxBlocks = 3
	c.MaxItemBlocks = 6
	c.MinGapToRecord = 0
	c.ModeblockPerFile = true
	c.NTopRecordsCount = 3
	c.ScoreStyle = CDefaultScoreStyle
	return c
}

func Test_rarityAnalyzerInit(t *testing.T) {
	testDir, err := initTestDir("Test_rarityAnalyzerInit")
	if err != nil {
		t.Errorf("%v", err)
		return
	}

	logPathRegex := fmt.Sprintf("%s/sample1.log*", testDir)
	rootDir := testDir + "/data"

	c := getTestConfNodate(rootDir, logPathRegex)
	a, err := newRarityAnalyzer(c)
	if err != nil {
		t.Errorf("%+v", err)
		return
	}

	if a.LogPathRegex != c.LogPathRegex || a.MinGapToRecord != c.MinGapToRecord {
		t.Errorf("Not properly initialized")
		return
	}

	if a.stats.maxBlocks != c.MaxBlocks || a.stats.blockSize != c.BlockSize {
		t.Errorf("stats params Not properly set")
		return
	}
	if a.trans.items.maxBlocks != c.MaxItemBlocks || a.trans.blockSize != c.BlockSize {
		t.Errorf("stats params Not properly set")
		return
	}
	if a.logRecs.maxBlocks != c.MaxBlocks || a.logRecs.blockSize != c.BlockSize {
		t.Errorf("logRecs params Not properly set")
		return
	}

	a.close()

	a, err = newRarityAnalyzer(c)
	if err != nil {
		t.Errorf("%v", err)
		return
	}

	if a.LogPathRegex != c.LogPathRegex || a.MinGapToRecord != c.MinGapToRecord {
		t.Errorf("Not properly initialized")
		return
	}

	if a.stats.maxBlocks != c.MaxBlocks || a.stats.blockSize != c.BlockSize {
		t.Errorf("stats params Not properly set")
		return
	}
	if a.trans.items.maxBlocks != c.MaxItemBlocks || a.trans.blockSize != c.BlockSize {
		t.Errorf("stats params Not properly set")
		return
	}
	if a.logRecs.maxBlocks != c.MaxBlocks || a.logRecs.blockSize != c.BlockSize {
		t.Errorf("logRecs params Not properly set")
		return
	}

	// test calc block
	a.calcBlocks(4000, 2)
	if err := getGotExpErr("blockSize", a.BlockSize, 3000); err != nil {
		t.Errorf("%+v", err)
		return
	}
	if err := getGotExpErr("ModeblockPerFile", a.ModeblockPerFile, true); err != nil {
		t.Errorf("%+v", err)
		return
	}
	if err := getGotExpErr("MinGapToRecord", a.MinGapToRecord, 0.3); err != nil {
		t.Errorf("%+v", err)
		return
	}
	if err := getGotExpErr("MaxBlocks", a.MaxBlocks, 28); err != nil {
		t.Errorf("%+v", err)
		return
	}

	a.calcBlocks(10000, 2)
	if err := getGotExpErr("blockSize", a.BlockSize, 1000); err != nil {
		t.Errorf("%+v", err)
		return
	}
	if err := getGotExpErr("ModeblockPerFile", a.ModeblockPerFile, false); err != nil {
		t.Errorf("%+v", err)
		return
	}
	if err := getGotExpErr("MinGapToRecord", a.MinGapToRecord, 0.5); err != nil {
		t.Errorf("%+v", err)
		return
	}
	if err := getGotExpErr("MaxBlocks", a.MaxBlocks, 70); err != nil {
		t.Errorf("%+v", err)
		return
	}
	if err := getGotExpErr("MaxItemBlocks", a.MaxItemBlocks, 350); err != nil {
		t.Errorf("%+v", err)
		return
	}

	a.calcBlocks(100000, 2)
	if err := getGotExpErr("blockSize", a.BlockSize, 10000); err != nil {
		t.Errorf("%+v", err)
		return
	}
	if err := getGotExpErr("ModeblockPerFile", a.ModeblockPerFile, false); err != nil {
		t.Errorf("%+v", err)
		return
	}
	if err := getGotExpErr("MinGapToRecord", a.MinGapToRecord, 1.2); err != nil {
		t.Errorf("%+v", err)
		return
	}
	if err := getGotExpErr("MaxBlocks", a.MaxBlocks, 70); err != nil {
		t.Errorf("%+v", err)
		return
	}
	if err := getGotExpErr("MaxItemBlocks", a.MaxItemBlocks, 350); err != nil {
		t.Errorf("%+v", err)
		return
	}

	a.calcBlocks(1000000, 2)
	if err := getGotExpErr("blockSize", a.BlockSize, 100000); err != nil {
		t.Errorf("%+v", err)
		return
	}
	if err := getGotExpErr("ModeblockPerFile", a.ModeblockPerFile, false); err != nil {
		t.Errorf("%+v", err)
		return
	}
	if err := getGotExpErr("MinGapToRecord", a.MinGapToRecord, 1.5); err != nil {
		t.Errorf("%+v", err)
		return
	}
	if err := getGotExpErr("MaxBlocks", a.MaxBlocks, 70); err != nil {
		t.Errorf("%+v", err)
		return
	}
	if err := getGotExpErr("MaxItemBlocks", a.MaxItemBlocks, 350); err != nil {
		t.Errorf("%+v", err)
		return
	}

	a.close()
}

func Test_rarityAnalyzerRun(t *testing.T) {
	testDir, err := initTestDir("Test_rarityAnalyzerRun")
	if err != nil {
		t.Errorf("%v", err)
		return
	}

	logPathRegex := fmt.Sprintf("%s/sample.log*", testDir)
	rootDir := testDir + "/data"
	useGzipInCircuitTables = false

	c := getTestConfNodate(rootDir, logPathRegex)
	a, err := newRarityAnalyzer(c)
	if err != nil {
		t.Errorf("%v", err)
		return
	}

	if _, err := copyFile("testdata/rarityAnalizer/001/sample.log.1",
		fmt.Sprintf("%s/sample.log.1", testDir)); err != nil {
		t.Errorf("%v", err)
		return
	}

	if err := a.analyze(0); err != nil {
		t.Errorf("%v", err)
		return
	} else {
		if err := getGotExpErr("lines processed", a.linesProcessed, 31); err != nil {
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

	a, err = newRarityAnalyzer(c)
	if err != nil {
		t.Errorf("%v", err)
		return
	}

	if _, err := copyFile("testdata/rarityAnalizer/001/sample.log",
		fmt.Sprintf("%s/sample.log", testDir)); err != nil {
		t.Errorf("%v", err)
		return
	}

	if err := a.analyze(0); err != nil {
		t.Errorf("%v", err)
		return
	} else {
		if err := getGotExpErr("lines processed", a.linesProcessed, 4); err != nil {
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
}

func Test_rarityAnalyzerRun2(t *testing.T) {
	testDir, err := initTestDir("Test_rarityAnalyzerRun2")
	if err != nil {
		t.Errorf("%v", err)
		return
	}

	rootDir := testDir + "/data"
	logPathRegex := fmt.Sprintf("%s/sample.log*", testDir)

	c := getTestConfNodate(rootDir, logPathRegex)
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

	a, err := newRarityAnalyzer(c)
	if err != nil {
		t.Errorf("%v", err)
		return
	}

	if err := a.analyze(6); err != nil {
		t.Errorf("%v", err)
		return
	} else {
		if err := getGotExpErr("lines processed", a.linesProcessed, 6); err != nil {
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

	a, err = newRarityAnalyzer(c)
	if err != nil {
		t.Errorf("%v", err)
		return
	}

	if err := a.analyze(5); err != nil {
		t.Errorf("%v", err)
		return
	} else {
		if err := getGotExpErr("lines processed", a.linesProcessed, 5); err != nil {
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

	a, err = newRarityAnalyzer(c)
	if err != nil {
		t.Errorf("%v", err)
		return
	}

	if err := a.analyze(100); err != nil {
		t.Errorf("%v", err)
		return
	} else {
		if err := getGotExpErr("lines processed", a.linesProcessed, 20); err != nil {
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

func Test_scanAndGetNTop(t *testing.T) {
	testDir, err := initTestDir("Test_scanAndGetNTop")
	if err != nil {
		t.Errorf("%v", err)
		return
	}

	logPathRegex := fmt.Sprintf("%s/sample.log*", testDir)
	rootDir := testDir + "/data"
	useGzipInCircuitTables = false

	c := getTestConfNodate(rootDir, logPathRegex)
	a, err := newRarityAnalyzer(c)
	if err != nil {
		t.Errorf("%v", err)
		return
	}

	if _, err := copyFile("testdata/rarityAnalizer/001/sample.log.1",
		fmt.Sprintf("%s/sample.log.1", testDir)); err != nil {
		t.Errorf("%v", err)
		return
	}

	if err := a.analyze(0); err != nil {
		t.Errorf("%v", err)
		return
	} else {
		if err := getGotExpErr("lines processed", a.linesProcessed, 31); err != nil {
			t.Errorf("%v", err)
			return
		}
	}

	a.close()

	a, err = newRarityAnalyzer(c)
	if err != nil {
		t.Errorf("%v", err)
		return
	}
	n := 5
	nTop, err := a.scanAndGetNTop(n, 0, 0, "", "", 0, 0, 0)
	if err != nil {
		t.Errorf("%v", err)
		return
	}
	cnt := 1
	var oldrec *colLogRecord
	recs := nTop.getRecords()
	for i, rec := range recs {
		if i == 0 {
			oldrec = rec
			continue
		}
		if i >= n {
			t.Error("more records than expected")
			return
		}
		cnt++
		if math.IsNaN(rec.score) {
			t.Errorf("score of [%d] is NaN", i)
			return
		}
		if rec.score > oldrec.score {
			t.Errorf("score %f of [%d] must be smaller than [%f]", rec.score, i, oldrec.score)
			return
		}
	}
	if err := getGotExpErr("ntop number", cnt, n); err != nil {
		t.Errorf("%v", err)
		return
	}
}

func Test_rarityAnalyzerNodb(t *testing.T) {
	testDir, err := initTestDir("Test_rarityAnalyzerNodb")
	if err != nil {
		t.Errorf("%v", err)
		return
	}

	//rootDir := testDir + "/data"
	logPathRegex := fmt.Sprintf("%s/sample.log*", testDir)

	c := getTestConfNodate("", logPathRegex)
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

	a, err := newRarityAnalyzer(c)
	if err != nil {
		t.Errorf("%v", err)
		return
	}

	if err := a.analyze(0); err != nil {
		t.Errorf("%v", err)
		return
	} else {
		if err := getGotExpErr("lines processed", a.linesProcessed, 31); err != nil {
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
	var oldrec *colLogRecord
	for i, rec := range a.nTopRareLogs.records {
		if i == 0 {
			oldrec = rec
			continue
		}
		if math.IsNaN(rec.score) {
			t.Errorf("score of [%d] is NaN", i)
			return
		}
		if rec.score > oldrec.score {
			t.Errorf("score %f of [%d] must be smaller than [%f]", rec.score, i, oldrec.score)
			return
		}
	}
}
