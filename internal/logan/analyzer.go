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
	"strconv"
	"time"

	"github.com/sirupsen/logrus"
)

type analConfig struct {
	DataDir             string   `json:"data_dir"`
	LogPath             string   `json:"log_path"`
	LogFormat           string   `json:"log_format"`
	MsgFormats          []string `json:"msg_formats"`
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
	KeyRegexes          []string `json:"keywords"`
	Ignorewords         []string `json:"ignorewords"`
	IgnoreRegexes       []string `json:"ignorewords"`
	CustomLogGroups     []string `json:"custom_log_groups"`
	Separators          string   `json:"separators"`
	IgnoreNumbers       bool     `json:"ignore_numbers"`
}

type analStatus struct {
	LastFileEpoch int64 `json:"last_file_epoch"`
	LastFileRow   int   `json:"last_file_row"`
	RowID         int   `json:"row_id"`
}

type Analyzer struct {
	*analConfig
	*analStatus
	trans          *trans
	fp             *filepointer.FilePointer
	readOnly       bool
	testMode       bool
	linesProcessed int
}

func NewAnalyzer(dataDir, logPath, logFormat, timestampLayout string, useUtcTime bool,
	searchRegex, exludeRegex []string,
	maxBlocks, blockSize int,
	keepPeriod int64, unitSecs int64,
	termCountBorderRate float64,
	termCountBorder int,
	minMatchRate float64,
	keywords, ignorewords,
	keyreRexes, ignoreRegexes,
	msgFormats []string,
	customLogGroups []string,
	separators string,
	readOnly, _debug, testMode, ignoreNumbers bool) (*Analyzer, error) {
	debug = _debug
	a := new(Analyzer)
	a.analConfig = new(analConfig)
	a.analStatus = new(analStatus)
	a.DataDir = dataDir
	a.LogPath = logPath
	a.LogFormat = logFormat
	a.MsgFormats = msgFormats
	a.UseUtcTime = useUtcTime
	a.Keywords = keywords
	a.Ignorewords = ignorewords
	a.KeyRegexes = keyreRexes
	a.IgnoreRegexes = ignoreRegexes
	a.TimestampLayout = timestampLayout
	a.MaxBlocks = maxBlocks
	a.BlockSize = blockSize
	a.readOnly = readOnly
	a.SearchRegex = searchRegex
	a.ExludeRegex = exludeRegex
	a.testMode = testMode
	a.IgnoreNumbers = ignoreNumbers

	// set defaults
	a.UnitSecs = utils.CFreqDay
	a.KeepPeriod = CDefaultKeepPeriod
	a.TermCountBorder = CDefaultTermCountBorder
	a.TermCountBorderRate = CDefaultTermCountBorderRate
	a.MinMatchRate = CDefaultMinMatchRate
	a.Separators = CDefaultSeparators

	a.LastFileEpoch = 0
	a.LastFileRow = 0
	a.RowID = 0

	// override passed params
	if unitSecs > 0 {
		a.UnitSecs = unitSecs
	}
	if keepPeriod > 0 {
		a.KeepPeriod = keepPeriod
	}

	if termCountBorder > 0 {
		a.TermCountBorder = termCountBorder
	}
	if termCountBorderRate > 0 {
		a.TermCountBorderRate = termCountBorderRate
	}
	if minMatchRate > 0 {
		a.MinMatchRate = minMatchRate
	}

	if separators != "" {
		a.Separators = separators
	}

	// load or init data.
	// Some params will be replaced by params in the DB
	if err := a.open(); err != nil {
		return nil, err
	}

	// for some parameters, the args takes place
	if logPath != "" {
		a.LogPath = logPath
	}

	a.CustomLogGroups = customLogGroups

	return a, nil
}

func LoadAnalyzer(dataDir, logPath string,
	termCountBorderRate float64,
	termCountBorder int,
	minMatchRate float64,
	customLogGroups []string,
	readOnly, _debug, testMode, ignoreNumbers bool) (*Analyzer, error) {
	a := new(Analyzer)
	a.analConfig = new(analConfig)
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
	data, err := json.MarshalIndent(a.analConfig, "", "  ")
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

	err = json.Unmarshal(data, a.analConfig)
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

func (a *Analyzer) Anomaly(N int, outdir string,
	searchString, excludeString string,
	minLastUpdate int64,
	maxRareCount, minFreqCount int,
	stdThreshold, minOccurrences float64) error {

	if stdThreshold == 0 {
		stdThreshold = CDefaultStdThreshold
	}
	if minOccurrences == 0 {
		minOccurrences = CDefaultMinOccurrences
	}

	if err := a.Feed(0); err != nil {
		return err
	}
	var rareGroupIds []int64
	if N > 0 {
		rareGroupIds = a.trans.getTopNGroupIds(N, minLastUpdate, searchString, excludeString, 0, maxRareCount, true)
	} else {
		rareGroupIds = a.trans.getTopNGroupIds(len(a.trans.lgs.alllg), minLastUpdate, searchString, excludeString, 0, maxRareCount, true)
	}

	if err := utils.EnsureDir(outdir); err != nil {
		return err
	}

	// output rare logs
	if err := a._outputLogGroups("rareGroups", outdir, rareGroupIds); err != nil {
		return err
	}

	var freqGroupIds []int64
	if N > 0 {
		freqGroupIds = a.trans.getTopNGroupIds(N, minLastUpdate,
			searchString, excludeString, minFreqCount, 0, false)
	} else {
		freqGroupIds = a.trans.getTopNGroupIds(len(a.trans.lgs.alllg), minLastUpdate,
			searchString, excludeString, minFreqCount, 0, false)
	}

	lgsh, err := a.trans.getLogGroupsHistory(freqGroupIds)
	if err != nil {
		return err
	}

	anomalGroupIds := make([]int64, 0)
	for _, groupId := range freqGroupIds {
		epochs := lgsh.detectAnomaly(groupId, stdThreshold, minOccurrences, minLastUpdate)
		if len(epochs) > 0 {
			anomalGroupIds = append(anomalGroupIds, groupId)
		}
	}
	// output history with anomly
	if err := a._outputLogGroupsHistoryToCsv("anomaly_history", outdir, anomalGroupIds); err != nil {
		return err
	}
	if err := a._outputLogGroups("anomaly", outdir, anomalGroupIds); err != nil {
		return err
	}

	return nil
}

func (a *Analyzer) OutputLogGroups(N int, outdir string,
	searchString, excludeString string,
	minLastUpdate int64, minCnt, maxCnt int,
	isHistory, asc bool) error {
	if err := a.Feed(0); err != nil {
		return err
	}
	var groupIds []int64
	if N > 0 {
		groupIds = a.trans.getTopNGroupIds(N, minLastUpdate, searchString, excludeString, minCnt, maxCnt, asc)
	} else {
		groupIds = a.trans.getTopNGroupIds(len(a.trans.lgs.alllg), minLastUpdate, searchString, excludeString, minCnt, maxCnt, asc)
	}

	if outdir == "" {
		a._printLogGroups(groupIds)
		return nil
	}

	if err := utils.EnsureDir(outdir); err != nil {
		return err
	}

	if isHistory {
		if err := a._outputLogGroupsHistoryToCsv("history", outdir, groupIds); err != nil {
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

func (a *Analyzer) _outputLogGroupsHistoryToCsv(title string, outdir string, groupIds []int64) error {
	lgsh, err := a.trans.getLogGroupsHistory(groupIds)
	if err != nil {
		return err
	}

	// Ensure output directory exists
	if err := utils.EnsureDir(outdir); err != nil {
		return err
	}

	// Prepare CSV rows
	//format := utils.GetDatetimeFormatFromUnitSecs(a.UnitSecs)
	var rows [][]string
	rows = append(rows, []string{"time", "metric", "value"}) // Add CSV header

	// Build rows from log group history
	for i, groupId := range lgsh.groupIds {
		for j, timestamp := range lgsh.timeline {
			count := lgsh.counts[i][j]
			// Add rows only for non-zero counts
			if count > 0 {
				rows = append(rows, []string{
					fmt.Sprint(timestamp), // time
					fmt.Sprint(groupId),   // metric
					strconv.Itoa(count),   // value
				})
			}
		}
	}

	// Write CSV file
	filePath := fmt.Sprintf("%s/%s.csv", outdir, title)
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("error creating file: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write rows to CSV
	for _, row := range rows {
		if err := writer.Write(row); err != nil {
			return fmt.Errorf("error writing row to CSV: %w", err)
		}
	}

	logrus.Infof("Data written to CSV in the format 'time,metric,value' at %s", filePath)
	return nil
}

func (a *Analyzer) OLD_outputLogGroupsHistoryToCsv(title string, outdir string, groupIds []int64) error {
	lgsh, err := a.trans.getLogGroupsHistory(groupIds)
	if err != nil {
		return err
	}

	// output simple history
	var writer *csv.Writer
	if err := utils.EnsureDir(outdir); err != nil {
		return err
	}
	file, err := os.Create(fmt.Sprintf("%s/%s.csv", outdir, title))
	if err != nil {
		return fmt.Errorf("error creating file: %w", err)
	}

	logrus.Infof("wrinting %s", file.Name())
	defer file.Close()
	writer = csv.NewWriter(file)
	defer writer.Flush()

	format := utils.GetDatetimeFormatFromUnitSecs(a.UnitSecs)
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
a.TermCountBorder, a.MinMatchRate, a.SearchRegex, a.ExludeRegex,a.Keywords, a.Ignorewords, a.CustomLogGroups
*/
func (a *Analyzer) rebuildTrans() error {
	tr2, err := newTrans(a.DataDir, "", "", a.UseUtcTime, a.MaxBlocks, a.BlockSize, a.UnitSecs, a.KeepPeriod,
		0, a.TermCountBorder, a.MinMatchRate, a.SearchRegex, a.ExludeRegex,
		a.Keywords, a.Ignorewords,
		a.KeyRegexes, a.IgnoreRegexes,
		a.MsgFormats,
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

}
