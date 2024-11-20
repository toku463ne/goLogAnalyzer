package logan

import (
	"goLogAnalyzer/pkg/utils"
	"os"
)

func (a *Analyzer) _checkLastStatus(expect_lastRowId, expect_lastFileRow int, filepath string) error {
	a.loadStatus()

	if err := utils.GetGotExpErr("lastRowId", a.LastFileRow, expect_lastRowId); err != nil {
		return err
	}

	if err := utils.GetGotExpErr("lastFileRow", a.LastFileRow, expect_lastFileRow); err != nil {
		return err
	}

	file, err := os.Stat(filepath)
	if err != nil {
		return err
	}
	expect_lastFileEpoch := file.ModTime().Unix()

	if err := utils.GetGotExpErr("lastFileEpoch", a.LastFileEpoch, expect_lastFileEpoch); err != nil {
		return err
	}

	return nil
}

func (a *Analyzer) _checkConfig(expect_logPath string,
	expect_blockSize, expect_maxBlocks int,
	expect_keepPeriod int64, expect_keepUnit int64,
	expect_termCountBorderRate float64, expect_termCountBorder int,
	expect_minMatchRate float64,
	expect_timestampLayout string, expect_useUtcTime, expect_ignoreNumbers bool,
	expect_separator string,
	expect_logFormat string) error {

	a.loadConfig()

	if err := utils.GetGotExpErr("logPath", a.LogPath, expect_logPath); err != nil {
		return err
	}
	if err := utils.GetGotExpErr("blockSize", a.BlockSize, expect_blockSize); err != nil {
		return err
	}
	if err := utils.GetGotExpErr("maxBlocks", a.MaxBlocks, expect_maxBlocks); err != nil {
		return err
	}
	if err := utils.GetGotExpErr("keepPeriod", a.KeepPeriod, expect_keepPeriod); err != nil {
		return err
	}
	if err := utils.GetGotExpErr("keepUnit", a.UnitSecs, expect_keepUnit); err != nil {
		return err
	}
	if err := utils.GetGotExpErr("termCountBorderRate", a.TermCountBorderRate, expect_termCountBorderRate); err != nil {
		return err
	}
	if err := utils.GetGotExpErr("termCountBorder", a.TermCountBorder, expect_termCountBorder); err != nil {
		return err
	}
	if err := utils.GetGotExpErr("minMatchRate", a.MinMatchRate, expect_minMatchRate); err != nil {
		return err
	}
	if err := utils.GetGotExpErr("timestampLayout", a.TimestampLayout, expect_timestampLayout); err != nil {
		return err
	}
	if err := utils.GetGotExpErr("useUtcTime", a.UseUtcTime, expect_useUtcTime); err != nil {
		return err
	}
	if err := utils.GetGotExpErr("ignoreNumbers", a.IgnoreNumbers, expect_ignoreNumbers); err != nil {
		return err
	}
	if err := utils.GetGotExpErr("separator", a.Separators, expect_separator); err != nil {
		return err
	}
	if err := utils.GetGotExpErr("logFormat", a.LogFormat, expect_logFormat); err != nil {
		return err
	}

	return nil
}
