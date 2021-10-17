package analyzer

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/pkg/errors"
	csvdb "github.com/toku463ne/goLogAnalyzer/analyzer/csvdb"
)

func newRarityAnalyzer(rootDir string) *rarityAnalyzer {
	a := new(rarityAnalyzer)
	a.rootDir = rootDir
	return a
}

func (a *rarityAnalyzer) clean() error {
	if pathExist(a.rootDir) {
		return os.RemoveAll(a.rootDir)
	}
	return nil
}

func (a *rarityAnalyzer) init(logPathRegex, filterStr, xFilterStr string,
	minGapToRecord float64, maxBlocks, maxItemBlocks, linesInBlock, nTopRecordsCount int,
	datetimeStartPos int, datetimeLayout string, scoreStyle int) error {
	a.logPathRegex = logPathRegex
	a.filterRe = getRegex(filterStr)
	a.xFilterRe = getRegex(xFilterStr)
	a.datetimeStartPos = datetimeStartPos
	a.datetimeLayout = datetimeLayout
	a.scoreStyle = scoreStyle
	a.nTopRecordsCount = nTopRecordsCount

	if a.rootDir != "" {
		if err := ensureDir(a.rootDir); err != nil {
			return err
		}
	}
	InitLog(a.rootDir)

	a.minGapToRecord = minGapToRecord

	if maxBlocks > 0 {
		a.maxBlocks = maxBlocks
	}
	if maxItemBlocks > 0 {
		a.maxItemBlocks = maxItemBlocks
	}

	if linesInBlock > 0 {
		a.linesInBlock = linesInBlock
	}

	if err := a.openObjs(); err != nil {
		return err
	}

	if a.rootDir != "" {
		if err := a.prepareDB(); err != nil {
			return err
		}

		if err := a.saveConfig(); err != nil {
			return err
		}
		if err := a.saveLastStatus(); err != nil {
			return err
		}
		if err := a.saveObjs(); err != nil {
			return err
		}
	}

	return nil
}

func (a *rarityAnalyzer) open(logPathRegex, filterStr, xFilterStr string,
	minGapToRecord float64, maxBlocks, maxItemBlocks, linesInBlock, nTopRecordsCount int,
	datetimeStartPos int, datetimeLayout string, scoreStyle int) error {
	if pathExist(a.rootDir) {
		if err := a.load(); err != nil {
			return err
		}
	} else {
		if err := a.init(logPathRegex, filterStr, xFilterStr,
			minGapToRecord, maxBlocks, maxItemBlocks, linesInBlock, nTopRecordsCount,
			datetimeStartPos, datetimeLayout, scoreStyle); err != nil {
			return err
		}
	}
	return nil
}

func (a *rarityAnalyzer) prepareDB() error {
	d, err := csvdb.NewCsvDB(a.rootDir)
	if err != nil {
		return err
	}
	ct, err := d.CreateTableIfNotExists("config", tableDefs["config"], false, 1)
	if err != nil {
		return err
	}
	a.configTable = ct

	ls, err := d.CreateTableIfNotExists("lastStatus", tableDefs["lastStatus"], false, 1)
	if err != nil {
		return err
	}
	a.lastStatusTable = ls

	a.CsvDB = d
	return nil
}

func (a *rarityAnalyzer) load() error {
	InitLog(a.rootDir)
	log.Printf("loading data from %s", a.rootDir)

	if err := a.prepareDB(); err != nil {
		return err
	}

	filterReStr := ""
	xFilterReStr := ""
	if err := a.configTable.Select1Row(nil,
		[]string{"logPathRegex", "linesInBlock", "maxBlocks",
			"maxItemBlocks", "filterRe", "xFilterRe", "minGapToRecord",
			"datetimeStartPos", "datetimeLayout", "scoreStyle"},
		&a.logPathRegex, &a.linesInBlock, &a.maxBlocks, &a.maxItemBlocks,
		&filterReStr, &xFilterReStr, &a.minGapToRecord,
		&a.datetimeStartPos, &a.datetimeLayout, &a.scoreStyle); err != nil {
		return err
	}

	a.filterRe = getRegex(filterReStr)
	a.xFilterRe = getRegex(xFilterReStr)

	if err := a.lastStatusTable.Select1Row(nil,
		[]string{"lastRowID", "lastFileEpoch", "lastFileRow"},
		&a.rowID, &a.lastFileEpoch, &a.lastFileRow); err != nil {
		return err
	}

	if err := a.openObjs(); err != nil {
		return err
	}
	if err := a.loadObjs(); err != nil {
		return err
	}

	return nil
}

func (a *rarityAnalyzer) openObjs() error {
	trans, err := newTrans(a.rootDir, a.maxItemBlocks, a.linesInBlock,
		a.datetimeStartPos, a.datetimeLayout, a.scoreStyle)
	if err != nil {
		return err
	}
	a.trans = trans
	a.nTopRareLogs = newNTopRecords(a.nTopRecordsCount, 0.0, trans, true)

	stats, err := newStats(a.rootDir, a.maxBlocks, a.linesInBlock)
	if err != nil {
		return err
	}
	a.stats = stats

	logRecs, err := newLogRecords(a.rootDir, a.maxBlocks, a.linesInBlock)
	if err != nil {
		return err
	}
	a.logRecs = logRecs
	return nil
}

func (a *rarityAnalyzer) loadObjs() error {
	if err := a.trans.load(); err != nil {
		return err
	}
	if err := a.stats.load(false); err != nil {
		return err
	}
	if err := a.logRecs.load(); err != nil {
		return err
	}
	if err := a.nTopRareLogs.load(a.rootDir, a.rowID, a.maxBlocks*a.linesInBlock, false); err != nil {
		return err
	}
	return nil
}

func (a *rarityAnalyzer) saveObjs() error {
	if err := a.trans.commit(false); err != nil {
		return err
	}
	if err := a.stats.commit(false); err != nil {
		return err
	}
	if err := a.logRecs.commit(false); err != nil {
		return err
	}
	if err := a.nTopRareLogs.save(a.rootDir); err != nil {
		return err
	}
	return nil
}

func (a *rarityAnalyzer) saveConfig() error {
	filterReStr := re2str(a.filterRe)
	xFilterReStr := re2str(a.xFilterRe)

	if err := a.configTable.Upsert(nil, map[string]interface{}{
		"logPathRegex":     a.logPathRegex,
		"linesInBlock":     a.linesInBlock,
		"maxBlocks":        a.maxBlocks,
		"maxItemBlocks":    a.maxItemBlocks,
		"filterRe":         filterReStr,
		"xFilterRe":        xFilterReStr,
		"minGapToRecord":   a.minGapToRecord,
		"datetimeStartPos": a.datetimeStartPos,
		"datetimeLayout":   a.datetimeLayout,
		"scoreStyle":       a.scoreStyle,
	}); err != nil {
		return err
	}
	return nil
}

func (a *rarityAnalyzer) saveLastStatus() error {
	var epoch int64
	rowNo := 0
	if a.fp != nil {
		epoch = a.fp.currFileEpoch()
		rowNo = a.fp.row()
	} else {
		epoch = 0
		rowNo = 0
	}

	if a.rootDir == "" {
		return nil
	}

	if err := a.lastStatusTable.Upsert(nil, map[string]interface{}{
		"lastRowID":     a.rowID,
		"lastFileEpoch": epoch,
		"lastFileRow":   rowNo,
	}); err != nil {
		return err
	}
	return nil
}

func (a *rarityAnalyzer) close() {
	if a.fp != nil {
		a.fp.close()
	}
	if a.stats != nil {
		a.stats.close()
	}
	if a.logRecs != nil {
		a.logRecs = nil
	}
	if a.trans != nil {
		a.trans.close()
	}
}

func (a *rarityAnalyzer) commit(completed bool) error {
	if a.rootDir == "" {
		return nil
	}
	if err := a.trans.commit(completed); err != nil {
		return err
	}
	if err := a.stats.commit(completed); err != nil {
		return err
	}
	if err := a.logRecs.commit(completed); err != nil {
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

func (a *rarityAnalyzer) analyze(targetLinesCnt int) (int, error) {
	linesProcessed := 0

	if a.fp == nil || !a.fp.isOpen() {
		a.fp = newFilePointer(a.logPathRegex, a.lastFileEpoch, a.lastFileRow)
		if err := a.fp.open(); err != nil {
			return 0, err
		}
	}
	var lastEpoch int64
	for a.fp.next() {

		te := a.fp.text()
		if te == "" {
			continue
		}

		timeTran, tran, dt, err := a.trans.tokenizeLine(te, a.filterRe, a.xFilterRe, true)
		if err != nil {
			return linesProcessed, err
		}
		lineEpoch := dt.Unix()
		tran = append(tran, timeTran...)
		if len(tran) == 0 {
			continue
		}
		if lineEpoch > 0 {
			lastEpoch = lineEpoch
		} else {
			lastEpoch = a.fp.currFileEpoch()
		}
		a.trans.items.lastEpoch = lastEpoch
		a.logRecs.lastEpoch = lastEpoch

		score := a.trans.calcScore(tran)
		err = a.stats.registerScore(score, lastEpoch)
		if err != nil {
			return linesProcessed, err
		}
		if a.stats.lastGap >= a.minGapToRecord {
			if err := a.logRecs.insert(a.rowID, score, te, lastEpoch); err != nil {
				return linesProcessed, err
			}
		}
		if a.rootDir == "" {
			a.nTopRareLogs.register(a.rowID, score, te, false)
		}

		if a.fp.isEOF && !a.fp.isLastFile() {
			if err := a.saveLastStatus(); err != nil {
				return linesProcessed, err
			}
		}

		if linesProcessed > 0 && linesProcessed%cLogPerLines == 0 {
			log.Printf("processed %d lines", linesProcessed)
		}
		linesProcessed++

		a.rowID++
		if targetLinesCnt > 0 && linesProcessed >= targetLinesCnt {
			break
		}
	}
	if err := a.commit(false); err != nil {
		return linesProcessed, err
	}
	return linesProcessed, nil
}

func (a *rarityAnalyzer) scanAndGetNTop(recordsToShow int, startEpoch, endEpoch int64,
	filterReStr, xFilterReStr string,
	minScore float64, maxScore float64) (*nTopRecords, error) {
	var conditionCheckFunc func(v []string) bool
	var statusLastEpoch int64

	lastEpochIdx := getColIdx("circuitDBStatus", "lastEpoch")
	if startEpoch > 0 {
		conditionCheckFunc = func(v []string) bool {
			lastEpoch, _ := strconv.ParseInt(v[lastEpochIdx], 10, 64)
			if endEpoch > 0 && endEpoch > startEpoch {
				return lastEpoch >= startEpoch && lastEpoch <= endEpoch
			} else {
				return lastEpoch >= startEpoch
			}
		}
	} else {
		conditionCheckFunc = nil
	}

	rows, err := a.logRecs.statusTable.SelectRows(conditionCheckFunc,
		[]string{"blockNo", "lastEpoch"})
	if err != nil {
		return nil, errors.WithStack(err)
	}

	blockNos := make([]int, 0)
	minEpoch := int64(-1)
	maxEpoch := int64(0)
	for rows.Next() {
		blockNo := 0
		if err := rows.Scan(&blockNo, &statusLastEpoch); err != nil {
			return nil, err
		}
		if minEpoch < 0 || statusLastEpoch < minEpoch {
			minEpoch = statusLastEpoch
		}
		if statusLastEpoch > maxEpoch {
			maxEpoch = statusLastEpoch
		}
		blockNos = append(blockNos, blockNo)
	}
	startEpoch = minEpoch
	endEpoch = maxEpoch

	lastEpochIdx = getColIdx("logRecords", "epoch")
	r, err := a.logRecs.selectRows(conditionCheckFunc,
		blockNos, []string{"rowID", "score", "record", "epoch"})
	if err != nil {
		return nil, err
	}

	ntopUniq := newNTopRecords(recordsToShow, minScore, a.trans, true)

	filterRe := getRegex(filterReStr)
	xFilterRe := getRegex(xFilterReStr)
	var rowID int64
	var score float64
	var record string
	var lastEpoch int64
	for r.next() {
		if err := r.scan(&rowID, &score, &record, &lastEpoch); err != nil {
			return nil, err
		}
		if filterRe != nil && !filterRe.Match([]byte(record)) {
			continue
		}
		if xFilterRe != nil && xFilterRe.Match([]byte(record)) {
			continue
		}
		if minScore > 0 && score < minScore {
			continue
		}
		if maxScore > 0 && score > maxScore {
			continue
		}

		ntopUniq.register(rowID, score, record, false)
	}

	return ntopUniq, nil
}

func (a *rarityAnalyzer) getRarStatsString(rootDir string, histSize int) (string, error) {
	var err error
	out, _, err := a.stats.getCountPerStatsString(0)
	if err != nil {
		return "", err
	}
	//out += fmt.Sprintf("score border %f\n", border)
	//println("")
	if a.rootDir == "" {
		out += "statistics\n"
		out += fmt.Sprintf("average= %f\n", a.stats.lastAverage)
		out += fmt.Sprintf("std=     %f\n", a.stats.lastStd)
		out += fmt.Sprintf("max=     %f\n", a.stats.scoreMax)
		out += "\n"
		return "", nil
	}

	if a.rootDir != "" {
		if out2, err := a.stats.getRecentStatsString(histSize); err != nil {
			return "", err
		} else {
			out += out2
		}
	}
	return out, nil
}

func (a *rarityAnalyzer) getNTop(recordsToShow int, startEpoch, endEpoch int64,
	filterReStr, xFilterReStr string,
	minScore float64, maxScore float64) (*nTopRecords, error) {
	var err error
	var ntop *nTopRecords
	if a.rootDir == "" {
		ntop = a.nTopRareLogs
	} else {
		ntop, err = a.scanAndGetNTop(recordsToShow,
			startEpoch, endEpoch, filterReStr, xFilterReStr, minScore, maxScore)
		if err != nil {
			return nil, err
		}
	}
	return ntop, nil
}

func (a *rarityAnalyzer) getNTopString(msg string,
	recordsToShow int, outFormat string, ntop *nTopRecords) (string, float64, error) {

	switch outFormat {
	case cFormatText:
		return ntop.nTop2string(msg, recordsToShow)
	case cFormatHtml:
		return ntop.nTop2html(msg, recordsToShow)
	}
	return "", -1, nil
}

func (a *rarityAnalyzer) getNTopDiffString(msg string,
	recordsToShow int, outFormat string, ntop *nTopRecords) (string, float64, error) {

	switch outFormat {
	case cFormatText:
		return ntop.nTop2string(msg, recordsToShow)
	case cFormatHtml:
		return ntop.nTop2html(msg, recordsToShow)
	}
	return "", -1, nil
}
