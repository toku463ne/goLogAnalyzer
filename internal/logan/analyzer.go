package logan

import (
	"encoding/csv"
	"fmt"
	"goLogAnalyzer/pkg/csvdb"
	"goLogAnalyzer/pkg/filepointer"
	"goLogAnalyzer/pkg/utils"
	"math"
	"os"
	"strconv"
	"time"

	"github.com/sirupsen/logrus"
)

type Analyzer struct {
	*csvdb.CsvDB
	dataDir                  string
	logPath                  string
	logFormat                string
	timestampLayout          string
	useUtcTime               bool
	blockSize                int
	maxBlocks                int
	keepPeriod               int64
	unitSecs                 int64
	configTable              *csvdb.Table
	lastStatusTable          *csvdb.Table
	trans                    *trans
	fp                       *filepointer.FilePointer
	searchRegex, exludeRegex []string
	lastFileEpoch            int64
	lastFileRow              int
	rowID                    int
	readOnly                 bool
	linesProcessed           int
	termCountBorderRate      float64
	termCountBorder          int
	minMatchRate             float64
	keywords                 []string
	ignorewords              []string
	customLogGroups          []string
	separators               string
}

func NewAnalyzer(dataDir, logPath, logFormat, timestampLayout string, useUtcTime bool,
	searchRegex, exludeRegex []string,
	maxBlocks, blockSize int,
	keepPeriod int64, unitSecs int64,
	termCountBorderRate float64,
	termCountBorder int,
	minMatchRate float64,
	keywords, ignorewords, customLogGroups []string,
	separators string,
	readOnly bool) (*Analyzer, error) {
	a := new(Analyzer)
	a.dataDir = dataDir
	a.logPath = logPath
	a.logFormat = logFormat
	a.useUtcTime = useUtcTime
	a.keywords = keywords
	a.ignorewords = ignorewords
	a.timestampLayout = timestampLayout
	a.maxBlocks = maxBlocks
	a.blockSize = blockSize
	a.readOnly = readOnly
	a.searchRegex = searchRegex
	a.exludeRegex = exludeRegex

	// set defaults
	a.unitSecs = utils.CFreqDay
	a.keepPeriod = CDefaultKeepPeriod
	a.termCountBorder = CDefaultTermCountBorder
	a.termCountBorderRate = CDefaultTermCountBorderRate
	a.minMatchRate = CDefaultMinMatchRate
	a.separators = CDefaultSeparators

	// override passed params
	if unitSecs > 0 {
		a.unitSecs = unitSecs
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
	if minMatchRate > 0 {
		a.minMatchRate = minMatchRate
	}

	if separators != "" {
		a.separators = separators
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
	minMatchRate float64,
	customLogGroups []string,
	readOnly bool) (*Analyzer, error) {
	a := new(Analyzer)

	if dataDir == "" {
		return nil, utils.ErrorStack("no data to load")
	}
	if !utils.PathExist(fmt.Sprintf("%s/lastStatus.tbl.ini", dataDir)) {
		return nil, utils.ErrorStack("no data to load")
	}

	a.dataDir = dataDir
	a.logPath = logPath
	a.readOnly = readOnly

	// set defaults
	a.termCountBorder = CDefaultTermCountBorder
	a.termCountBorderRate = CDefaultTermCountBorderRate
	a.minMatchRate = CDefaultMinMatchRate

	if termCountBorder > 0 {
		a.termCountBorder = termCountBorder
	}
	if termCountBorderRate > 0 {
		a.termCountBorderRate = termCountBorderRate
	}
	if minMatchRate > 0 {
		a.minMatchRate = minMatchRate
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
		a.useUtcTime,
		a.maxBlocks, a.blockSize, a.unitSecs, a.keepPeriod,
		a.termCountBorderRate, a.termCountBorder, a.minMatchRate,
		a.searchRegex, a.exludeRegex,
		a.keywords, a.ignorewords, a.customLogGroups, a.separators, true, a.readOnly,
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
		a.lastFileEpoch = epoch
		a.rowID = rowNo
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
		&a.keepPeriod, &a.unitSecs,
		&a.termCountBorderRate,
		&a.termCountBorder,
		&a.minMatchRate,
		&a.timestampLayout,
		&a.useUtcTime,
		&a.separators,
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
		"unitSecs":            a.unitSecs,
		"termCountBorderRate": a.termCountBorderRate,
		"termCountBorder":     a.termCountBorder,
		"minMatchRate":        a.minMatchRate,
		"timestampLayout":     a.timestampLayout,
		"useUtcTime":          a.useUtcTime,
		"separators":          a.separators,
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
	a.trans.setCountBorder()

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

		if err := a.trans.lineToLogGroup(line, 1, a.fp.CurrFileEpoch()); err != nil {
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

	if !a.readOnly {
		if err := a._commit(false); err != nil {
			return err
		}
		logrus.Infof("processed %d lines", linesProcessed)
	}

	a.fp.Close()

	a.linesProcessed = linesProcessed

	return nil
}

func (a *Analyzer) OutputLogGroups(N int, outdir string, isHistory bool) error {
	if err := a.Feed(0); err != nil {
		return err
	}
	var groupIds []int64
	if N > 0 {
		groupIds = a.trans.getBiggestGroupIds(N)
	}

	if err := utils.EnsureDir(outdir); err != nil {
		return err
	}

	if isHistory {
		if err := a._outputLogGroupsHistory(outdir, groupIds); err != nil {
			return err
		}
	}
	return a._outputLogGroups(outdir, groupIds)
}

func (a *Analyzer) _outputLogGroups(outdir string, groupIds []int64) error {
	var writer *csv.Writer
	file, err := os.Create(fmt.Sprintf("%s/logGroups.csv", outdir))
	if err != nil {
		return fmt.Errorf("error creating file: %w", err)
	}
	defer file.Close()
	writer = csv.NewWriter(file)
	defer writer.Flush()
	lgs := a.trans.lgs.alllg
	writer.Write([]string{"groupId", "count", "text"})
	for _, groupId := range groupIds {
		lg := lgs[groupId]
		writer.Write([]string{fmt.Sprint(groupId), fmt.Sprint(lg.count), lg.displayString})
	}
	return nil
}

func (a *Analyzer) _outputLogGroupsHistory(outdir string, groupIds []int64) error {
	lgsh, err := a.trans.getLogGroupsHistory(groupIds)
	if err != nil {
		return err
	}
	var writer *csv.Writer
	if err := utils.EnsureDir(outdir); err != nil {
		return err
	}
	file, err := os.Create(fmt.Sprintf("%s/history.csv", outdir))
	if err != nil {
		return fmt.Errorf("error creating file: %w", err)
	}
	defer file.Close()
	writer = csv.NewWriter(file)
	defer writer.Flush()

	format := utils.GetDatetimeFormatFromUnitSecs(a.unitSecs)
	header := []string{"groupId"}
	for _, ep := range lgsh.timeline {
		header = append(header, time.Unix(ep, 0).Format(format))
	}
	writer.Write(header)
	for i, groupId := range lgsh.groupIds {
		row := []string{fmt.Sprint(groupId)}
		for _, cnt := range lgsh.counts[i] {
			row = append(row, strconv.Itoa(cnt))
		}
		writer.Write(row)
	}

	return nil
}