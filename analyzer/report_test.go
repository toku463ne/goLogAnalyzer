package analyzer

import "testing"

func TestReport_json(t *testing.T) {

	ls, err := newLogInfoMap("testdata/abnormal/config_test.json")
	if err != nil {
		t.Errorf("%v", err)
		return
	}

	if err := getGotExpErr("a.search", ls.Logs["a"].Search, "test333"); err != nil {
		t.Errorf("%v", err)
		return
	}

	if err := getGotExpErr("a.search", ls.TopN, 3); err != nil {
		t.Errorf("%v", err)
		return
	}

	if err := getGotExpErr("b.exclude", ls.Logs["b"].Exclude, "test333"); err != nil {
		t.Errorf("%v", err)
		return
	}

	if err := getGotExpErr("c.linesInBlock", ls.Logs["c"].LinesInBlock, 2); err != nil {
		t.Errorf("%v", err)
		return
	}

	if err := getGotExpErr("c.maxBlocks", ls.Logs["c"].MaxBlocks, 2); err != nil {
		t.Errorf("%v", err)
		return
	}

}

func TestReport_run(t *testing.T) {
	jsonFile := "testdata/abnormal/config_test.json"
	err := removeTestDir("TestReport_run")
	if err != nil {
		t.Errorf("%v", err)
		return
	}

	dataDir, err := ensureTestDir("TestReport_run")
	if err != nil {
		t.Errorf("%v", err)
		return
	}
	rs := newReports()
	rs.dataDir = dataDir
	err = rs.run(jsonFile, 0, "",
		0,
		3, 3,
		4, 3, 3,
		0, "Jan _2 15:04:05")
	if err != nil {
		t.Errorf("%v", err)
		return
	}
	if len(rs.rep) != 3 {
		t.Errorf("Unexpected item count")
		return
	}

	if len(rs.rep["a"].nTopNorm.getRecords()) != 1 {
		t.Errorf("Unexpected nTopNorm record count")
		return
	}
	if len(rs.rep["a"].nTopErr.getRecords()) != 1 {
		t.Errorf("Unexpected nTopErr record count")
		return
	}
	if len(rs.rep["b"].nTopNorm.getRecords()) != 2 {
		t.Errorf("Unexpected nTopNorm record count")
		return
	}
	if len(rs.rep["b"].nTopErr.getRecords()) != 1 {
		t.Errorf("Unexpected nTopErr record count")
		return
	}
	if len(rs.rep["c"].nTopNorm.getRecords()) != 3 {
		t.Errorf("Unexpected nTopNorm record count")
		return
	}
	if len(rs.rep["c"].nTopErr.getRecords()) != 1 {
		t.Errorf("Unexpected nTopErr record count")
		return
	}

}
