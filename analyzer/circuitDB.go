package analyzer

import (
	"fmt"
	"strconv"

	"github.com/pkg/errors"
	csvdb "github.com/toku463ne/goLogAnalyzer/csvdb"
)

func newCircuitDB(rootDir, name string, columns []string,
	maxBlocks, maxRowsInBlock int) (*circuitDB, error) {
	cdb := new(circuitDB)
	cdb.name = name
	cdb.maxRowsInBlock = maxRowsInBlock
	cdb.maxBlocks = maxBlocks

	if rootDir == "" {
		return cdb, nil
	}
	cdb.dataDir = fmt.Sprintf("%s/%s", rootDir, name)

	db, err := csvdb.NewCsvDB(cdb.dataDir)
	if err != nil {
		return nil, err
	}
	_, err = db.CreateGroup(name, columns, useGzipInCircuitTables, maxRowsInBlock)
	if err != nil {
		return nil, err
	}

	cdb.writeMode = csvdb.CWriteModeAppend

	st, err := db.CreateTableIfNotExists("circuitDBStatus",
		tableDefs["circuitDBStatus"], false, maxBlocks)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	cdb.statusTable = st

	cdb.CsvDB = db

	t, err := cdb.getBlockTable(cdb.blockNo)
	if err != nil {
		return nil, err
	}
	cdb.currTable = t

	return cdb, nil
}

func (cdb *circuitDB) getBlockTable(blockNo int) (*csvdb.CsvTable, error) {
	blockID := cdb.getBlockTableName(blockNo)
	return cdb.Groups[cdb.name].GetTable(blockID)
}

func (cdb *circuitDB) loadCircuitDBStatus() error {
	if cdb.dataDir == "" {
		return nil
	}

	var lastIndex, lastEpoch int64
	var blockNo, rowNo int
	var completed bool

	t := cdb.statusTable
	if err := t.Max(nil, "lastIndex", &lastIndex); err != nil {
		return errors.WithStack(err)
	}

	lastIndexStr := strconv.Itoa(int(lastIndex))
	idx := getColIdx("circuitDBStatus", "lastIndex")
	if err := t.Select1Row(func(v []string) bool {
		return v[idx] == lastIndexStr
	}, []string{"blockNo", "rowNo", "lastEpoch", "completed"},
		&blockNo, &rowNo, &lastEpoch, &completed); err != nil {
		return errors.WithStack(err)
	}

	cdb.blockNo = blockNo
	cdb.lastIndex = lastIndex
	cdb.lastEpoch = lastEpoch
	cdb.rowNo = rowNo
	cdb.writeMode = csvdb.CWriteModeAppend

	if completed {
		if err := cdb.nextBlock(); err != nil {
			return err
		}
	}

	t, err := cdb.getBlockTable(cdb.blockNo)
	if err != nil {
		return err
	}
	cdb.currTable = t

	return nil
}

func (cdb *circuitDB) nextBlock() error {
	if err := cdb.updateBlockStatus(true); err != nil {
		return err
	}

	cdb.rowNo = 0
	cdb.blockNo++
	if cdb.blockNo >= cdb.maxBlocks {
		cdb.blockNo = 0
	}
	cdb.lastIndex++

	// if don't save to file
	if cdb.dataDir == "" {
		return nil
	}

	cdb.writeMode = "w"

	t, err := cdb.getBlockTable(cdb.blockNo)
	if err != nil {
		return err
	}
	cdb.currTable = t
	//rt.rows = make([][]interface{}, rt.maxRowsInBlock)
	return nil
}

func (cdb *circuitDB) getBlockID(blockNo int) string {
	return fmt.Sprintf("%0"+strconv.Itoa(cMaxBlockDitigs)+"d", blockNo)
}

func (cdb *circuitDB) getBlockTableName(blockNo int) string {
	return fmt.Sprintf("BLK%s", cdb.getBlockID(blockNo))
}

func (cdb *circuitDB) insertRow(columns []string, row ...interface{}) error {
	if cdb.dataDir == "" {
		return nil
	}
	if cdb.writeMode == csvdb.CWriteModeWrite {
		if err := cdb.currTable.Delete(nil); err != nil {
			return errors.WithStack(err)
		}
		cdb.writeMode = csvdb.CWriteModeAppend
	}

	if err := cdb.currTable.InsertRow(columns, row...); err != nil {
		return errors.WithStack(err)
	}
	cdb.writeMode = csvdb.CWriteModeAppend
	return nil
}

func (cdb *circuitDB) updateBlockStatus(completed bool) error {
	if cdb.dataDir == "" {
		return nil
	}
	idx := getColIdx("circuitDBStatus", "blockNo")
	blockID := cdb.getBlockTableName(cdb.blockNo)

	if err := cdb.statusTable.Upsert(func(v []string) bool {
		return v[idx] == strconv.Itoa(cdb.blockNo)
	}, map[string]interface{}{
		"lastIndex": cdb.lastIndex,
		"blockNo":   cdb.blockNo,
		"blockID":   blockID,
		"rowNo":     cdb.rowNo,
		"lastEpoch": cdb.lastEpoch,
		"completed": completed,
	}); err != nil {
		return errors.WithStack(err)
	}

	return nil
}

func (cdb *circuitDB) commit(completed bool) error {
	if cdb.dataDir == "" {
		return nil
	}
	if err := cdb.currTable.Flush(); err != nil {
		return errors.WithStack(err)
	}
	if err := cdb.updateBlockStatus(false); err != nil {
		return err
	}
	return nil
}

func (cdb *circuitDB) getBlockNos() ([]int, error) {
	cnt := cdb.statusTable.Count(nil)
	if cnt <= 0 {
		return nil, nil
	}
	blockNos := make([]int, cnt)
	rows, err := cdb.statusTable.SelectRows(nil, []string{"blockNo"})
	if err != nil {
		return nil, err
	}

	if rows == nil {
		return nil, nil
	}

	i := 0
	for rows.Next() {
		if err := rows.Scan(&blockNos[i]); err != nil {
			return nil, err
		}
		i++
	}
	return blockNos, nil
}

func (cdb *circuitDB) selectRows(conditionCheckFunc func([]string) bool,
	blockNos []int, columns []string) (*circuitRows, error) {
	var err error
	if blockNos == nil {
		blockNos, err = cdb.getBlockNos()
		if err != nil {
			return nil, err
		}
	}

	r := new(circuitRows)
	r.groupName = cdb.name
	r.tableNames = make([]string, len(blockNos))
	pos := 0
	for _, blockNo := range blockNos {
		r.tableNames[pos] = cdb.getBlockTableName(blockNo)
		pos++
	}

	r.statusTable = cdb.statusTable
	r.CsvDB = cdb.CsvDB
	r.conditionCheckFunc = conditionCheckFunc
	r.columns = columns
	r.blockIDIdx = getColIdx("circuitDBStatus", "blockID")
	r.completedIdx = getColIdx("circuitDBStatus", "completed")

	r.pos = 0
	return r, nil
}

func (cdb *circuitDB) countAll(conditionCheckFunc func([]string) bool) int {
	blockNos, err := cdb.getBlockNos()
	if err != nil {
		return -1
	}
	cnt := 0
	for _, blockNo := range blockNos {
		t, err := cdb.Groups[cdb.name].GetTable(cdb.getBlockTableName(blockNo))
		if err != nil {
			return -1
		}
		tcnt := t.Count(conditionCheckFunc)
		if tcnt > 0 {
			cnt += tcnt
		}
	}
	return cnt
}
