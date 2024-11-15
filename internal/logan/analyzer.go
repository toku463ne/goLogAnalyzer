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

type analConfig struct {
	dataDir                  string
	logPath                  string
	logFormat                string
	timestampLayout          string
	useUtcTime               bool
	blockSize                int
	maxBlocks                int
	keepPeriod               int64
	unitSecs                 int64
	searchRegex, exludeRegex []string
	termCountBorderRate      float64
	termCountBorder          int
	minMatchRate             float64
	keywords                 []string
	ignorewords              []string
	customLogGroups          []string
	separators               string
}

type Analyzer struct {
	*csvdb.CsvDB
	*analConfig
	configTable     *csvdb.Table
	lastStatusTable *csvdb.Table
	trans           *trans
	fp              *filepointer.FilePointer
	lastFileEpoch   int64
	lastFileRow     int
	rowID           int
	readOnly        bool
	linesProcessed  int
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
	readOnly, _debug bool) (*Analyzer, error) {
	debug = _debug
	a := new(Analyzer)
	a.analConfig = new(analConfig)
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
	readOnly, _debug bool) (*Analyzer, error) {
	a := new(Analyzer)
	a.analConfig = new(analConfig)
	debug = _debug

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

	needRebuild := false
	if termCountBorder > 0 && a.termCountBorder != termCountBorder {
		a.termCountBorder = termCountBorder
		needRebuild = true
	}
	if termCountBorderRate > 0 && a.termCountBorderRate != termCountBorderRate {
		a.termCountBorderRate = termCountBorderRate
		a.trans.setCountBorder()
		needRebuild = true
	}
	if minMatchRate > 0 && a.minMatchRate != minMatchRate {
		a.minMatchRate = minMatchRate
		needRebuild = true
	}
	if needRebuild {
		logrus.Info("rebuilding log groups")
		if err := a.rebuildTrans(); err != nil {
			return nil, err
		}
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
	logrus.Infof("starting terms registering")
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
	logrus.Infof("starting logGroups registering")
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

		if _, err := a.trans.lineToLogGroup(line, 1, a.fp.CurrFileEpoch()); err != nil {
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
		if linesProcessed > 0 {
			logrus.Infof("processed %d lines", linesProcessed)
		}
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

	if outdir == "" {
		a._printLogGroups(groupIds)
		return nil
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

	logrus.Infof("writing %s", file.Name())
	defer file.Close()
	writer = csv.NewWriter(file)
	defer writer.Flush()
	lgs := a.trans.lgs.alllg
	// header
	writer.Write([]string{"groupId", "count", "text"})
	for _, groupId := range groupIds {
		lg := lgs[groupId]
		writer.Write([]string{fmt.Sprint(groupId), fmt.Sprint(lg.count), lg.displayString})
	}
	writer.Flush()
	file.Close()

	// lastMessages
	file, err = os.Create(fmt.Sprintf("%s/lastMessages.csv", outdir))
	if err != nil {
		return fmt.Errorf("error creating file: %w", err)
	}

	logrus.Infof("wrinting %s", file.Name())

	defer file.Close()
	writer = csv.NewWriter(file)
	defer writer.Flush()
	// header
	writer.Write([]string{"groupId", "text"})
	for _, groupId := range groupIds {
		writer.Write([]string{fmt.Sprint(groupId), a.trans.lgs.lastMessages[groupId]})
	}
	writer.Flush()
	file.Close()

	return nil
}

func (a *Analyzer) _printLogGroups(groupIds []int64) error {
	lgs := a.trans.lgs.alllg
	// Print header for log groups
	fmt.Println("Log Groups")
	fmt.Println("==========")
	fmt.Printf("%-10s %-10s %-s\n", "Group ID", "Count", "Text")
	for _, groupId := range groupIds {
		lg := lgs[groupId]
		fmt.Printf("%-10d %-10d %s\n", groupId, lg.count, lg.displayString)
	}
	fmt.Println()

	return nil
}

// Function to generate cicle pattern as a string of "1" and "0" for each row
func (a *Analyzer) _generateCiclePattern(row []int) string {
	pattern := ""
	for _, value := range row {
		if value != 0 {
			pattern += "1"
		} else {
			pattern += "0"
		}
	}
	return pattern
}

func (a *Analyzer) _outputLogGroupsHistory(outdir string, groupIds []int64) error {
	lgsh, err := a.trans.getLogGroupsHistory(groupIds)
	if err != nil {
		return err
	}

	// output simple history
	var writer *csv.Writer
	if err := utils.EnsureDir(outdir); err != nil {
		return err
	}
	file, err := os.Create(fmt.Sprintf("%s/history.csv", outdir))
	if err != nil {
		return fmt.Errorf("error creating file: %w", err)
	}

	logrus.Infof("wrinting %s", file.Name())
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
			if cnt > 0 {
				row = append(row, strconv.Itoa(cnt))
			} else {
				row = append(row, "")
			}
		}
		writer.Write(row)
	}
	writer.Flush()
	file.Close()

	// output history grouped by apparance cicle
	cicleGroups := make(map[string][]int)
	cicleGroupIDs := make(map[string]int64)
	cicleIDs := make(map[int64]string)

	for i, groupId := range lgsh.groupIds {
		row := lgsh.counts[i]
		cicle := a._generateCiclePattern(row)
		if _, exists := cicleGroups[cicle]; !exists {
			cicleGroups[cicle] = make([]int, len(lgsh.timeline))
			if _, ok := cicleGroupIDs[cicle]; !ok {
				cicleGroupIDs[cicle] = groupId
				cicleIDs[groupId] = cicle
			}
		}
		for j, cnt := range lgsh.counts[i] {
			cicleGroups[cicle][j] += cnt
		}
	}
	file, err = os.Create(fmt.Sprintf("%s/history_sum.csv", outdir))
	if err != nil {
		return fmt.Errorf("error creating file: %w", err)
	}
	defer file.Close()
	writer = csv.NewWriter(file)
	defer writer.Flush()
	writer.Write(header)
	for _, groupId := range groupIds {
		row := []string{fmt.Sprint(groupId)}
		if cicle, ok := cicleIDs[groupId]; ok {
			for _, sum := range cicleGroups[cicle] {
				if sum > 0 {
					row = append(row, strconv.Itoa(sum))
				} else {
					row = append(row, "")
				}
			}
			writer.Write(row)
		}
	}
	writer.Flush()
	file.Close()

	return nil
}

/*
In case some of below have changed since the saved config, rebuild trans with read only
a.termCountBorder, a.minMatchRate, a.searchRegex, a.exludeRegex,a.keywords, a.ignorewords, a.customLogGroups
*/
func (a *Analyzer) rebuildTrans() error {
	tr2, err := newTrans(a.dataDir, "", "", a.useUtcTime, a.maxBlocks, a.blockSize, a.unitSecs, a.keepPeriod,
		0, a.termCountBorder, a.minMatchRate, a.searchRegex, a.exludeRegex,
		a.keywords, a.ignorewords, a.customLogGroups, a.separators, true, true)
	if err != nil {
		return err
	}
	tr2.te = a.trans.te
	for _, lg := range a.trans.lgs.alllg {
		//tokens, displayString, err := tr2.toTokens(lg.displayString, 0, true, true, true)
		//if err != nil {
		//	return err
		//}
		//tr2.lgs.registerLogTree(tokens, lg.count, displayString, lg.created, lg.created, true, -1, -1)
		groupId, err := tr2.lineToLogGroup(lg.displayString, lg.count, lg.updated)
		if err != nil {
			return err
		}
		if lg2, ok := tr2.lgs.alllg[groupId]; ok {
			if lg2.created == 0 || lg2.created < lg.created {
				lg2.created = lg.created
			}
		}

	}
	tr2.lgs.orgDisplayStrings = a.trans.lgs.displayStrings
	a.trans = tr2
	return nil
}

func (a *Analyzer) ParseLogLine(line string) {
	if _, err := a.trans.lineToLogGroup(line, 1, 0); err != nil {
		logrus.Errorf("%+v", err)
	}

	line, updated, _ := a.trans.parseLine(line, 0)
	format := utils.GetDatetimeFormatFromUnitSecs(a.unitSecs)
	dt := ""
	if updated > 0 {
		dt = time.Unix(updated, 0).Format(format)
	} else {
		dt = "PARSE ERROR"
	}
	println("the line parsed as:")
	println("timestamp: ", dt)
	println("message: ", line)

}
