package logan

import (
	"goLogAnalyzer/pkg/utils"
	"os"
)

func (a *Analyzer) _checkLastStatusTable(expect_lastRowId, expect_lastFileRow int, filepath string) error {
	var got_lastRowId int
	var got_lastFileEpoch int64
	var got_lastFileRow int
	a.lastStatusTable.Select1Row(nil,
		tableDefs["lastStatus"],
		&got_lastRowId, &got_lastFileEpoch, &got_lastFileRow)

	if err := utils.GetGotExpErr("lastRowId", got_lastRowId, expect_lastRowId); err != nil {
		return err
	}

	if err := utils.GetGotExpErr("lastFileRow", got_lastFileRow, expect_lastFileRow); err != nil {
		return err
	}

	file, err := os.Stat(filepath)
	if err != nil {
		return err
	}
	expect_astFileEpoch := file.ModTime().Unix()

	if err := utils.GetGotExpErr("lastFileEpoch", got_lastFileEpoch, expect_astFileEpoch); err != nil {
		return err
	}

	return nil
}

func (a *Analyzer) _checkConfigTable(expect_logPath string,
	expect_blockSize, expect_maxBlocks int,
	expect_keepPeriod int64, expect_keepUnit int64,
	expect_termCountBorderRate float64, expect_termCountBorder int,
	expect_minMatchRate float64,
	expect_timestampLayout string, expect_useUtcTime bool,
	expect_separator string,
	expect_logFormat string) error {

	var got_logPath string
	var got_blockSize, got_maxBlocks int
	var got_keepPeriod int64
	var got_keepUnit int64
	var got_termCountBorderRate float64
	var got_termCountBorder int
	var got_minMatchRate float64
	var got_timestampLayout, got_logFormat string
	var got_useUtcTime bool
	var got_separator string
	a.configTable.Select1Row(nil,
		tableDefs["config"],
		&got_logPath,
		&got_blockSize, &got_maxBlocks,
		&got_keepPeriod,
		&got_keepUnit,
		&got_termCountBorderRate,
		&got_termCountBorder,
		&got_minMatchRate,
		&got_timestampLayout,
		&got_useUtcTime,
		&got_separator,
		&got_logFormat)

	if err := utils.GetGotExpErr("logPath", got_logPath, expect_logPath); err != nil {
		return err
	}
	if err := utils.GetGotExpErr("blockSize", got_blockSize, expect_blockSize); err != nil {
		return err
	}
	if err := utils.GetGotExpErr("maxBlocks", got_maxBlocks, expect_maxBlocks); err != nil {
		return err
	}
	if err := utils.GetGotExpErr("keepPeriod", got_keepPeriod, expect_keepPeriod); err != nil {
		return err
	}
	if err := utils.GetGotExpErr("keepUnit", got_keepUnit, expect_keepUnit); err != nil {
		return err
	}
	if err := utils.GetGotExpErr("termCountBorderRate", got_termCountBorderRate, expect_termCountBorderRate); err != nil {
		return err
	}
	if err := utils.GetGotExpErr("termCountBorder", got_termCountBorder, expect_termCountBorder); err != nil {
		return err
	}
	if err := utils.GetGotExpErr("minMatchRate", got_minMatchRate, expect_minMatchRate); err != nil {
		return err
	}
	if err := utils.GetGotExpErr("timestampLayout", got_timestampLayout, expect_timestampLayout); err != nil {
		return err
	}
	if err := utils.GetGotExpErr("useUtcTime", got_useUtcTime, expect_useUtcTime); err != nil {
		return err
	}
	if err := utils.GetGotExpErr("separator", got_separator, expect_separator); err != nil {
		return err
	}
	if err := utils.GetGotExpErr("logFormat", got_logFormat, expect_logFormat); err != nil {
		return err
	}

	return nil
}
