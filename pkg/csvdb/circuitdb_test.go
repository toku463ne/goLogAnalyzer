package csvdb

import (
	"goLogAnalyzer/pkg/utils"
	"strconv"
	"testing"
)

func Test_CircuitDB_by_blocks(t *testing.T) {
	dataDir, err := utils.InitTestDir("Test_CircuitDB_by_blocks")
	if err != nil {
		t.Errorf("%v", err)
		return
	}

	_initCircuitDB(dataDir, "testcirdb")

	cols := []string{"itemid", "name", "count"}

	maxBlocks := 3
	cirdb, err := NewCircuitDB(dataDir, "testcirdb",
		cols, maxBlocks, 0, 0, 0, false)
	if err != nil {
		t.Errorf("%v", err)
		return
	}

	// block 0
	inRows := [][]interface{}{
		{"row001", "name001", "10"},
		{"row002", "name002", "20"},
		{"row003", "name003", "30"},
	}

	if err := _insertRows(cirdb, cols, inRows, 0, true); err != nil {
		t.Errorf("%v", err)
		return
	}

	if err := _checkItemCountInBlock(cirdb, 0, "name001", "name", 10); err != nil {
		t.Errorf("%v", err)
		return
	}

	if err := _checkCircuitDBStatus(cirdb, 0, []interface{}{0, "BLK0000000000", 3, true}); err != nil {
		t.Errorf("%v", err)
		return
	}

	// block 1
	inRows = [][]interface{}{
		{"row104", "name001", "10"},
		{"row105", "name002", "20"},
		{"row106", "name003", "30"},
	}

	if err := _insertRows(cirdb, cols, inRows, 0, true); err != nil {
		t.Errorf("%v", err)
		return
	}

	if err := _checkItemCountInBlock(cirdb, 1, "name001", "name", 10); err != nil {
		t.Errorf("%v", err)
		return
	}

	if err := _checkCount(cirdb, "name", "name001", 2); err != nil {
		t.Errorf("%v", err)
		return
	}

	// block 2
	inRows = [][]interface{}{
		{"row207", "name001", "10"},
		{"row208", "name002", "20"},
	}
	if err := _insertRows(cirdb, cols, inRows, 0, true); err != nil {
		t.Errorf("%v", err)
		return
	}

	if err := _checkCount(cirdb, "name", "name001", 3); err != nil {
		t.Errorf("%v", err)
		return
	}

	// back to block 0
	inRows = [][]interface{}{
		{"row304", "name001", "5"},
	}
	if err := _insertRows(cirdb, cols, inRows, 0, true); err != nil {
		t.Errorf("%v", err)
		return
	}

	if err := _checkItemCountInBlock(cirdb, 0, "name001", "name", 5); err != nil {
		t.Errorf("%v", err)
		return
	}

	if err := _checkCount(cirdb, "name", "name001", 3); err != nil {
		t.Errorf("%v", err)
		return
	}

	if err := _checkCount(cirdb, "name", "name002", 2); err != nil {
		t.Errorf("%v", err)
		return
	}

	if err := _checkCount(cirdb, "name", "name003", 1); err != nil {
		t.Errorf("%v", err)
		return
	}

	if err := _checkCircuitDBStatus(cirdb, 0, []interface{}{3, "BLK0000000000", 1, true}); err != nil {
		t.Errorf("%v", err)
		return
	}
	if err := _checkCircuitDBStatus(cirdb, 1, []interface{}{1, "BLK0000000001", 3, true}); err != nil {
		t.Errorf("%v", err)
		return
	}
	if err := _checkCircuitDBStatus(cirdb, 2, []interface{}{2, "BLK0000000002", 2, true}); err != nil {
		t.Errorf("%v", err)
		return
	}

	cirdb.CloseAll()

	cirdb, err = NewCircuitDB(dataDir, "testcirdb",
		cols, maxBlocks, 0, 0, 0, false)
	if err != nil {
		t.Errorf("%v", err)
		return
	}

	if err := _checkCount(cirdb, "name", "name001", 3); err != nil {
		t.Errorf("%v", err)
		return
	}

	if err := _checkCount(cirdb, "name", "name002", 2); err != nil {
		t.Errorf("%v", err)
		return
	}

	if err := _checkCount(cirdb, "name", "name003", 1); err != nil {
		t.Errorf("%v", err)
		return
	}
}

func Test_CircuitDB_by_keepPeriod(t *testing.T) {
	dataDir, err := utils.InitTestDir("Test_CircuitDB_by_keepPeriod")
	if err != nil {
		t.Errorf("%v", err)
		return
	}

	_initCircuitDB(dataDir, "testcirdb")

	cols := []string{"itemid", "name", "count"}

	//frequency := utils.CFreqDay
	frequency := int64(3600 * 24)
	keepPeriod := 3
	maxBlocks := 10
	cirdb, err := NewCircuitDB(dataDir, "testcirdb",
		cols, maxBlocks, 0, int64(keepPeriod), frequency, false)
	if err != nil {
		t.Errorf("%v", err)
		return
	}

	// block 0
	d := int64(1728745200) //2024-10-13
	inRows := [][]interface{}{
		{"row001", "name001", "10"},
		{"row002", "name002", "20"},
		{"row003", "name003", "30"},
	}
	if err := _insertRows(cirdb, cols, inRows, d, true); err != nil {
		t.Errorf("%v", err)
		return
	}

	if err := _checkItemCountInBlock(cirdb, 0, "name001", "name", 10); err != nil {
		t.Errorf("%v", err)
		return
	}

	if err := _checkCircuitDBStatus(cirdb, 0, []interface{}{0, "BLK0000000000", 3, true}); err != nil {
		t.Errorf("%v", err)
		return
	}

	// block 1
	d += 3600 * 24 //2024-10-14
	inRows = [][]interface{}{
		{"row104", "name001", "10"},
		{"row105", "name002", "20"},
		{"row106", "name003", "30"},
	}

	if err := _insertRows(cirdb, cols, inRows, d, true); err != nil {
		t.Errorf("%v", err)
		return
	}

	if err := _checkItemCountInBlock(cirdb, 1, "name001", "name", 10); err != nil {
		t.Errorf("%v", err)
		return
	}

	if err := _checkCount(cirdb, "name", "name001", 2); err != nil {
		t.Errorf("%v", err)
		return
	}

	// block 2
	d += 3600 * 24 //2024-10-15
	inRows = [][]interface{}{
		{"row207", "name001", "10"},
		{"row208", "name002", "20"},
	}
	if err := _insertRows(cirdb, cols, inRows, d, true); err != nil {
		t.Errorf("%v", err)
		return
	}

	if err := _checkCount(cirdb, "name", "name001", 3); err != nil {
		t.Errorf("%v", err)
		return
	}

	// block 3
	d += 3600 * 24 //2024-10-16
	inRows = [][]interface{}{
		{"row307", "name001", "10"},
	}
	if err := _insertRows(cirdb, cols, inRows, d, true); err != nil {
		t.Errorf("%v", err)
		return
	}

	// loop until maxblocks
	for blockNo := 4; blockNo < 10; blockNo++ {
		d += 3600 * 24
		inRows = [][]interface{}{
			{"row" + strconv.Itoa(blockNo) + "00", "name001", 10},
		}
		if err := _insertRows(cirdb, cols, inRows, d, true); err != nil {
			t.Errorf("%v", err)
			return
		}

		if err := _checkItemCountInBlock(cirdb, blockNo-keepPeriod, "name001", "name", 0); err != nil {
			t.Errorf("%v", err)
			return
		}

		if err := _checkItemCountInBlock(cirdb, blockNo, "name001", "name", 10); err != nil {
			t.Errorf("%v", err)
			return
		}

		if err := _checkCount(cirdb, "name", "name001", keepPeriod); err != nil {
			t.Errorf("%v", err)
			return
		}
	}

	// return to block 0 at it reached the maxBlocks
	d += 3600 * 24
	inRows = [][]interface{}{
		{"row007", "name001", "10"},
	}
	if err := _insertRows(cirdb, cols, inRows, d, true); err != nil {
		t.Errorf("%v", err)
		return
	}

	if err := _checkItemCountInBlock(cirdb, 0, "name001", "name", 10); err != nil {
		t.Errorf("%v", err)
		return
	}

	if err := _checkCount(cirdb, "name", "name001", keepPeriod); err != nil {
		t.Errorf("%v", err)
		return
	}

}
