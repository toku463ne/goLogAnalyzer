package analyzer

import "github.com/pkg/errors"

func newLogRecords(dataDir string,
	maxBlocks, maxRowsInBlock int) (*logRecords, error) {
	lr := new(logRecords)
	cdb, err := newCircuitDB(dataDir, "logRecords",
		tableDefs["logRecords"], maxBlocks, maxRowsInBlock)
	if err != nil {
		return nil, err
	}
	lr.circuitDB = cdb

	return lr, nil
}

func (lr *logRecords) load() error {
	if lr.dataDir == "" {
		return nil
	}
	cnt := lr.statusTable.Count(nil)
	if cnt <= 0 {
		return nil
	}

	if err := lr.loadCircuitDBStatus(); err != nil {
		return err
	}
	return nil
}

func (lr *logRecords) insert(rowID int64, score float64, record string, lastEpoch int64) error {
	if lr.dataDir == "" {
		return nil
	}
	if err := lr.insertRow([]string{"rowID", "score", "epoch", "record"},
		rowID, score, lastEpoch, record); err != nil {
		return errors.WithStack(err)
	}
	lr.rowNo++

	if lr.maxRowsInBlock > 0 && lr.rowNo >= lr.maxRowsInBlock {
		if lr.dataDir != "" {
			lr.lastEpoch = lastEpoch
			if err := lr.currTable.Flush(); err != nil {
				return errors.WithStack(err)
			}
		}
		if err := lr.nextBlock(); err != nil {
			return err
		}
	}
	return nil
}
