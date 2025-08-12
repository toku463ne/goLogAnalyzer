package logan

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"goLogAnalyzer/pkg/filepointer"
	"goLogAnalyzer/pkg/utils"
	"io/ioutil"
	"math"
	"os"
	"time"

	"github.com/sirupsen/logrus"
)

type AnalConfig struct {
	DataDir             string   `json:"data_dir"`
	LogPath             string   `json:"log_path"`
	LogFormat           string   `json:"log_format"`
	MsgFormats          []string `json:"msg_formats"`
	PatternKeyRegexes   []string `json:"pattern_key_regexes"`
	TimestampLayout     string   `json:"timestamp_layout"`
	UseUtcTime          bool     `json:"use_utc_time"`
	BlockSize           int      `json:"block_size"`
	MaxBlocks           int      `json:"max_blocks"`
	KeepPeriod          int64    `json:"keep_period"`
	UnitSecs            int64    `json:"unit_secs"`
	SearchRegex         []string `json:"search_regex"`
	ExludeRegex         []string `json:"exclude_regex"`
	TermCountBorderRate float64  `json:"term_count_border_rate"`
	TermCountBorder     int      `json:"term_count_border"`
	MinMatchRate        float64  `json:"min_match_rate"`
	Keywords            []string `json:"keywords"`
	KeyRegexes          []string `json:"key_regexes"`
	Ignorewords         []string `json:"ignorewords"`
	IgnoreRegexes       []string `json:"ignore_regexes"`
	CustomLogGroups     []string `json:"custom_log_groups"`
	Separators          string   `json:"separators"`
	IgnoreNumbers       bool     `json:"ignore_numbers"`
}

type analStatus struct {
	LastFileEpoch int64 `json:"last_file_epoch"`
	LastFileRow   int   `json:"last_file_row"`
	RowID         int   `json:"row_id"`
}

type historyInfo struct {
	Start    int64 `json:"start"`
	End      int64 `json:"end"`
	UnitSecs int64 `json:"unit_secs"`
}

type Analyzer struct {
	*AnalConfig
	*analStatus
	trans          *trans
	fp             *filepointer.FilePointer
	readOnly       bool
	testMode       bool
	linesProcessed int
}

// NewAnalyzer creates a new Analyzer instance with the provided configuration
func NewAnalyzer(conf *AnalConfig, lastFileEpoch int64, readOnly, testMode bool) (*Analyzer, error) {
	a := new(Analyzer)
	a.AnalConfig = new(AnalConfig)
	a.analStatus = new(analStatus)
	a.DataDir = conf.DataDir
	a.LogPath = conf.LogPath
	a.LogFormat = conf.LogFormat
	a.MsgFormats = conf.MsgFormats
	a.UseUtcTime = conf.UseUtcTime
	a.Keywords = conf.Keywords
	a.Ignorewords = conf.Ignorewords
	a.KeyRegexes = conf.KeyRegexes
	a.PatternKeyRegexes = conf.PatternKeyRegexes
	a.IgnoreRegexes = conf.IgnoreRegexes
	a.TimestampLayout = conf.TimestampLayout
	a.MaxBlocks = conf.MaxBlocks
	a.BlockSize = conf.BlockSize
	a.readOnly = readOnly
	a.SearchRegex = conf.SearchRegex
	a.ExludeRegex = conf.ExludeRegex
	a.testMode = testMode
	a.IgnoreNumbers = conf.IgnoreNumbers

	// set defaults
	a.UnitSecs = utils.CFreqDay
	a.KeepPeriod = CDefaultKeepPeriod
	a.TermCountBorder = CDefaultTermCountBorder
	a.TermCountBorderRate = CDefaultTermCountBorderRate
	a.MinMatchRate = CDefaultMinMatchRate
	a.Separators = CDefaultSeparators

	a.LastFileEpoch = lastFileEpoch
	a.LastFileRow = 0
	a.RowID = 0

	// override passed params
	if conf.UnitSecs > 0 {
		a.UnitSecs = conf.UnitSecs
	}
	if conf.KeepPeriod > 0 {
		a.KeepPeriod = conf.KeepPeriod
	}

	if conf.TermCountBorder > 0 {
		a.TermCountBorder = conf.TermCountBorder
	}
	if conf.TermCountBorderRate > 0 {
		a.TermCountBorderRate = conf.TermCountBorderRate
	}
	if conf.MinMatchRate > 0 {
		a.MinMatchRate = conf.MinMatchRate
	}

	if conf.Separators != "" {
		a.Separators = conf.Separators
	}

	// load or init data.
	// Some params will be replaced by params in the DB
	if err := a.open(); err != nil {
		return nil, err
	}

	// for some parameters, the args takes place
	if conf.LogPath != "" {
		a.LogPath = conf.LogPath
	}

	a.CustomLogGroups = conf.CustomLogGroups

	return a, nil
}

func LoadAnalyzer(dataDir, logPath string,
	termCountBorderRate float64,
	termCountBorder int,
	minMatchRate float64,
	customLogGroups []string,
	readOnly, _debug, testMode, ignoreNumbers bool) (*Analyzer, error) {
	a := new(Analyzer)
	a.AnalConfig = new(AnalConfig)
	a.analStatus = new(analStatus)

	a.testMode = testMode
	a.IgnoreNumbers = ignoreNumbers
	debug = _debug
	a.DataDir = dataDir
	a.LogPath = logPath

	if dataDir == "" {
		return nil, utils.ErrorStack("no data to load")
	}
	if !utils.PathExist(a._getConfigPath()) {
		return nil, utils.ErrorStack("no data to load")
	}

	a.readOnly = readOnly

	// set defaults
	a.TermCountBorder = CDefaultTermCountBorder
	a.TermCountBorderRate = CDefaultTermCountBorderRate
	a.MinMatchRate = CDefaultMinMatchRate

	if termCountBorder > 0 {
		a.TermCountBorder = termCountBorder
	}
	if termCountBorderRate > 0 {
		a.TermCountBorderRate = termCountBorderRate
	}
	if minMatchRate > 0 {
		a.MinMatchRate = minMatchRate
	}

	// load data.
	// Some params will be replaced by params in the DB
	if err := a.open(); err != nil {
		return nil, err
	}

	// for some parameters, the args takes place
	if logPath != "" {
		a.LogPath = logPath
	}

	needRebuild := false
	if termCountBorder > 0 && a.TermCountBorder != termCountBorder {
		a.TermCountBorder = termCountBorder
		needRebuild = true
	}
	if termCountBorderRate > 0 && a.TermCountBorderRate != termCountBorderRate {
		a.TermCountBorderRate = termCountBorderRate
		a.trans.setCountBorder()
		needRebuild = true
	}
	if minMatchRate > 0 && a.MinMatchRate != minMatchRate {
		a.MinMatchRate = minMatchRate
		needRebuild = true
	}
	if needRebuild {
		logrus.Info("rebuilding log groups")
		if err := a.rebuildTrans(); err != nil {
			return nil, err
		}
	}

	a.CustomLogGroups = customLogGroups

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
	if err := utils.RemoveDirectory(a.DataDir); err != nil {
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
	if err := a.trans.load(); err != nil {
		return err
	}
	return nil
}

func (a *Analyzer) open() error {
	if a.DataDir == "" {
		a.initBlocks()
		if err := a.init(); err != nil {
			return err
		}
	} else {
		if utils.PathExist(a.DataDir) {
			if err := a.loadConfig(); err != nil {
				return err
			}
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

			if err := a.saveLastStatus(); err != nil {
				return err
			}
			if err := a.saveConfig(); err != nil {
				return err
			}
		}
	}
	return nil
}

func (a *Analyzer) init() error {
	if a.DataDir != "" && !a.readOnly && a.testMode {
		if err := utils.EnsureDir(a.DataDir); err != nil {
			return err
		}
	}
	trans, err := newTrans(a.DataDir, a.LogFormat, a.TimestampLayout,
		a.UseUtcTime,
		a.MaxBlocks, a.BlockSize, a.UnitSecs, a.KeepPeriod,
		a.TermCountBorderRate, a.TermCountBorder, a.MinMatchRate,
		a.SearchRegex, a.ExludeRegex,
		a.Keywords, a.Ignorewords,
		a.KeyRegexes, a.IgnoreRegexes,
		a.MsgFormats,
		a.PatternKeyRegexes,
		a.CustomLogGroups, a.Separators, true,
		a.readOnly, a.testMode, a.IgnoreNumbers)
	if err != nil {
		return err
	}
	a.trans = trans
	return nil
}

func (a *Analyzer) _getLastStatusPath() string {
	return fmt.Sprintf("%s/status.json", a.DataDir)
}

func (a *Analyzer) saveLastStatus() error {
	if a.DataDir == "" || a.readOnly || a.testMode {
		return nil
	}
	if !utils.PathExist(a.DataDir) {
		return fmt.Errorf("%s does not exist", a.DataDir)
	}

	var epoch int64
	rowNo := 0
	if a.fp != nil {
		epoch = a.fp.CurrFileEpoch()
		rowNo = a.fp.Row()
		a.LastFileEpoch = epoch
		a.RowID = rowNo
		a.LastFileRow = a.fp.Row()
	} else {
		epoch = 0
		rowNo = 0
	}

	data, err := json.MarshalIndent(a.analStatus, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config to JSON: %w", err)
	}

	err = ioutil.WriteFile(a._getLastStatusPath(), data, 0644)
	if err != nil {
		return fmt.Errorf("failed to write JSON to file: %w", err)
	}

	return err
}

func (a *Analyzer) loadStatus() error {
	if a.testMode {
		return nil
	}

	if a.LastFileEpoch == 0 {
		data, err := ioutil.ReadFile(a._getLastStatusPath())
		if err != nil {
			return fmt.Errorf("failed to read JSON file: %w", err)
		}

		err = json.Unmarshal(data, a.analStatus)
		if err != nil {
			return fmt.Errorf("failed to unmarshal JSON: %w", err)
		}

	}

	return nil
}

func (a *Analyzer) _getConfigPath() string {
	return fmt.Sprintf("%s/config.json", a.DataDir)
}

func (a *Analyzer) saveConfig() error {
	if a.testMode {
		return nil
	}
	data, err := json.MarshalIndent(a.AnalConfig, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config to JSON: %w", err)
	}

	err = ioutil.WriteFile(a._getConfigPath(), data, 0644)
	if err != nil {
		return fmt.Errorf("failed to write JSON to file: %w", err)
	}

	return nil
}

func (a *Analyzer) loadConfig() error {
	if a.testMode {
		return nil
	}
	data, err := ioutil.ReadFile(a._getConfigPath())
	if err != nil {
		return fmt.Errorf("failed to read JSON file: %w", err)
	}

	err = json.Unmarshal(data, a.AnalConfig)
	if err != nil {
		return fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	return nil
}

func (a *Analyzer) initBlocks() {
	if a.MaxBlocks > 0 && a.BlockSize > 0 {
		if a.trans != nil {
			a.trans.setBlockSize(a.BlockSize)
			a.trans.setMaxBlocks(a.MaxBlocks)
		}
		return
	}

	if a.trans == nil || a.trans.maxCountByBlock == 0 {
		return
	}

	maxCountByBlock := a.trans.maxCountByBlock
	keepPeriod := a.KeepPeriod
	if keepPeriod == 0 {
		keepPeriod = 30
	}

	if a.BlockSize == 0 {
		if maxCountByBlock < 3000 {
			a.BlockSize = 10000
		} else if maxCountByBlock < 30000 {
			a.BlockSize = 100000
		} else if maxCountByBlock < 300000 {
			a.BlockSize = 100000
		} else {
			a.BlockSize = 1000000
		}
	}

	if a.MaxBlocks == 0 {
		n := int(math.Ceil(float64(a.trans.maxCountByBlock) / float64(a.BlockSize)))
		a.MaxBlocks = n * int(a.KeepPeriod)
	}
	a.trans.setBlockSize(a.BlockSize)
	a.trans.setMaxBlocks(a.MaxBlocks)

}

func (a *Analyzer) _commit(completed bool) error {
	if a.readOnly {
		return nil
	}
	if a.DataDir == "" {
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

	return nil
}

func (a *Analyzer) _initFilePointer() error {
	var err error
	if a.fp == nil || !a.fp.IsOpen() {
		a.fp, err = filepointer.NewFilePointer(a.LogPath, a.LastFileEpoch, a.LastFileRow)
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
		a.RowID++
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

func (a *Analyzer) OutputLogGroups(N int, outdir string,
	searchString, excludeString string,
	minLastUpdate int64, minCnt, maxCnt int,
	isHistory, asc bool) error {
	if err := a.Feed(0); err != nil {
		return err
	}

	//if len(a.trans.lgs.alllg) == 0 {
	//	return fmt.Errorf("no log groups found")
	//}

	allgroupIds := a.trans.getTopNGroupIds(len(a.trans.lgs.alllg), minLastUpdate, searchString, excludeString, minCnt, maxCnt, asc)

	var groupIds []int64
	if N > 0 {
		groupIds = a.trans.getTopNGroupIds(N, minLastUpdate, searchString, excludeString, minCnt, maxCnt, asc)
	} else {
		groupIds = allgroupIds
	}

	if outdir == "" {
		a._printLogGroups(groupIds)
		return nil
	}

	if err := utils.EnsureDir(outdir); err != nil {
		return err
	}

	if isHistory {
		if err := a._outputLogGroupsHistoryToCsv("history", outdir, allgroupIds, N); err != nil {
			return err
		}
	}
	return a._outputLogGroups("logGroups", outdir, groupIds)
}

func (a *Analyzer) _outputLogGroups(title, outdir string, groupIds []int64) error {
	var writer *csv.Writer
	file, err := os.Create(fmt.Sprintf("%s/%s.csv", outdir, title))
	if err != nil {
		return fmt.Errorf("error creating file: %w", err)
	}

	logrus.Infof("writing %s", file.Name())
	defer file.Close()
	writer = csv.NewWriter(file)
	defer writer.Flush()
	lgs := a.trans.lgs.alllg
	// header
	writer.Write([]string{"groupId", "count", "score", "text"})
	for _, groupId := range groupIds {
		lg := lgs[groupId]
		writer.Write([]string{fmt.Sprint(groupId), fmt.Sprint(lg.count),
			fmt.Sprintf("%.2f", lg.rareScore), lg.displayString})
	}
	writer.Flush()
	file.Close()

	// lastMessages
	file, err = os.Create(fmt.Sprintf("%s/%s_last.csv", outdir, title))
	if err != nil {
		return fmt.Errorf("error creating file: %w", err)
	}

	logrus.Infof("wrinting %s", file.Name())

	defer file.Close()
	writer = csv.NewWriter(file)
	defer writer.Flush()
	// header
	writer.Write([]string{"groupId", "lastUpdate", "count", "score", "text"})
	for _, groupId := range groupIds {
		lg := lgs[groupId]
		writer.Write([]string{fmt.Sprint(groupId), utils.EpochToString(lg.updated), fmt.Sprint(lg.count),
			fmt.Sprintf("%.3f", lg.rareScore), a.trans.lgs.lastMessages[groupId]})
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

func (a *Analyzer) _outputMetrics(title, outdir string, rows [][]string) error {
	if err := utils.EnsureDir(outdir); err != nil {
		return err
	}

	file, err := os.Create(fmt.Sprintf("%s/%s.csv", outdir, title))
	if err != nil {
		return fmt.Errorf("error creating file: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header to CSV
	if err := writer.Write([]string{"itemid", "clock", "value"}); err != nil {
		return fmt.Errorf("error writing header to CSV: %w", err)
	}

	// Write rows to CSV
	for _, row := range rows {
		if err := writer.Write(row); err != nil {
			return fmt.Errorf("error writing row to CSV: %w", err)
		}
	}
	logrus.Infof("Data written to CSV at %s", file.Name())

	return nil

}

func (a *Analyzer) _outputLogGroupsHistoryToCsv(title string, outdir string,
	groupIds []int64, topN int) error {
	lgsh, err := a.trans.getLogGroupsHistory(groupIds)
	if err != nil {
		return err
	}

	// Ensure output directory exists
	if err := utils.EnsureDir(outdir); err != nil {
		return err
	}

	//format := utils.GetDatetimeFormatFromUnitSecs(a.UnitSecs)
	// Prepare CSV rows
	rows := lgsh.buildRows(topN)

	// Write CSV file
	if err := a._outputMetrics(title, outdir, rows); err != nil {
		return err
	}

	// Write start, end and unitSecs to JSON file
	historyInfo := historyInfo{
		Start:    lgsh.timeline[0],
		End:      lgsh.timeline[len(lgsh.timeline)-1],
		UnitSecs: a.UnitSecs,
	}
	data, err := json.MarshalIndent(historyInfo, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal history info to JSON: %w", err)
	}

	historyInfoFile, err := os.Create(fmt.Sprintf("%s/%s_info.json", outdir, title))
	if err != nil {
		return fmt.Errorf("error creating file: %w", err)
	}
	defer historyInfoFile.Close()

	if _, err := historyInfoFile.Write(data); err != nil {
		return fmt.Errorf("error writing history info to JSON file: %w", err)
	}

	return nil
}

/*
In case some of below have changed since the saved config, rebuild trans with read only
a.TermCountBorder, a.MinMatchRate, a.SearchRegex, a.ExludeRegex,a.Keywords, a.Ignorewords, a.CustomLogGroups
*/
func (a *Analyzer) rebuildTrans() error {
	tr2, err := newTrans(a.DataDir, "", "", a.UseUtcTime, a.MaxBlocks, a.BlockSize, a.UnitSecs, a.KeepPeriod,
		0, a.TermCountBorder, a.MinMatchRate, a.SearchRegex, a.ExludeRegex,
		a.Keywords, a.Ignorewords,
		a.KeyRegexes, a.IgnoreRegexes,
		a.MsgFormats, a.PatternKeyRegexes,
		a.CustomLogGroups, a.Separators,
		true, true, a.testMode, a.IgnoreNumbers)
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

func (a *Analyzer) DetectPatterns(minCnt int) error {
	if err := a.Feed(0); err != nil {
		return err
	}

	if len(a.PatternKeyRegexes) == 0 {
		logrus.Warn("no pattern key regexes defined")
		return nil
	}

	a.trans.detectPaterns(minCnt)

	return nil
}

func (a *Analyzer) ParseLogLine(line string) {
	if _, err := a.trans.lineToLogGroup(line, 1, 0); err != nil {
		logrus.Errorf("%+v", err)
	}

	line, updated, _, err := a.trans.parseLine(line, 0)
	if err != nil {
		print(err)
		return
	}
	format := utils.GetDatetimeFormatFromUnitSecs(a.UnitSecs)
	dt := ""
	if updated > 0 {
		dt = time.Unix(updated, 0).Format(format)
	} else {
		dt = "PARSE ERROR"
	}
	println("the line parsed as:")
	println("timestamp: ", dt)
	println("message: ", line)

	if len(a.PatternKeyRegexes) > 0 {
		PatternKeyId, matched, err := a.trans.pk.findAndRegister(line)
		if err != nil {
			print(err)
			return
		}
		if matched {
			println("key group matched: ", PatternKeyId)
		} else {
			println("no key group matched")
		}
	}

}
