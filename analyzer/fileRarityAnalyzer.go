package analyzer

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/go-ini/ini"
)

type fileRarityAnalyzer struct {
	rarityAnalyzer
	logPathRegex        string
	currLogPathRegex    string
	currFilterRe        string
	currXFilterRe       string
	currLinesInBlock    int
	currMaxBlocks       int
	currRarityThreshold float64
	fp                  *filePointer
}

func newFileRarityAnalyzer() *fileRarityAnalyzer {
	a := new(fileRarityAnalyzer)

	a.linesInBlock = cDefaultBlockSize
	a.maxBlocks = cDefaultMaxBlocks
	a.rarityThreshold = cDefaultRarityThreshold

	a.outputFunc = func(name string, rowID int64,
		scoreThreshold float64,
		score, scoreGap, scoreAvg, scoreStd float64,
		cnt int,
		text []string) {
		if verbose || scoreGap > scoreThreshold {
			msg := fmt.Sprintf("%s %8d gap=%3.2f | %s\n",
				name,
				rowID,
				scoreGap,
				text[0],
			)
			fmt.Printf(msg)
		}
	}

	a.countTargetLines = func() (int, error) {
		targetLinesCnt := 0
		if err := a.pointerOpen(); err != nil {
			return -1, err
		}
		defer a.fp.close()
		for a.fp.next() {
			targetLinesCnt++
		}
		if err := a.fp.err(); err != nil && err != io.EOF {
			return -1, err
		}
		return targetLinesCnt, nil
	}

	a.setTargetLinesCnt = func(targetLinesCnt int) error {
		if a.maxTargetLinesCnt == 0 && a.logPathRegex != "" {
			maxTargetLinesCnt, err := a.countTargetLines()
			if err != nil {
				return err
			}
			a.maxTargetLinesCnt = maxTargetLinesCnt
		}

		if targetLinesCnt <= 0 {
			a.targetLinesCnt = a.maxTargetLinesCnt
			return nil
		}
		a.targetLinesCnt += targetLinesCnt
		if a.targetLinesCnt > a.maxTargetLinesCnt {
			a.targetLinesCnt = a.maxTargetLinesCnt
		}
		return nil
	}

	a.pointerNext = func() bool {
		return a.fp.next()
	}
	a.pointerText = func() []string {
		return []string{a.fp.text()}
	}
	a.pointerOpen = func() error {
		if a.fp == nil || !a.fp.isOpen() {
			a.fp = newFilePointer(a.logPathRegex, a.lastFileEpoch, a.lastFileRow)
			if err := a.fp.open(); err != nil {
				return err
			}
		}
		return nil
	}
	a.pointerClose = func() {
		if a.fp != nil {
			a.fp.close()
		}
	}
	a.pointerCurrEpoch = func() int64 {
		return a.fp.currFileEpoch()
	}
	a.pointerCurrName = func() string {
		return a.fp.currFile()
	}
	a.pointerPos = func() int {
		return a.fp.row()
	}
	a.pointerErr = func() error {
		return a.fp.err()
	}

	return a
}

func newFileRarityAnalyzerByVars(logPathRegex,
	rootDir string,
	filterRe, xFilterRe string,
	rarityThreshold float64,
	linesInBlock, maxBlocks int) (*fileRarityAnalyzer, error) {

	a := newFileRarityAnalyzer()
	a.useDB = false
	if rootDir != "" {
		a.rootDir = rootDir
		a.name = filepath.Base(rootDir)
		a.useDB = true
		a.linesInBlock = cDefaultBlockSizeNoDb
		a.maxBlocks = cDefaultMaxBlocksNoDb
	}
	a.currLogPathRegex = logPathRegex
	a.currFilterRe = filterRe
	a.currXFilterRe = xFilterRe
	a.currRarityThreshold = rarityThreshold
	a.currLinesInBlock = linesInBlock
	a.currMaxBlocks = maxBlocks
	a.haveStatistics = true

	if a.useDB {
		if err := a.loadIni(); err != nil {
			return nil, err
		}
	}
	a.setNewParams()

	if err := a.init(); err != nil {
		return nil, err
	}
	return a, nil
}

func (a *fileRarityAnalyzer) getIniPath() string {
	return fmt.Sprintf("%s/cfg.ini", a.rootDir)
}

func (a *fileRarityAnalyzer) setNewParams() {
	if a.currLogPathRegex != "" {
		a.logPathRegex = a.currLogPathRegex
	}
	if a.currFilterRe != "" {
		a.filterRe = a.currFilterRe
	}
	if a.currXFilterRe != "" {
		a.xFilterRe = a.currXFilterRe
	}
	if a.currRarityThreshold >= 0.0 {
		a.rarityThreshold = a.currRarityThreshold
	}
	if a.currLinesInBlock >= 0 {
		a.linesInBlock = a.currLinesInBlock
	}
	if a.currMaxBlocks >= 0 {
		a.maxBlocks = a.currMaxBlocks
	}
}

func (a *fileRarityAnalyzer) loadIni() error {
	iniFile := a.getIniPath()
	if !pathExist(iniFile) {
		return nil
	}

	cfg, err := ini.Load(iniFile)
	if err != nil {
		return err
	}
	for _, k := range cfg.Section("LogFile").Keys() {
		switch k.Name() {
		case "logPathRegex":
			a.logPathRegex = k.MustString(a.logPathRegex)

		case "linesInBlock":
			a.linesInBlock = k.MustInt(a.linesInBlock)

		case "maxBlocks":
			a.maxBlocks = k.MustInt(a.maxBlocks)

		case "rarityThreshold":
			a.rarityThreshold = k.MustFloat64(a.rarityThreshold)

		case "filterRe":
			a.filterRe = k.MustString(a.filterRe)

		case "xFilterRe":
			a.xFilterRe = k.MustString(a.xFilterRe)
		}
	}
	return nil
}

func (a *fileRarityAnalyzer) SaveIni() error {
	iniFile := a.getIniPath()
	if !pathExist(iniFile) {
		file, err := os.OpenFile(iniFile, os.O_CREATE, 0640)
		if err != nil {
			return err
		}
		defer file.Close()
	}
	cfg, err := ini.Load(iniFile)
	if err != nil {
		return err
	}
	cfg.Section("LogFile").Key("logPathRegex").SetValue(a.logPathRegex)
	cfg.Section("LogFile").Key("linesInBlock").SetValue(fmt.Sprint(a.linesInBlock))
	cfg.Section("LogFile").Key("maxBlocks").SetValue(fmt.Sprint(a.maxBlocks))
	cfg.Section("LogFile").Key("rarityThreshold").SetValue(fmt.Sprint(a.rarityThreshold))
	cfg.Section("LogFile").Key("filterRe").SetValue(a.filterRe)
	cfg.Section("LogFile").Key("xFilterRe").SetValue(a.xFilterRe)

	return cfg.SaveTo(iniFile)
}

/*
func (a *fileRarityAnalyzer) clean() error {
	iniFile := a.getIniPath()
	if pathExist(iniFile) {
		if err := os.Remove(iniFile); err != nil {
			return err
		}
	}
	return a.db.dropAllTables()
}
*/
