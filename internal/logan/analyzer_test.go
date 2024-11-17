package logan

import (
	"errors"
	"fmt"
	"goLogAnalyzer/pkg/utils"
	"testing"
	"time"
)

/*
 */
func Test_Analyzer_daily_Feed(t *testing.T) {
	testDir, err := utils.InitTestDir("Test_Analyzer_daily_Feed")
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
	logFormat := `^(?P<timestamp>\d+-\d+-\d+T\d+:\d+:\d+)] (?P<message>.+)$`
	layout := "2006-01-02T15:04:05"
	dataDir := testDir + "/data"
	maxBlocks := 100
	blockSize := 100
	keepPeriod := int64(100)
	countBorder := 10
	minMatchRate := 0.5
	unitSecs := int64(3600 * 24)
	useUtcTime := true
	separator := " ,<>"

	a, err := NewAnalyzer(dataDir, logPath, logFormat, layout, useUtcTime, nil, nil,
		maxBlocks, blockSize, keepPeriod,
		unitSecs, 0, countBorder, minMatchRate, nil, nil, nil, separator, false, false, false)
	if err != nil {
		t.Errorf("%v", err)
		return
	}

	err = a.Feed(35)
	if err != nil {
		t.Errorf("%v", err)
		return
	}

	// lines processed
	if err := utils.GetGotExpErr("a.linesProcessed", a.linesProcessed, 35); err != nil {
		t.Errorf("%v", err)
		return
	}

	// block size estimation
	// each line represents one hour. daily rotate so block size is 24, but as specified blockSize=100, it must be 100
	if err := utils.GetGotExpErr("maxCountByBlock", a.trans.maxCountByBlock, 100); err != nil {
		t.Errorf("%v", err)
		return
	}

	// check config table
	if err := a._checkConfigTable(logPath,
		maxBlocks, blockSize,
		keepPeriod, unitSecs,
		0.999, countBorder, minMatchRate,
		layout, useUtcTime, separator, logFormat); err != nil {
		t.Errorf("%v", err)
		return
	}

	// check status table
	if err := a._checkLastStatusTable(a.rowID, a.rowID, targetFile); err != nil {
		t.Errorf("%v", err)
		return
	}

	// number of "com1"
	if err := utils.GetGotExpErr("com1 count", a.trans.te.getCount("com1"), 35); err != nil {
		t.Errorf("%v", err)
		return
	}

	// number of "grpa10"
	if err := utils.GetGotExpErr("grpa10 count", a.trans.te.getCount("grpa10"), 10); err != nil {
		t.Errorf("%v", err)
		return
	}
	// number of "grpd10"
	if err := utils.GetGotExpErr("grpd10 count", a.trans.te.getCount("grpd10"), 5); err != nil {
		t.Errorf("%v", err)
		return
	}

	// number of displayStrings
	if err := utils.GetGotExpErr("displayStrings", len(a.trans.lgs.displayStrings), 4); err != nil {
		t.Errorf("%v", err)
		return
	}
	if cnt, err := utils.CountGzFileLines(dataDir + "/logGroups/displaystrings.txt.gz"); err != nil {
		t.Errorf("%v", err)
		return
	} else {
		// number of displayStrings saved
		if err := utils.GetGotExpErr("displayStrings saved", cnt, 4); err != nil {
			t.Errorf("%v", err)
			return
		}
	}

	// number of lastMessages
	if err := utils.GetGotExpErr("lastMessages", len(a.trans.lgs.lastMessages), 4); err != nil {
		t.Errorf("%v", err)
		return
	}
	if cnt, err := utils.CountGzFileLines(dataDir + "/logGroups/lastMessages.txt.gz"); err != nil {
		t.Errorf("%v", err)
		return
	} else {
		// number of lastMessages saved
		if err := utils.GetGotExpErr("lastMessages saved", cnt, 4); err != nil {
			t.Errorf("%v", err)
			return
		}
	}

	lgstr := "Com1, grpa10 Com2 * grpa50 * <coM3> * grpa20 *"
	lg := a.trans.searchLogGroup(lgstr)
	// number of the log group
	if err := utils.GetGotExpErr(lgstr, lg.count, 10); err != nil {
		t.Errorf("%v", err)
		return
	}

	lgstr = "Com1, grpd10 Com2 * grpa50 * <coM3> * grpb20 *"
	lg = a.trans.searchLogGroup(lgstr)
	// number of the log group
	if err := utils.GetGotExpErr(lgstr, lg.count, 5); err != nil {
		t.Errorf("%v", err)
		return
	}
	// grpd10 is actually converted to "*" because it is less than countBorder(10)
	expected := "Com1, * Com2 * grpa50 * <coM3> * grpb20 *"
	if err := utils.GetGotExpErr("Display string in the last logGroup", lg.displayString, expected); err != nil {
		t.Errorf("%v", err)
		return
	}

	// get block 0 data
	bd0, err := a.trans.lgs.getBlockData(0)
	if err != nil {
		t.Errorf("%v", err)
		return
	}
	// get block 1 data
	bd1, err := a.trans.lgs.getBlockData(1)
	if err != nil {
		t.Errorf("%v", err)
		return
	}

	// get block 0 data
	tebd0, err := a.trans.te.getBlockData(0)
	if err != nil {
		t.Errorf("%v", err)
		return
	}
	// get block 1 data
	tebd1, err := a.trans.te.getBlockData(1)
	if err != nil {
		t.Errorf("%v", err)
		return
	}

	// block0 has 3 logGroups
	if err := utils.GetGotExpErr("block0 count", len(bd0), 3); err != nil {
		t.Errorf("%v", err)
		return
	}
	// block1 has 3 logGroups
	if err := utils.GetGotExpErr("block1 count", len(bd1), 2); err != nil {
		t.Errorf("%v", err)
		return
	}

	// block0 consists of 10+10+4=24
	// groupIdstr=d4kqiqizi0w3 has 4
	if err := utils.GetGotExpErr("group 1727812800000000003 count", bd0["1727812800000000003"].count, 4); err != nil {
		t.Errorf("%v", err)
		return
	}

	// block1 consists of 6+10=16
	// groupIdstr=d4kqiqizi0w3 has 6
	if err := utils.GetGotExpErr("group 1727812800000000003 count", bd1["1727812800000000003"].count, 6); err != nil {
		t.Errorf("%v", err)
		return
	}

	// total number must be 35
	sum := 0
	for _, bd := range []map[string]logGroup{bd0, bd1} {
		for _, lg := range bd {
			sum += lg.count
		}
	}
	if err := utils.GetGotExpErr("total count", sum, 35); err != nil {
		t.Errorf("%v", err)
		return
	}

	if err := utils.GetGotExpErr("term 'com1' in block0", tebd0["com1"], 24); err != nil {
		t.Errorf("%v", err)
		return
	}
	if err := utils.GetGotExpErr("term 'com1' in block1", tebd1["com1"], 11); err != nil {
		t.Errorf("%v", err)
		return
	}

	// close
	a.Close()

	a, err = LoadAnalyzer(dataDir, "", 0, 0, 0, nil, false, false, false)
	if err != nil {
		t.Errorf("%v", err)
		return
	}

	// check the last status is loaded
	if err := utils.GetGotExpErr("rowId", a.rowID, int(35)); err != nil {
		t.Errorf("%v", err)
		return
	}

	// check logPath
	if err := utils.GetGotExpErr("logPath after load", a.logPath, logPath); err != nil {
		t.Errorf("%v", err)
		return
	}

	if err := utils.GetGotExpErr("logFormat after load", a.logFormat, logFormat); err != nil {
		t.Errorf("%v", err)
		return
	}
	if err := utils.GetGotExpErr("keepPeriod after load", a.keepPeriod, keepPeriod); err != nil {
		t.Errorf("%v", err)
		return
	}
	if err := utils.GetGotExpErr("maxBlocks after load", a.maxBlocks, maxBlocks); err != nil {
		t.Errorf("%v", err)
		return
	}
	if err := utils.GetGotExpErr("termCountBorder after load", a.termCountBorder, countBorder); err != nil {
		t.Errorf("%v", err)
		return
	}

	sum = 0
	for _, lg := range a.trans.lgs.alllg {
		sum += lg.count
	}
	if err := utils.GetGotExpErr("number of total logGroups", sum, 35); err != nil {
		t.Errorf("%v", err)
		return
	}
	sum = 0
	for _, lg := range a.trans.lgs.curlg {
		sum += lg.count
	}
	if err := utils.GetGotExpErr("number of current logGroups", sum, 11); err != nil {
		t.Errorf("%v", err)
		return
	}

	// feed the rest lines
	err = a.Feed(0)
	if err != nil {
		t.Errorf("%v", err)
		return
	}

	// check the rowID is properly updated
	if err := utils.GetGotExpErr("rowId", a.rowID, int(50)); err != nil {
		t.Errorf("%v", err)
		return
	}

	// number of "com1"
	if err := utils.GetGotExpErr("com1 count", a.trans.te.getCount("com1"), 50); err != nil {
		t.Errorf("%v", err)
		return
	}

	// number of "grpa10"
	if err := utils.GetGotExpErr("grpa10 count", a.trans.te.getCount("grpa10"), 10); err != nil {
		t.Errorf("%v", err)
		return
	}

	// number of "grpd10"
	// it was 5 before the 2nd Feed. Not it is 10
	if err := utils.GetGotExpErr("grpd10 count", a.trans.te.getCount("grpd10"), 10); err != nil {
		t.Errorf("%v", err)
		return
	}

	// check logGroup parsed in the first feed
	lgstr = "Com1, grpa10 Com2 * grpa50 * <coM3> * grpa20 *"
	lg = a.trans.searchLogGroup(lgstr)
	// number of the log group
	if err := utils.GetGotExpErr(lgstr, lg.count, 10); err != nil {
		t.Errorf("%v", err)
		return
	}

	// in the 2nd Feed grpd10 counted up from 5 to 10
	lgstr = "Com1, grpd10 Com2 * grpa50 * <coM3> * grpb20 *"
	lg = a.trans.searchLogGroup(lgstr)
	// number of the log group
	if err := utils.GetGotExpErr(lgstr, lg.count, 5); err != nil {
		t.Errorf("%v", err)
		return
	}

	// in the 2nd Feed grpd10 counted up from 5 to 10
	// we still have "Com1, * Com2 * grpa50 * <coM3> * grpb20 *" generated in the 1st Feed
	lgstr = "Com1, testuniq001 Com2 * grpa50 * <coM3> * grpb20 *"
	lg = a.trans.searchLogGroup(lgstr)
	// number of the log group
	if err := utils.GetGotExpErr(lgstr, lg.count, 5); err != nil {
		t.Errorf("%v", err)
		return
	}
	if err := utils.GetGotExpErr("displayString", lg.displayString, "Com1, * Com2 * grpa50 * <coM3> * grpb20 *"); err != nil {
		t.Errorf("%v", err)
		return
	}

	// new group
	lgstr = "Com1, grpe10 Com2 * grpa50 * <coM3> * grpc20 *"
	lg = a.trans.searchLogGroup(lgstr)
	// number of the log group
	if err := utils.GetGotExpErr(lgstr, lg.count, 10); err != nil {
		t.Errorf("%v", err)
		return
	}

	// get block 1 data
	bd1, err = a.trans.lgs.getBlockData(1)
	if err != nil {
		t.Errorf("%v", err)
		return
	}
	// get block 2 data
	bd2, err := a.trans.lgs.getBlockData(2)
	if err != nil {
		t.Errorf("%v", err)
		return
	}
	// get block 1 data
	tebd1, err = a.trans.te.getBlockData(1)
	if err != nil {
		t.Errorf("%v", err)
		return
	}
	// get block 2 data
	tebd2, err := a.trans.te.getBlockData(2)
	if err != nil {
		t.Errorf("%v", err)
		return
	}

	// block1 has 3 logGroups
	if err := utils.GetGotExpErr("block1 count", len(bd1), 4); err != nil {
		t.Errorf("%v", err)
		return
	}

	// block2 has 1 logGroups
	if err := utils.GetGotExpErr("block2 count", len(bd2), 1); err != nil {
		t.Errorf("%v", err)
		return
	}

	if err := utils.GetGotExpErr("term 'com1' in block1", tebd1["com1"], 24); err != nil {
		t.Errorf("%v", err)
		return
	}

	if err := utils.GetGotExpErr("term 'com1' in block2", tebd2["com1"], 2); err != nil {
		t.Errorf("%v", err)
		return
	}

	a.Close()

	// number of displayStrings saved
	if cnt, err := utils.CountGzFileLines(dataDir + "/logGroups/displaystrings.txt.gz"); err != nil {
		t.Errorf("%v", err)
		return
	} else {
		if err := utils.GetGotExpErr("displayStrings saved", cnt, 6); err != nil {
			t.Errorf("%v", err)
			return
		}
	}

	// number of lastMessages
	if cnt, err := utils.CountGzFileLines(dataDir + "/logGroups/lastMessages.txt.gz"); err != nil {
		t.Errorf("%v", err)
		return
	} else {
		// number of lastMessages saved
		if err := utils.GetGotExpErr("lastMessages saved", cnt, 6); err != nil {
			t.Errorf("%v", err)
			return
		}
	}

	// wait 2 sec
	time.Sleep(2000000000)

	// copy the 2nd log to the test dir
	targetFile = fmt.Sprintf("%s/sample50_2.log", testDir)
	if _, err := utils.CopyFile("../../testdata/loganal/sample50_2.log",
		targetFile); err != nil {
		t.Errorf("%v", err)
		return
	}

	a, err = LoadAnalyzer(dataDir, "", 0, 0, 0, nil, false, false, false)
	if err != nil {
		t.Errorf("%v", err)
		return
	}
	// feed the rest lines
	err = a.Feed(0)
	if err != nil {
		t.Errorf("%v", err)
		return
	}

	// number of "com1"
	if err := utils.GetGotExpErr("com1 count", a.trans.te.getCount("com1"), 100); err != nil {
		t.Errorf("%v", err)
		return
	}

	if err := utils.GetGotExpErr("grpe20 count", a.trans.te.getCount("grpe20"), 20); err != nil {
		t.Errorf("%v", err)
		return
	}

	if err := utils.GetGotExpErr("displayStrings", len(a.trans.lgs.displayStrings), 11); err != nil {
		t.Errorf("%v", err)
		return
	}

	a.Close()

	a, err = LoadAnalyzer(dataDir, "", 0, 0, 0, nil, false, false, false)
	if err != nil {
		t.Errorf("%v", err)
		return
	}

	if err := a.OutputLogGroups(10, testDir, true, false); err != nil {
		t.Errorf("%v", err)
		return
	}
	header, records, err := utils.ReadCsv(testDir+"/history.csv", ',', false)
	if err != nil {
		t.Errorf("%v", err)
		return
	}
	if err := utils.GetGotExpErr("len(header)", len(header), 6); err != nil {
		t.Errorf("%v", err)
		return
	}
	if err := utils.GetGotExpErr("len(records)", len(records), 10); err != nil {
		t.Errorf("%v", err)
		return
	}

	if err := utils.GetGotExpErr("records[0][1]", records[0][1], "10"); err != nil {
		t.Errorf("%v", err)
		return
	}
	if err := utils.GetGotExpErr("records[2][1]", records[2][1], "4"); err != nil {
		t.Errorf("%v", err)
		return
	}
	if err := utils.GetGotExpErr("records[2][2]", records[2][2], "6"); err != nil {
		t.Errorf("%v", err)
		return
	}

	header, records, err = utils.ReadCsv(testDir+"/logGroups.csv", ',', false)
	if err != nil {
		t.Errorf("%v", err)
		return
	}

	if err := utils.GetGotExpErr("len(header)", len(header), 4); err != nil {
		t.Errorf("%v", err)
		return
	}
	if err := utils.GetGotExpErr("len(records)", len(records), 10); err != nil {
		t.Errorf("%v", err)
		return
	}
	if err := utils.GetGotExpErr("records[0][1]", records[0][0], "1727740800000000001"); err != nil {
		t.Errorf("%v", err)
		return
	}

	a.Close()

	// rebuild trans
	a, err = LoadAnalyzer(dataDir, "", 0, 20, 0.5, nil, false, false, false)
	if err != nil {
		t.Errorf("%v", err)
		return
	}

	if err := a.OutputLogGroups(10, testDir, true, false); err != nil {
		t.Errorf("%v", err)
		return
	}
	_, records, err = utils.ReadCsv(testDir+"/history.csv", ',', false)
	if err != nil {
		t.Errorf("%v", err)
		return
	}
	if err := utils.GetGotExpErr("len(records)", len(records), 6); err != nil {
		t.Errorf("%v", err)
		return
	}

	if err := utils.GetGotExpErr("records[0][1]", records[0][1], "20"); err != nil {
		t.Errorf("%v", err)
		return
	}

}

func _test_Trans_parse(line, logFormat, layout string,
	useUtcTime bool, unitSecs int64,
	expect_line string) error {
	a, err := NewAnalyzer("", "", logFormat, layout, useUtcTime, nil, nil,
		0, 0, 0,
		unitSecs, 0, 0, 0, nil, nil, nil, "", false, false, false)
	if err != nil {
		return err
	}

	line, updated, _ := a.trans.parseLine(line, 0)

	if err := utils.GetGotExpErr("line", line, expect_line); err != nil {
		return err
	}
	if updated == 0 {
		return errors.New("timestamp was not parsed correctly")
	}

	return nil
}

func Test_Trans_parse(t *testing.T) {
	line := "10th, 02:14:49.143+0900 TBLV1 DAO : CTBCAFLogDao::Sync reseting GzipPendingEvent"
	expect_line := "TBLV1 DAO : CTBCAFLogDao::Sync reseting GzipPendingEvent"
	logFormat := `^(?P<timestamp>\d+\w+, \d{2}:\d{2}:\d{2}\.\d+\+\d{4}) (?P<message>.+)$`
	layout := "02th, 15:04:05.000-0700"
	useUtcTime := false
	unitSecs := int64(3600)
	_test_Trans_parse(line, logFormat, layout, useUtcTime, unitSecs,
		expect_line)

	logFormat = `^(?P<timestamp>\d+\-\d+\-\d+ \d+:\d+:\d+)\|(?P<timestamp2>\d+\-\d+\-\d+ \d+:\d+:\d+)\|(?P<message>.*)$`
	layout = "2006-01-02 15:04:05"
	line = "2024-08-31 23:21:43|2024-08-31 23:21:49|INVITE sip:0678786395@PRO-FE.ziptelecom.tel;user=phone;"
	expect_line = "INVITE sip:0678786395@PRO-FE.ziptelecom.tel;user=phone;"
	_test_Trans_parse(line, logFormat, layout, useUtcTime, unitSecs,
		expect_line)

}

func Test_Analyzer_multisize(t *testing.T) {
	testDir, err := utils.InitTestDir("Test_Analyzer_multisize")
	if err != nil {
		t.Errorf("%v", err)
		return
	}

	// new analyzer
	logPath := "../../testdata/loganal/sample_multisize.log"
	logFormat := `^(?P<timestamp>\d+-\d+-\d+T\d+:\d+:\d+)] (?P<message>.+)$`
	layout := "2006-01-02T15:04:05"
	dataDir := testDir + "/data"
	maxBlocks := 100
	blockSize := 100
	keepPeriod := int64(100)
	countBorder := 2
	minMatchRate := 0.6
	unitSecs := int64(3600 * 24)
	useUtcTime := true
	separator := " ,<>"

	a, err := NewAnalyzer(dataDir, logPath, logFormat, layout, useUtcTime, nil, nil,
		maxBlocks, blockSize, keepPeriod,
		unitSecs, 0, countBorder, minMatchRate, nil, nil, nil, separator, false, false, false)
	if err != nil {
		t.Errorf("%v", err)
		return
	}

	err = a.OutputLogGroups(10, dataDir, false, true)
	if err != nil {
		t.Errorf("%v", err)
		return
	}

	// lines processed
	if err := utils.GetGotExpErr("a.linesProcessed", a.linesProcessed, 34); err != nil {
		t.Errorf("%v", err)
		return
	}

	header, records, err := utils.ReadCsv(dataDir+"/logGroups.csv", ',', false)
	if err != nil {
		t.Errorf("%v", err)
		return
	}

	if err := utils.GetGotExpErr("len(header)", len(header), 4); err != nil {
		t.Errorf("%v", err)
		return
	}
	if err := utils.GetGotExpErr("len(header)", len(records), 5); err != nil {
		t.Errorf("%v", err)
		return
	}
	if err := utils.GetGotExpErr("records[0][1]: count", records[0][1], "3"); err != nil {
		t.Errorf("%v", err)
		return
	}
	if err := utils.GetGotExpErr("records[4][1]: count", records[4][1], "10"); err != nil {
		t.Errorf("%v", err)
		return
	}

	err = a.OutputLogGroups(10, dataDir, false, false)
	if err != nil {
		t.Errorf("%v", err)
		return
	}
	_, records, err = utils.ReadCsv(dataDir+"/logGroups.csv", ',', false)
	if err != nil {
		t.Errorf("%v", err)
		return
	}
	if err := utils.GetGotExpErr("records[0][1]: count", records[0][1], "10"); err != nil {
		t.Errorf("%v", err)
		return
	}
}
