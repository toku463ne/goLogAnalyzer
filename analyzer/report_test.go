package analyzer

import (
	"fmt"
	"os"
	"strings"
	"testing"
	"time"
)

func TestReport_json(t *testing.T) {

	ls, err := newLogInfoMap("testdata/report/config_test.json")
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

/*
func TestReport_run1(t *testing.T) {
	jsonFile := "testdata/report/config_test.json"
	err := removeTestDir("TestReport_run1")
	if err != nil {
		t.Errorf("%v", err)
		return
	}

	dataDir, err := ensureTestDir("TestReport_run1")
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
*/

func TestReport_run2(t *testing.T) {
	err := removeTestDir("TestReport_run2")
	if err != nil {
		t.Errorf("%v", err)
		return
	}

	dataDir, err := ensureTestDir("TestReport_run2")
	if err != nil {
		t.Errorf("%v", err)
		return
	}
	dataDir = strings.Replace(dataDir, "\\", "/", -1)

	jstring := fmt.Sprintf(`{
	"dataDir": "%s",
	"topN": 4,
	"dateLayout": "Jan _2 15:04:05",
	"minGapToRecord": 0,
	"scoreStyle": 1,
	"logs": {
		"test": {
			"path": "%s/test.log*",
			"maxBlocks": 3,
			"linesInBlock": 5
		}
	}
}`, dataDir, dataDir)
	jsonFile := fmt.Sprintf("%s/config_test.json", dataDir)
	if err := os.WriteFile(jsonFile, []byte(jstring), 0644); err != nil {
		t.Errorf("%v", err)
		return
	}

	data := `Oct 11 01:01:14 test101 test201 test301
Oct 11 01:02:14 test101 test201 test302
Oct 11 01:03:14 test101 test201 test303
Oct 11 01:04:14 test101 test201 test304
Oct 11 01:05:14 test101 error201 test305
Oct 11 01:06:14 test101 test202 test306
Oct 11 01:07:14 test101 test202 test307
Oct 11 01:08:14 test101 test202 test308
Oct 11 01:09:14 test101 test202 test309
Oct 11 01:10:14 test101 error202 test310
Oct 11 01:11:14 test101 test203 test311
Oct 11 01:12:14 test101 test203 test312
Oct 11 01:13:14 test101 test203 test313
Oct 11 01:14:14 test101 test203 test314
Oct 11 01:15:14 test101 error203 test315
Oct 11 01:16:14 test101 test204 test316
`

	dataFile := fmt.Sprintf("%s/test.log", dataDir)
	if err := os.WriteFile(dataFile, []byte(data), 0644); err != nil {
		t.Errorf("%v", err)
		return
	}

	rs := newReports()
	err = rs.run(jsonFile, 0, "",
		0,
		0, 0,
		0, 0, 3,
		0, "")
	if err != nil {
		t.Errorf("%v", err)
		return
	}

	if rs.rep == nil || len(rs.rep) == 0 {
		t.Errorf("No report")
		return
	}

	r := rs.rep["test"]

	/*
		for _, rr1 := range r.nTopErr.records {
			if rr1 == nil {
				break
			}
			log.Printf("rowid=%d score=%f count=%d text=%s", rr1.rowid, rr1.score, rr1.count, rr1.record)
		}*/

	recs := r.nTopNorm.getRecords()
	if len(recs) != 4 {
		t.Errorf("len is incorrect")
		return
	}
	if recs[0].rowid != 15 {
		t.Errorf("rowid is incorrect")
		return
	}
	recs = r.nTopErr.getRecords()
	if len(recs) != 2 {
		t.Errorf("len is incorrect")
		return
	}

	rs = nil

	time.Sleep(2.0)

	data = `Oct 12 01:17:14 test101 test201 test301
Oct 12 01:18:14 test101 test201 test302
Oct 12 01:19:14 test101 test204 test316
Oct 12 01:20:14 test102 test204 test319
Oct 12 01:21:14 test102 error204 test321`

	f, err := os.OpenFile(dataFile,
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		t.Errorf("%+v", err)
		return
	}
	defer f.Close()
	if _, err := f.WriteString(data); err != nil {
		t.Errorf("%+v", err)
		return
	}
	f.Close()

	rs2 := newReports()
	err = rs2.run(jsonFile, 0, "",
		0,
		0, 0,
		0, 0, 3,
		0, "")
	if err != nil {
		t.Errorf("%v", err)
		return
	}

	if rs2.rep == nil || len(rs2.rep) == 0 {
		t.Errorf("No report")
		return
	}
	r2 := rs2.rep["test"]
	if r2 == nil {
		t.Errorf("No report")
		return
	}

	recs = r2.nTopNorm.getRecords()
	if len(recs) != 4 {
		t.Errorf("len is incorrect")
		return
	}
	if recs[0].rowid != 15 {
		t.Errorf("rowid is incorrect")
		return
	}

	recs = r2.nTopNorm.getDiffRecords()
	if len(recs) == 1 {
		t.Errorf("len is incorrect")
		return
	}

	recs = r2.nTopErr.getRecords()
	if len(recs) != 3 {
		t.Errorf("len is incorrect")
		return
	}

	recs = r2.nTopErr.getDiffRecords()
	if len(recs) != 1 {
		t.Errorf("len is incorrect")
		return
	}

}
