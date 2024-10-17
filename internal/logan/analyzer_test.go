package logan

import (
	"fmt"
	"goLogAnalyzer/pkg/utils"
	"testing"
)

/*
 */
func Test_Analyzer_daily(t *testing.T) {
	testDir, err := utils.InitTestDir("Test_Analyzer_daily")
	if err != nil {
		t.Errorf("%v", err)
		return
	}

	// copy log to the test dir
	targetFile := fmt.Sprintf("%s/sample50_1.log", testDir)
	if _, err := utils.CopyFile("../../testdata/loganal/sample50_1.log",
		targetFile); err != nil {
		t.Errorf("%v", err)
		return
	}

	// new analyzer
	logPath := fmt.Sprintf("%s/sample50_*.log", testDir)
	logFormat := `^(?P<timestamp>\d+-\d+-\d+ \d+:\d+:\d+)] (?P<message>.+)$`
	layout := "2006-01-02 15:04:05"
	dataDir := testDir + "/data"
	maxBlocks := 100
	blockSize := 100
	keepPeriod := int64(100)
	countBorder := 10

	a, err := NewAnalyzer(dataDir, logPath, logFormat, layout, nil, nil,
		maxBlocks, blockSize, keepPeriod,
		utils.CFreqDay, 0, countBorder, nil, nil, nil, false)
	if err != nil {
		t.Errorf("%v", err)
		return
	}

	err = a.Feed(0)
	if err != nil {
		t.Errorf("%v", err)
		return
	}

	// lines processed
	if err := utils.GetGotExpErr("a.linesProcessed", a.linesProcessed, 50); err != nil {
		t.Errorf("%v", err)
		return
	}

	// block size estimation
	// each line represents one hour. daily rotate so block size must be 24
	if err := utils.GetGotExpErr("maxCountByBlock", a.trans.maxCountByBlock, 24); err != nil {
		t.Errorf("%v", err)
		return
	}

	// check config table
	if err := a._checkConfigTable(logPath,
		maxBlocks, blockSize,
		keepPeriod, utils.CFreqDay,
		0, countBorder,
		logFormat, layout); err != nil {
		t.Errorf("%v", err)
		return
	}

	// check status table
	if err := a._checkLastStatusTable(a.rowID, a.rowID, targetFile); err != nil {
		t.Errorf("%v", err)
		return
	}

}
