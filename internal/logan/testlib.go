package logan

import (
	"goLogAnalyzer/pkg/utils"
	"os"
)

func (a *Analyzer) _checkLastStatusTable(except_lastRowId, except_lastFileRow int64, filepath string) error {
	var got_lastRowId int64
	var got_lastFileEpoch int64
	var got_lastFileRow int64
	a.lastStatusTable.Select1Row(nil,
		tableDefs["lastStatus"],
		&got_lastRowId, &got_lastFileEpoch, &got_lastFileRow)

	if err := utils.GetGotExpErr("lastRowId", got_lastRowId, except_lastRowId); err != nil {
		return err
	}

	if err := utils.GetGotExpErr("lastFileRow", got_lastFileRow, except_lastFileRow); err != nil {
		return err
	}

	file, err := os.Stat(filepath)
	if err != nil {
		return err
	}
	except_astFileEpoch := file.ModTime().Unix()

	if err := utils.GetGotExpErr("lastFileEpoch", got_lastFileEpoch, except_astFileEpoch); err != nil {
		return err
	}

	return nil
}

func (a *Analyzer) _checkConfigTable(except_logPath string,
	except_blockSize, except_maxBlocks int,
	except_keepPeriod int64, except_keepUnit int64,
	except_termCountBorderRate float64, except_termCountBorder int,
	except_timestampLayout, except_logFormat string) error {

	var got_logPath string
	var got_blockSize, got_maxBlocks int
	var got_keepPeriod int
	var got_keepUnit int64
	var got_termCountBorderRate float64
	var got_termCountBorder int
	var got_timestampLayout, got_logFormat string
	a.lastStatusTable.Select1Row(nil,
		tableDefs["config"],
		&got_logPath,
		&got_blockSize, &got_maxBlocks,
		&got_keepPeriod,
		&got_keepUnit,
		&got_termCountBorderRate,
		&got_termCountBorder,
		&got_timestampLayout, got_logFormat)

	if err := utils.GetGotExpErr("ogPath", except_logPath, got_logPath); err != nil {
		return err
	}
	if err := utils.GetGotExpErr("blockSize", got_blockSize, except_blockSize); err != nil {
		return err
	}
	if err := utils.GetGotExpErr("maxBlocks", got_maxBlocks, except_maxBlocks); err != nil {
		return err
	}
	if err := utils.GetGotExpErr("keepPeriod", got_keepPeriod, except_keepPeriod); err != nil {
		return err
	}
	if err := utils.GetGotExpErr("keepUnit", got_keepUnit, except_keepUnit); err != nil {
		return err
	}
	if err := utils.GetGotExpErr("termCountBorderRate", got_termCountBorderRate, except_termCountBorderRate); err != nil {
		return err
	}
	if err := utils.GetGotExpErr("termCountBorder", got_termCountBorder, except_termCountBorder); err != nil {
		return err
	}
	if err := utils.GetGotExpErr("timestampLayout", got_timestampLayout, except_timestampLayout); err != nil {
		return err
	}
	if err := utils.GetGotExpErr("logFormat", got_logFormat, except_logFormat); err != nil {
		return err
	}

	return nil
}
