package analyzer

import "testing"

func TestReport_json(t *testing.T) {

	ls, err := newLogSetInfo("testdata/abnormal/config_test.json")
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
	ls, err := newLogSetInfo("testdata/abnormal/config_test.json")
	if err != nil {
		t.Errorf("%v", err)
		return
	}
	ls.DataDir, err = ensureTestDir("TestReport_run")
	if err != nil {
		t.Errorf("%v", err)
		return
	}

	err = ls.run(1000, 0.0, 3, 3, 5, 5, 5)
	if err != nil {
		t.Errorf("%v", err)
		return
	}
}

func TestReport_run2(t *testing.T) {
	ls, err := newLogSetInfo("c:/Users/kot/loganal/zimconfwin.json")
	if err != nil {
		t.Errorf("%v", err)
		return
	}

	//recentNdays int,
	//defaultMinGapToRecord float64,
	//defaultMaxBlocks, defaultMaxItemBlocks,
	//defaultLinesInBlock, defaultNTopRecords, defaultHistSize int
	err = ls.run(0, 0.0, 3, 3, 5, 5, 5)
	if err != nil {
		t.Errorf("%v", err)
		return
	}
}
