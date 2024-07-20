package analyzer

import (
	"fmt"
	csvdb "goLogAnalyzer/csvdb"
	"log"
	"math"
	"strconv"

	"github.com/pkg/errors"
)

func newRarityAnalyzer(conf *AnalConf) (*rarityAnalyzer, error) {
	a := new(rarityAnalyzer)
	a.AnalConf = conf
	if err := a.open(); err != nil {
		return nil, err
	}
	return a, nil
}

func (a *rarityAnalyzer) open() error {
	if a.RootDir == "" {
		if err := a.initBlocks(); err != nil {
			return err
		}
		if err := a.init(); err != nil {
			return err
		}
	} else {
		if PathExist(a.RootDir) {
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
			if err := a.initBlocks(); err != nil {
				return err
			}
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
		}
		if IsDebug {
			msg := "alanyzer.open(): "
			msg += fmt.Sprintf("rootDir=%s LogPathRegex=%s ScoreStyle=%d ScoreNSize=%d ignoreCount=%d",
				a.RootDir, a.LogPathRegex, a.ScoreStyle, a.ScoreNSize, a.ignoreCount)
			ShowDebug(msg)
		}
	}

	return nil
}

func (a *rarityAnalyzer) loadStatus() error {
	if a.RootDir != "" {
		if err := a.prepareDB(); err != nil {
			return err
		}
	}
	filterReStr := ""
	xFilterReStr := ""
	if err := a.configTable.Select1Row(nil,
		[]string{"logPathRegex", "blockSize", "maxBlocks",
			"maxItemBlocks", "filterRe", "xFilterRe", "minGapToRecord",
			"datetimeStartPos", "datetimeLayout", "scoreStyle"},
		&a.LogPathRegex, &a.BlockSize, &a.MaxBlocks, &a.MaxItemBlocks,
		&filterReStr, &xFilterReStr, &a.MinGapToRecord,
		&a.DatetimeStartPos, &a.DatetimeLayout, &a.ScoreStyle); err != nil {
		return err
	}

	a.filterRe = getRegex(filterReStr)
	a.xFilterRe = getRegex(xFilterReStr)

	if err := a.lastStatusTable.Select1Row(nil,
		[]string{"lastRowID", "lastFileEpoch", "lastFileRow"},
		&a.rowID, &a.lastFileEpoch, &a.lastFileRow); err != nil {
		if err.Error() != cErrPathNotExists {
			return err
		}
	}
	return nil
}

func (a *rarityAnalyzer) load() error {

	if err := a.trans.load(); err != nil {
		return err
	}
	if err := a.stats.load(false); err != nil {
		return err
	}
	if err := a.logRecs.load(); err != nil {
		return err
	}
	return nil
}

func (a *rarityAnalyzer) init() error {
	if a.RootDir != "" {
		if err := ensureDir(a.RootDir); err != nil {
			return err
		}
	}

	trans, err := newTrans(a.RootDir, a.MaxItemBlocks, a.BlockSize,
		a.DatetimeStartPos, a.DatetimeLayout, a.ScoreStyle, a.ScoreNSize, a.ignoreCount)
	if err != nil {
		return err
	}
	a.trans = trans
	if a.RootDir == "" {
		a.nTopRareLogs, err = newNTopRecords("ntop", a.NTopRecordsCount,
			0.0, trans, true, "", a.NRareTerms)
		if err != nil {
			return err
		}
	}

	stats, err := newStats(a.RootDir, a.MaxBlocks, a.BlockSize)
	if err != nil {
		return err
	}
	a.stats = stats

	logRecs, err := newLogRecords(a.RootDir, a.MaxBlocks, a.BlockSize)
	if err != nil {
		return err
	}
	a.logRecs = logRecs
	return nil
}

func (a *rarityAnalyzer) saveLastStatus() error {
	if a.RootDir == "" {
		return nil
	}

	var epoch int64
	rowNo := 0
	if a.fp != nil {
		epoch = a.fp.currFileEpoch()
		rowNo = a.fp.row()
	} else {
		epoch = 0
		rowNo = 0
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

func (a *rarityAnalyzer) saveConfig() error {
	filterReStr := re2str(a.filterRe)
	xFilterReStr := re2str(a.xFilterRe)

	if err := a.configTable.Upsert(nil, map[string]interface{}{
		"logPathRegex":     a.LogPathRegex,
		"blockSize":        a.BlockSize,
		"maxBlocks":        a.MaxBlocks,
		"maxItemBlocks":    a.MaxItemBlocks,
		"filterRe":         filterReStr,
		"xFilterRe":        xFilterReStr,
		"minGapToRecord":   a.MinGapToRecord,
		"datetimeStartPos": a.DatetimeStartPos,
		"datetimeLayout":   a.DatetimeLayout,
		"scoreStyle":       a.ScoreStyle,
	}); err != nil {
		return err
	}
	return nil
}

func (a *rarityAnalyzer) close() {
	if a == nil {
		return
	}

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
	if a.RootDir == "" {
		return nil
	}
	if IsDebug {
		msg := "rarityAnalyzer.commit(): started"
		ShowDebug(msg)
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
	if IsDebug {
		msg := "rarityAnalyzer.commit(): completed"
		ShowDebug(msg)
	}
	return nil
}

func (a *rarityAnalyzer) prepareDB() error {
	d, err := csvdb.NewCsvDB(a.RootDir)
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

func (a *rarityAnalyzer) initBlocks() error {
	if a.MaxBlocks == 0 && a.BlockSize == 0 {
		cnt, fileCnt, err := countNFiles(cNFilesToCheckCount, a.LogPathRegex)
		if err != nil {
			return err
		}
		a.calcBlocks(cnt, fileCnt)
	}
	return nil
}

func (a *rarityAnalyzer) setMaxBlock(m float64) {
	a.MaxBlocks = int(cLogCycle * (float64(m) / float64(a.BlockSize)))
	a.MaxItemBlocks = a.MaxBlocks * 5
}

func (a *rarityAnalyzer) calcBlocks(totalCount int, nFiles int) {
	if nFiles == 0 {
		nFiles = 1
	}
	m := float64(totalCount) / float64(nFiles)
	if m < 3000 {
		a.ModeblockPerFile = true
		a.MinGapToRecord = 0.3
		a.MaxBlocks = cLogCycle * 2
		a.BlockSize = 3000
		return
	} else {
		a.ModeblockPerFile = false
	}

	if m >= 3000 && m < 30000 {
		a.BlockSize = 1000
		a.MinGapToRecord = 0.5
		a.setMaxBlock(m)
	}
	if m >= 30000 && m < 300000 {
		a.BlockSize = 10000
		a.MinGapToRecord = 1.2
		a.setMaxBlock(m)
	}
	if m >= 300000 {
		a.BlockSize = 100000
		a.MinGapToRecord = 1.5
		a.setMaxBlock(m)
	}
}

func (a *rarityAnalyzer) analyze(targetLinesCnt int) error {
	if IsDebug {
		msg := "rarityAnalyzer.analyze(): "
		msg += fmt.Sprintf("blockSize=%d maxBlocks=%d maxItemBlocks=%d minGap=%1.1f",
			a.BlockSize, a.MaxBlocks, a.MaxItemBlocks, a.MinGapToRecord)
		ShowDebug(msg)
	}

	linesProcessed := 0
	var err error

	if a.fp == nil || !a.fp.isOpen() {
		a.fp, err = newFilePointer(a.LogPathRegex, a.lastFileEpoch, a.lastFileRow)
		if err != nil {
			return err
		}
		if err := a.fp.open(); err != nil {
			return err
		}
	}
	var lastEpoch int64
	for a.fp.next() {
		if linesProcessed > 0 && linesProcessed%cLogPerLines == 0 {
			log.Printf("processed %d lines", linesProcessed)
		}

		te := a.fp.text()
		if te == "" {
			//linesProcessed++
			continue
		}

		timeTran, tran, prh, dt, err := a.trans.tokenizeLine(te, a.filterRe, a.xFilterRe, true)
		if err != nil {
			a.linesProcessed = linesProcessed
			return err
		}
		lineEpoch := dt.Unix()
		tran = append(tran, timeTran...)
		if len(tran) == 0 {
			linesProcessed++
			continue
		}
		if lineEpoch > 0 {
			lastEpoch = lineEpoch
		} else {
			lastEpoch = a.fp.currFileEpoch()
		}
		a.trans.items.lastEpoch = lastEpoch
		a.logRecs.lastEpoch = lastEpoch

		score := a.trans.calcScore(prh, tran)
		err = a.stats.registerScore(score, lastEpoch)
		if err != nil {
			a.linesProcessed = linesProcessed
			return err
		}
		if a.stats.lastGap >= a.MinGapToRecord {
			if err := a.logRecs.insert(a.rowID, score, te, lastEpoch); err != nil {
				a.linesProcessed = linesProcessed
				return err
			}
		}
		if a.RootDir == "" {
			if !math.IsNaN(score) {
				a.nTopRareLogs.register(a.rowID, score, te, false)
			}
		}

		if a.fp.isEOF && (!a.fp.isLastFile() || a.ModeblockPerFile) {
			if err := a.saveLastStatus(); err != nil {
				a.linesProcessed = linesProcessed
				return err
			}
		}
		linesProcessed++

		a.rowID++
		if targetLinesCnt > 0 && linesProcessed >= targetLinesCnt {
			break
		}
	}
	a.linesProcessed = linesProcessed
	if err := a.commit(false); err != nil {
		return err
	}
	log.Printf("processed %d lines", linesProcessed)
	return nil
}

func (a *rarityAnalyzer) scanAndGetNTop(nTopname string, recordsToShow int, startEpoch, endEpoch int64,
	filterReStr, xFilterReStr string,
	minScore float64, maxScore float64, itemsToShow int) (*nTopRecords, error) {
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
	if err := rows.OrderBy([]string{"lastIndex"}, []string{"int64"},
		csvdb.CorderByAsc); err != nil {
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
	//startEpoch = minEpoch
	//endEpoch = maxEpoch

	lastEpochIdx = getColIdx("logRecords", "epoch")
	r, err := a.logRecs.selectRows(conditionCheckFunc,
		blockNos, []string{"rowID", "score", "record", "epoch"})
	if err != nil {
		return nil, err
	}

	ntop, err := newNTopRecords(nTopname, recordsToShow,
		minScore, a.trans, true, a.RootDir, itemsToShow)
	if err != nil {
		return nil, err
	}
	if err := ntop.load(a.rowID, a.BlockSize*a.MaxBlocks, false); err != nil {
		return nil, err
	}

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
		if math.IsNaN(score) {
			continue
		}
		if minScore > 0 && score < minScore {
			continue
		}
		if maxScore > 0 && score > maxScore {
			continue
		}

		b := []byte(record)
		if filterRe != nil && !filterRe.Match(b) {
			continue
		}
		if xFilterRe != nil && xFilterRe.Match(b) {
			continue
		}

		ntop.register(rowID, score, record, false)
	}

	return ntop, nil
}

func (a *rarityAnalyzer) getNTop(nTopName string, recordsToShow int, startEpoch, endEpoch int64,
	filterReStr, xFilterReStr string,
	minScore float64, maxScore float64, itemsToShow int) (*nTopRecords, error) {
	var err error
	var ntop *nTopRecords
	if IsDebug {
		msg := "rarityAnalyzer.getNTop(): "
		msg += fmt.Sprintf("nTopName=%s recordsToShow=%d filterReStr=%s xFilterReStr=%s",
			nTopName, recordsToShow, filterReStr, xFilterReStr)
		ShowDebug(msg)
	}
	if a.RootDir == "" {
		ntop = a.nTopRareLogs
	} else {
		ntop, err = a.scanAndGetNTop(nTopName, recordsToShow,
			startEpoch, endEpoch, filterReStr, xFilterReStr, minScore, maxScore, itemsToShow)
		if err != nil {
			return nil, err
		}
	}
	return ntop, nil
}

func (a *rarityAnalyzer) getRarStatsString(rootDir string, histSize int) (string, error) {
	var err error
	out, _, err := a.stats.getCountPerStatsString(0)
	if err != nil {
		return "", err
	}
	if a.RootDir == "" {
		out += "statistics\n"
		out += fmt.Sprintf("average= %f\n", a.stats.lastAverage)
		out += fmt.Sprintf("std=     %f\n", a.stats.lastStd)
		out += fmt.Sprintf("max=     %f\n", a.stats.scoreMax)
		out += "\n"
		return "", nil
	}
	return out, nil
}
