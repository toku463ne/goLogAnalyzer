package analyzer

import (
	"fmt"

	"github.com/pkg/errors"
)

func newLogRecords(dataDir string,
	maxBlocks, maxRowsInBlock int) (*logRecords, error) {
	rt, err := newCircuitDB(dataDir, "logRecords", maxBlocks, maxRowsInBlock)
	if err != nil {
		return nil, err
	}
	lr := new(logRecords)
	lr.circuitDB = rt
	lr.rows = make([]colLogRecords, maxRowsInBlock)
	lr.startRowNo = lr.rowNo
	return lr, nil
}

func (lr *logRecords) commit(completed bool) error {
	logInfo("logRecords.commit() start")

	if lr.rowNo == 0 {
		return nil
	}
	if err := lr.dropBlockIfCompeted(lr.blockNo); err != nil {
		return err
	}
	if err := lr.createBlock(lr.blockNo); err != nil {
		return err
	}

	sqlstr := fmt.Sprintf(`INSERT INTO %s(rowId, score, record) VALUES(?,?,?);`,
		lr.getBlockTableName(lr.blockNo))

	stmt, err := lr.conn.Prepare(sqlstr)
	if err != nil {
		return errors.WithStack(err)
	}

	for pos, row := range lr.rows {
		if pos >= lr.rowNo {
			break
		}
		if pos < lr.startRowNo {
			continue
		}
		_, err := stmt.Exec(row.rowid, row.score, row.record)
		if err != nil {
			return errors.WithStack(err)
		}
	}
	lr.startRowNo = 0

	err = lr.updateBlockStatus(completed)
	logInfo("logRecords.commit() %d rows completed")

	return err
}

func (lr *logRecords) insertRow(rowID int64, score float64, record string) error {
	lr.rows[lr.rowNo] = colLogRecords{rowID, score, record}
	lr.rowNo++
	if lr.maxRowsInBlock > 0 && lr.rowNo >= lr.maxRowsInBlock {
		if lr.dataDir != "" {
			if err := lr.commit(true); err != nil {
				return err
			}
		}
		lr.nextBlock()
	}
	return nil
}
