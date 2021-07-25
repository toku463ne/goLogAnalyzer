package analyzer

import (
	"fmt"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/pkg/errors"
)

func Test_logRecords(t *testing.T) {
	// checks lastIndex, blockID, rowNo, completed and count of the blockID
	checkCircuitDBStatus := func(lr *logRecords, blockNo int, expected []interface{}) error {
		cnt := lr.count("circuitDBStatus", fmt.Sprintf(`blockNO = %d`, blockNo))
		if cnt != 1 {
			return errors.New("blockNo must be uniq")
		}

		var lastIndex int
		var blockID string
		var rowNo int
		var completed bool
		if err := lr.select1rec(fmt.Sprintf(`SELECT lastIndex, blockID, rowNo, completed 
FROM circuitDBStatus WHERE blockNo = %d;`, blockNo),
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

	dataDir, err := ensureTestDir("logRecords")
	if err != nil {
		t.Errorf("%v", err)
		return
	}

	err = dropDB(dataDir, "logRecords")
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
		err := lr.insertRow(row.rowid, row.score, row.record)
		if err != nil {
			t.Errorf("%v", err)
			return
		}
	}

	cnt := lr.count("sqlite_master", "type = 'table'")
	if cnt != 1 {
		t.Errorf("Number of tables must be 1")
		return
	}

	inRows = []colLogRecords{
		{4, 2.5, "test1"},
		{5, 2.5, "test1"},
		{6, 2.5, "test2"},
	}

	for _, row := range inRows {
		err := lr.insertRow(row.rowid, row.score, row.record)
		if err != nil {
			t.Errorf("%v", err)
			return
		}
	}

	cnt = lr.count("sqlite_master", "type = 'table'")
	if cnt != 2 {
		t.Errorf("Number of tables must be 2")
		return
	}

	cnt = lr.count("circuitDBStatus", "")
	if cnt != 1 {
		t.Errorf("Number of status must be 1")
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
		err := lr.insertRow(row.rowid, row.score, row.record)
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

	cnt = lr.count("sqlite_master", "type = 'table'")
	if cnt != 4 {
		t.Errorf("got=%d expect=%d", cnt, 4)
		return
	}

	cnt = lr.count("BLK0000000000", "score = 1.5")
	if cnt != 0 {
		t.Errorf("got=%d expect=%d", cnt, 0)
		return
	}

	cnt = lr.count("BLK0000000000", "score = 3.5")
	if cnt != 1 {
		t.Errorf("got=%d expect=%d", cnt, 1)
		return
	}

	cnt = lr.count("BLK0000000000", "")
	if cnt != 1 {
		t.Errorf("got=%d expect=%d", cnt, 1)
		return
	}

	cnt = lr.count("BLK0000000001", "score = 3.5")
	if cnt != 4 {
		t.Errorf("Count must be 4")
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
		err := lr.insertRow(row.rowid, row.score, row.record)
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

	cnt = lr.count("BLK0000000000", "")
	if cnt != 5 {
		t.Errorf("got=%d expect=%d", cnt, 5)
		return
	}

	rows, err := lr.selectRows2([]string{"rowId", "score"}, "score > 1.0",
		"score desc", 3, []int{0, 1})
	if err != nil {
		t.Errorf("%v", err)
		return
	}
	cnt = 0
	for rows.Next() {
		//var rowId int
		//var score float64
		//rows.Scan(&rowId, &score)
		//fmt.Printf("rowId=%d score=%f", rowId, score)
		cnt++
	}
	if err := getGotExpErr("limit count", cnt, 3); err != nil {
		t.Errorf("%v", err)
		return
	}

	lr.close()
}
