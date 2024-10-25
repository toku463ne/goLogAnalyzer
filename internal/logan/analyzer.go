package logan

import (
	"errors"
	"fmt"
	"goLogAnalyzer/pkg/csvdb"
	"goLogAnalyzer/pkg/filepointer"
	"goLogAnalyzer/pkg/utils"
	"math"

	"github.com/sirupsen/logrus"
)

type Analyzer struct {
	*csvdb.CsvDB
	dataDir                  string
	logPath                  string
	logFormat                string
	timestampLayout          string
	blockSize                int
	maxBlocks                int
	keepPeriod               int64
	keepUnit                 int
	configTable              *csvdb.Table
	lastStatusTable          *csvdb.Table
	trans                    *trans
	fp                       *filepointer.FilePointer
	searchRegex, exludeRegex []string
	lastFileEpoch            int64
	lastFileRow              int
	rowID                    int64
	readOnly                 bool
	linesProcessed           int
	termCountBorderRate      float64
	termCountBorder          int
	keywords                 []string
	ignorewords              []string
	customLogGroups          []string
}

func NewAnalyzer(dataDir, logPath, logFormat, timestampLayout string,
	searchRegex, exludeRegex []string,
	maxBlocks, blockSize int,
	keepPeriod int64, keepUnit int,
	termCountBorderRate float64,
	termCountBorder int,
	keywords, ignorewords, customLogGroups []string,
	readOnly bool) (*Analyzer, error) {
	a := new(Analyzer)
	a.dataDir = dataDir
	a.logPath = logPath
	a.logFormat = logFormat
	a.keywords = keywords
	a.ignorewords = ignorewords
	a.timestampLayout = timestampLayout
	a.maxBlocks = maxBlocks
	a.blockSize = blockSize
	a.readOnly = readOnly

	// set defaults
	a.keepUnit = utils.CFreqDay
	a.keepPeriod = CDefaultKeepPeriod
	a.termCountBorder = CDefaultTermCountBorder
	a.termCountBorderRate = CDefaultTermCountBorderRate

	// override passed params
	if keepUnit > 0 {
		a.keepUnit = keepUnit
	}
	if keepPeriod > 0 {
		a.keepPeriod = keepPeriod
	}

	if termCountBorder > 0 {
		a.termCountBorder = termCountBorder
	}
	if termCountBorderRate > 0 {
		a.termCountBorderRate = termCountBorderRate
	}

	// load or init data.
	// Some params will be replaced by params in the DB
	if err := a.open(); err != nil {
		return nil, err
	}

	// for some parameters, the args takes place
	if logPath != "" {
		a.logPath = logPath
	}

	a.customLogGroups = customLogGroups

	return a, nil
}

func LoadAnalyzer(dataDir, logPath string,
	termCountBorderRate float64,
	termCountBorder int,
	customLogGroups []string,
	readOnly bool) (*Analyzer, error) {
	a := new(Analyzer)

	if dataDir == "" {
		return nil, errors.New("no data to load")
	}
	if !utils.PathExist(fmt.Sprintf("%s/lastStatus.tbl.ini", dataDir)) {
		return nil, errors.New("no data to load")
	}

	a.dataDir = dataDir
	a.logPath = logPath
	a.readOnly = readOnly

	// set defaults
	a.termCountBorder = CDefaultTermCountBorder
	a.termCountBorderRate = CDefaultTermCountBorderRate

	if termCountBorder > 0 {
		a.termCountBorder = termCountBorder
	}
	if termCountBorderRate > 0 {
		a.termCountBorderRate = termCountBorderRate
	}

	// load data.
	// Some params will be replaced by params in the DB
	if err := a.open(); err != nil {
		return nil, err
	}

	// for some parameters, the args takes place
	if logPath != "" {
		a.logPath = logPath
	}

	a.customLogGroups = customLogGroups

	return a, nil
}

func (a *Analyzer) Close() {
	if a == nil {
		return
	}

	if a.fp != nil {
		a.fp.Close()
	}
	if a.trans != nil {
		a.trans.close()
	}
}

func (a *Analyzer) Purge() error {
	a.Close()
	if err := utils.RemoveDirectory(a.dataDir); err != nil {
		return err
	}
	return nil
}

// Register terms and convert log lines to logGroups
func (a *Analyzer) Feed(targetLinesCnt int) error {
	targetLinesCnt, err := a._registerTerms(targetLinesCnt)
	if err != nil {
		return err
	}
	if err := a._registerLogGroups(targetLinesCnt); err != nil {
		return err
	}
	return nil
}

func (a *Analyzer) load() error {
	if err := a.loadKeywords(); err != nil {
		return err
	}

	if err := a.trans.load(); err != nil {
		return err
	}
	return nil
}

func (a *Analyzer) open() error {
	if a.dataDir == "" {
		a.initBlocks()
		if err := a.init(); err != nil {
			return err
		}
	} else {
		if utils.PathExist(a.dataDir) {
			if err := a.loadStatus(); err != nil {
				return err
			}
			if err := a.init(); err != nil {
				return err
			}
			if err := a.load(); err != nil {
				return err
			}
		} else {
			if err := a.init(); err != nil {
				return err
			}
			if err := a.prepareDB(); err != nil {
				return err
			}
			if err := a.saveLastStatus(); err != nil {
				return err
			}
			if err := a.saveConfig(); err != nil {
				return err
			}
			if err := a.saveKeywords(); err != nil {
				return err
			}
		}
	}
	return nil
}

func (a *Analyzer) prepareDB() error {
	d, err := csvdb.NewCsvDB(a.dataDir)
	if err != nil {
		return err
	}

	ct, err := d.CreateTableIfNotExists("config", tableDefs["config"], false, 1, 1)
	if err != nil {
		return err
	}
	a.configTable = ct

	ls, err := d.CreateTableIfNotExists("lastStatus", tableDefs["lastStatus"], false, 1, 1)
	if err != nil {
		return err
	}
	a.lastStatusTable = ls

	a.CsvDB = d
	return nil
}

func (a *Analyzer) init() error {
	if a.dataDir != "" && !a.readOnly {
		if err := utils.EnsureDir(a.dataDir); err != nil {
			return err
		}
	}
	trans, err := newTrans(a.dataDir, a.logFormat, a.timestampLayout,
		a.maxBlocks, a.blockSize, a.keepUnit, a.keepPeriod,
		a.termCountBorderRate, a.termCountBorder,
		a.searchRegex, a.exludeRegex,
		a.keywords, a.ignorewords, a.customLogGroups, true, a.readOnly,
	)
	if err != nil {
		return err
	}
	a.trans = trans
	return nil
}

func (a *Analyzer) saveLastStatus() error {
	if a.dataDir == "" || a.readOnly {
		return nil
	}

	var epoch int64
	rowNo := 0
	if a.fp != nil {
		epoch = a.fp.CurrFileEpoch()
		rowNo = a.fp.Row()
	} else {
		epoch = 0
		rowNo = 0
	}

	err := a.lastStatusTable.Upsert(nil, map[string]interface{}{
		"lastRowId":     a.rowID,
		"lastFileEpoch": epoch,
		"lastFileRow":   rowNo,
	})

	return err
}

func (a *Analyzer) loadStatus() error {
	if a.dataDir != "" {
		if err := a.prepareDB(); err != nil {
			return err
		}
	}

	if err := a.configTable.Select1Row(nil,
		tableDefs["config"],
		&a.logPath,
		&a.blockSize, &a.maxBlocks,
		&a.keepPeriod, &a.keepUnit,
		&a.termCountBorderRate,
		&a.termCountBorder,
		&a.timestampLayout,
		&a.logFormat); err != nil {
		return err
	}

	if a.lastFileEpoch == 0 {
		if err := a.lastStatusTable.Select1Row(nil,
			[]string{"lastRowId", "lastFileEpoch", "lastFileRow"},
			&a.rowID, &a.lastFileEpoch, &a.lastFileRow); err != nil {
			return err
		}

	}

	return nil
}

func (a *Analyzer) saveConfig() error {
	if a.readOnly {
		return nil
	}

	if err := a.configTable.Upsert(nil, map[string]interface{}{
		"logPath":             a.logPath,
		"blockSize":           a.blockSize,
		"maxBlocks":           a.maxBlocks,
		"keepPeriod":          a.keepPeriod,
		"keepUnit":            a.keepUnit,
		"termCountBorderRate": a.termCountBorderRate,
		"termCountBorder":     a.termCountBorder,
		"timestampLayout":     a.timestampLayout,
		"logFormat":           a.logFormat,
	}); err != nil {
		return err
	}
	return nil
}

func (a *Analyzer) getKeywordsFilePath() string {
	return fmt.Sprintf("%s/keywords.txt", a.dataDir)
}
func (a *Analyzer) getIgnorewordsFilePath() string {
	return fmt.Sprintf("%s/ignorewords.txt", a.dataDir)
}

func (a *Analyzer) saveKeywords() error {
	if err := utils.Slice2File(a.keywords, a.getKeywordsFilePath()); err != nil {
		return err
	}
	return utils.Slice2File(a.keywords, a.getIgnorewordsFilePath())
}

func (a *Analyzer) loadKeywords() error {
	var err error
	keywordsPath := a.getKeywordsFilePath()
	if utils.PathExist(keywordsPath) {
		a.keywords, err = utils.ReadFile2Slice(keywordsPath)
		if err != nil {
			return err
		}
	}
	ignorewordsPath := a.getIgnorewordsFilePath()
	if utils.PathExist(ignorewordsPath) {
		a.ignorewords, err = utils.ReadFile2Slice(ignorewordsPath)
		if err != nil {
			return err
		}
	}
	return nil
}

func (a *Analyzer) initBlocks() {
	if a.maxBlocks > 0 && a.blockSize > 0 {
		if a.trans != nil {
			a.trans.setBlockSize(a.blockSize)
			a.trans.setMaxBlocks(a.maxBlocks)
		}
		return
	}

	if a.trans == nil || a.trans.maxCountByBlock == 0 {
		return
	}

	maxCountByBlock := a.trans.maxCountByBlock
	keepPeriod := a.keepPeriod
	if keepPeriod == 0 {
		keepPeriod = 30
	}

	if a.blockSize == 0 {
		if maxCountByBlock < 3000 {
			a.blockSize = 10000
		} else if maxCountByBlock < 30000 {
			a.blockSize = 100000
		} else if maxCountByBlock < 300000 {
			a.blockSize = 100000
		} else {
			a.blockSize = 1000000
		}
	}

	if a.maxBlocks == 0 {
		n := int(math.Ceil(float64(a.trans.maxCountByBlock) / float64(a.blockSize)))
		a.maxBlocks = n * int(a.keepPeriod)
	}
	a.trans.setBlockSize(a.blockSize)
	a.trans.setMaxBlocks(a.maxBlocks)

}

func (a *Analyzer) _commit(completed bool) error {
	if a.readOnly {
		return nil
	}
	if a.dataDir == "" {
		return nil
	}
	if err := a.trans.commit(completed); err != nil {
		return err
	}
	if err := a.saveConfig(); err != nil {
		return err
	}
	if err := a.saveLastStatus(); err != nil {
		return err
	}
	if err := a.saveKeywords(); err != nil {
		return err
	}

	return nil
}

func (a *Analyzer) _initFilePointer() error {
	var err error
	if a.fp == nil || !a.fp.IsOpen() {
		a.fp, err = filepointer.NewFilePointer(a.logPath, a.lastFileEpoch, a.lastFileRow)
		if err != nil {
			return err
		}
		if err := a.fp.Open(); err != nil {
			return err
		}
	}
	return nil
}

func (a *Analyzer) _registerTerms(targetLinesCnt int) (int, error) {
	linesProcessed := 0

	if err := a._initFilePointer(); err != nil {
		return -1, err
	}

	for a.fp.Next() {
		if linesProcessed > 0 && linesProcessed%cLogPerLines == 0 {
			logrus.Infof("processed %d lines", linesProcessed)
		}

		line := a.fp.Text()
		if line == "" {
			//linesProcessed++
			continue
		}

		a.trans.lineToTerms(line, 1)
		linesProcessed++

		if targetLinesCnt > 0 && linesProcessed >= targetLinesCnt {
			break
		}
	}

	a.fp.Close()
	a.initBlocks()
	a.trans.initCounters()

	return linesProcessed, nil
}

func (a *Analyzer) _registerLogGroups(targetLinesCnt int) error {
	linesProcessed := 0

	if err := a._initFilePointer(); err != nil {
		return err
	}

	for a.fp.Next() {
		if linesProcessed > 0 && linesProcessed%cLogPerLines == 0 {
			logrus.Infof("processed %d lines", linesProcessed)
		}

		line := a.fp.Text()
		if line == "" {
			//linesProcessed++
			continue
		}

		if err := a.trans.lineToLogGroup(line, 1); err != nil {
			return err
		}
		a.rowID++
		if a.fp.IsEOF && (!a.fp.IsLastFile()) {
			if err := a.saveLastStatus(); err != nil {
				return err
			}
		}
		linesProcessed++

		if targetLinesCnt > 0 && linesProcessed >= targetLinesCnt {
			break
		}
	}

	a.fp.Close()

	if !a.readOnly {
		if err := a._commit(false); err != nil {
			return err
		}
		logrus.Infof("processed %d lines", linesProcessed)
	}
	a.linesProcessed = linesProcessed

	return nil
}
