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
		if score >= 0 {
			cnt = lr.count(tableName, func(v []string) bool {
				_score, _ := strconv.ParseFloat(v[scoreIdx], 64)
				return score == _score
			})
		} else {
			cnt = lr.count(tableName, nil)
		}
		return getGotExpErr(fmt.Sprintf("score %f count", score), cnt, want)
	}

	dataDir, err := ensureTestDir("logRecords")
	if err != nil {
		t.Errorf("%v", err)
		return
	}

	err = dropCsvDB(dataDir, "logRecords")
	if err != nil {
		t.Errorf("%v", err)
		return
	}

	maxBlocks := 3
	maxRowsInBlock := 5
	//tl, err := newTableLogRecords(dataDir, maxBlocks, maxRowsInBlock)
	lr, err := newLogRecords(dataDir, maxBlocks, maxRowsInBlock)
	if err != nil {
		t.Errorf("%v", err)
		return
	}
	inRows := []colLogRecords{
		{1, 1.5, "test1"},
		{2, 1.5, "test1"},
		{3, 1.5, "test1"},
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

	inRows = []colLogRecords{
		{4, 2.5, "test1"},
		{5, 2.5, "test1"},
		{6, 2.5, "test2"},
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

	inRows = []colLogRecords{
		{7, 3.5, "test2"},
		{8, 3.5, "test2"},
		{9, 3.5, "test2"},
		{10, 3.5, "test2"},
		{11, 3.5, "test3"},
		{12, 3.5, "test3"},
		{13, 3.5, "test3"},
		{14, 3.5, "test3"},
		{15, 3.5, "test3"},
		{16, 3.5, "test4"},
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

	lr.close()

	lr, err = newLogRecords(dataDir, maxBlocks, maxRowsInBlock)
	if err != nil {
		t.Errorf("%v", err)
		return
	}

	inRows = []colLogRecords{
		{17, 4.5, "test4"},
		{18, 4.5, "test4"},
		{19, 4.5, "test4"},
		{20, 4.5, "test4"},
		{21, 5.5, "test5"},
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

	lr.close()
}
