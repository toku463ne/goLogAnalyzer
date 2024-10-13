package csvdb

import (
	"goLogAnalyzer/pkg/utils"
	"strconv"

	"github.com/pkg/errors"
)

func _initCircuitDB(dataDir, name string) {
	cirdb, _ := NewCircuitDB(dataDir, name,
		nil, 0, 0, 0, 0, false)
	cirdb.DropAll()
}

func _blockExists(cirdb *CircuitDB, blockNo int) bool {
	cnt := cirdb.CountFromStatusTable(func(v []string) bool {
		return v[ColBlockNo] == strconv.Itoa(blockNo)
	})
	return cnt > 0
}

func _checkCircuitDBStatus(cirdb *CircuitDB, blockNo int, expected []interface{}) error {
	if !_blockExists(cirdb, blockNo) {
		return errors.New("blockNo must be uniq")
	}

	var lastIndex int
	var blockID string
	var rowNo int
	var completed bool
	if err := cirdb.Select1RowFromStatusTable(func(v []string) bool {
		return v[ColBlockNo] == strconv.Itoa(blockNo)
	}, []string{"lastIndex", "blockID", "rowNo", "completed"},
		&lastIndex, &blockID, &rowNo, &completed); err != nil {
		return err
	}

	if err := utils.GetGotExpErr("lastIndex", lastIndex, expected[0]); err != nil {
		return err
	}
	if err := utils.GetGotExpErr("blockID", blockID, expected[1]); err != nil {
		return err
	}
	if err := utils.GetGotExpErr("rowNo", rowNo, expected[2]); err != nil {
		return err
	}
	if err := utils.GetGotExpErr("completed", completed, expected[3]); err != nil {
		return err
	}

	return nil
}

func _getItemCountInBlock(cirdb *CircuitDB, blockNo int, item string, colname string) int {
	t, _ := cirdb.GetBlockTable(blockNo)
	if !_blockExists(cirdb, blockNo) {
		return 0
	}

	itemIdx := t.GetColIdx(colname)

	itemCount := 0
	if err := t.Sum(func(v []string) bool {
		return v[itemIdx] == item
	}, "count", &itemCount); err != nil {
		return -1
	}

	return itemCount
}

func _checkItemCountInBlock(cirdb *CircuitDB, blockNo int, item string, colname string, expCnt int) error {
	cnt := _getItemCountInBlock(cirdb, blockNo, item, colname)
	if cnt != expCnt {
		if err := utils.GetGotExpErr(item, cnt, expCnt); err != nil {
			return err
		}
	}
	return nil
}

func _insertRows(cirdb *CircuitDB, cols []string, a [][]interface{}, updateTime int64, goNextBlock bool) error {
	for _, row := range a {
		if err := cirdb.InsertRow(cols, row...); err != nil {
			return err
		}
	}
	if goNextBlock {
		if err := cirdb.Commit(); err != nil {
			return err
		}
		if err := cirdb.NextBlock(updateTime); err != nil {
			return err
		}
	} else {
		if err := cirdb.Commit(); err != nil {
			return err
		}
	}
	return nil
}

func _checkCount(cirdb *CircuitDB, colName, colVal string, expct int) error {
	cnt := cirdb.CountAll(func(row []string) bool {
		return row[cirdb.currTable.GetColIdx(colName)] == colVal
	})
	return utils.GetGotExpErr("cnt", cnt, expct)
}
