package analyzer

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/pkg/errors"
)

func Test_logRecords(t *testing.T) {
	blockIdx := getColIdx("circuitDBStatus", "blockNo")
	// checks lastIndex, blockID, rowNo, completed and count of the blockID
	checkCircuitDBStatus := func(lr *logRecords, blockNo int, expected []interface{}) error {
		cnt := lr.statusTable.Count(func(v []string) bool {
			return v[blockIdx] == strconv.Itoa(blockNo)
		})

		if cnt != 1 {
			return errors.New("blockNo must be uniq")
		}

		var lastIndex int
		var blockID string
		var rowNo int
		var completed bool
		if err := lr.statusTable.Select1Row(func(v []string) bool {
			return v[blockIdx] == strconv.Itoa(blockNo)
		}, []string{"lastIndex", "blockID", "rowNo", "completed"},
			&lastIndex, &blockID, &rowNo, &completed); err != nil {
			return err
		}

		if err := getGotExpErr("lastIndex", lastIndex, expected[0]); err != nil {
			return err
		}
		if err := getGotExpErr("blockID", blockID, expected[1]); err != nil {
			return err
		}
		if err := getGotExpErr("rowNo", rowNo, expected[2]); err != nil {
			return err
		}
		if err := getGotExpErr("completed", completed, expected[3]); err != nil {
			return err
		}

		return nil
	}

	checkScore := func(lr *logRecords, tableName string, score float64, want int) error {
		scoreIdx := getColIdx("logRecords", "score")
		var cnt int
		g, err := lr.GetGroup("logRecords")
		if err != nil {
			return nil
		}
		t, err := g.GetTable(tableName)
		if err != nil {
			return nil
		}
		if score >= 0 {
			cnt = t.Count(func(v []string) bool {
				_score, _ := strconv.ParseFloat(v[scoreIdx], 64)
				return score == _score
			})
		} else {
			cnt = t.Count(nil)
		}
		return getGotExpErr(fmt.Sprintf("score %f count", score), cnt, want)
	}

	dataDir, err := ensureTestDir("logrecTest")
	if err != nil {
		t.Errorf("%v", err)
		return
	}

	maxBlocks := 3
	maxRowsInBlock := 5
	useGzipInCircuitTables = false
	//tl, err := newTableLogRecords(dataDir, maxBlocks, maxRowsInBlock)
	lr, err := newLogRecords(dataDir, maxBlocks, maxRowsInBlock)
	if err != nil {
		t.Errorf("%v", err)
		return
	}
	lr.DropAll()

	lr, err = newLogRecords(dataDir, maxBlocks, maxRowsInBlock)
	if err != nil {
		t.Errorf("%v", err)
		return
	}

	inRows := []colLogRecord{
		{1, 1.5, "test1", nil, 0, nil},
		{2, 1.5, "test1", nil, 0, nil},
		{3, 1.5, "test1", nil, 0, nil},
	}

	for _, row := range inRows {
		err := lr.insert(row.rowid, row.score, row.record, 1)
		if err != nil {
			t.Errorf("%v", err)
			return
		}
	}

	cnt := lr.currTable.Count(nil)
	if err := getGotExpErr("first count", cnt, 0); err != nil {
		t.Errorf("%v", err)
		return
	}

	inRows = []colLogRecord{
		{4, 2.5, "test1", nil, 0, nil},
		{5, 2.5, "test1", nil, 0, nil},
		{6, 2.5, "test2", nil, 0, nil},
	}

	for _, row := range inRows {
		err := lr.insert(row.rowid, row.score, row.record, 2)
		if err != nil {
			t.Errorf("%v", err)
			return
		}
	}

	if err := checkScore(lr, "BLK0000000000", -1, 5); err != nil {
		t.Errorf("%v", err)
		return
	}

	blockIDs, err := lr.getBlockNos()
	if err != nil {
		t.Errorf("%v", err)
		return
	}
	if err := getGotExpErr("number of blockIDs", len(blockIDs), 1); err != nil {
		t.Errorf("%v", err)
		return
	}

	if err := checkCircuitDBStatus(lr, 0,
		[]interface{}{0, "BLK0000000000", 5, true}); err != nil {
		t.Errorf("%v", err)
		return
	}

	inRows = []colLogRecord{
		{7, 3.5, "test2", nil, 0, nil},
		{8, 3.5, "test2", nil, 0, nil},
		{9, 3.5, "test2", nil, 0, nil},
		{10, 3.5, "test2", nil, 0, nil},
		{11, 3.5, "test3", nil, 0, nil},
		{12, 3.5, "test3", nil, 0, nil},
		{13, 3.5, "test3", nil, 0, nil},
		{14, 3.5, "test3", nil, 0, nil},
		{15, 3.5, "test3", nil, 0, nil},
		{16, 3.5, "test4", nil, 0, nil},
	}

	for _, row := range inRows {
		err := lr.insert(row.rowid, row.score, row.record, 3)
		if err != nil {
			t.Errorf("%v", err)
			return
		}
	}
	if err := checkCircuitDBStatus(lr, 0,
		[]interface{}{0, "BLK0000000000", 5, true}); err != nil {
		t.Errorf("%v", err)
		return
	}
	if err := checkCircuitDBStatus(lr, 1,
		[]interface{}{1, "BLK0000000001", 5, true}); err != nil {
		t.Errorf("%v", err)
		return
	}
	if err := checkCircuitDBStatus(lr, 2,
		[]interface{}{2, "BLK0000000002", 5, true}); err != nil {
		t.Errorf("%v", err)
		return
	}

	err = lr.commit(false)
	if err != nil {
		t.Errorf("%v", err)
		return
	}

	if err := checkCircuitDBStatus(lr, 0,
		[]interface{}{3, "BLK0000000000", 1, false}); err != nil {
		t.Errorf("%v", err)
		return
	}

	blockIDs, err = lr.getBlockNos()
	if err != nil {
		t.Errorf("%v", err)
		return
	}
	if err := getGotExpErr("number of blockIDs", len(blockIDs), 3); err != nil {
		t.Errorf("%v", err)
		return
	}

	if err := checkScore(lr, "BLK0000000000", 1.5, 0); err != nil {
		t.Errorf("%v", err)
		return
	}

	if err := checkScore(lr, "BLK0000000000", 3.5, 1); err != nil {
		t.Errorf("%v", err)
		return
	}

	if err := checkScore(lr, "BLK0000000000", -1.0, 1); err != nil {
		t.Errorf("%v", err)
		return
	}

	if err := checkScore(lr, "BLK0000000001", 3.5, 4); err != nil {
		t.Errorf("%v", err)
		return
	}

	lr = nil

	lr, err = newLogRecords(dataDir, maxBlocks, maxRowsInBlock)
	if err != nil {
		t.Errorf("%v", err)
		return
	}
	if err := lr.load(); err != nil {
		t.Errorf("%v", err)
		return
	}

	inRows = []colLogRecord{
		{17, 4.5, "test4", nil, 0, nil},
		{18, 4.5, "test4", nil, 0, nil},
		{19, 4.5, "test4", nil, 0, nil},
		{20, 4.5, "test4", nil, 0, nil},
		{21, 5.5, "test5", nil, 0, nil},
	}

	for _, row := range inRows {
		err := lr.insert(row.rowid, row.score, row.record, 4)
		if err != nil {
			t.Errorf("%v", err)
			return
		}
	}
	if err := checkCircuitDBStatus(lr, 0,
		[]interface{}{3, "BLK0000000000", 5, true}); err != nil {
		t.Errorf("%v", err)
		return
	}

	if err := checkScore(lr, "BLK0000000000", -1.0, 5); err != nil {
		t.Errorf("%v", err)
		return
	}

	lr = nil
}
