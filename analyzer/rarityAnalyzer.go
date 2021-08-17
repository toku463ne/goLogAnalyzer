package analyzer

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/pkg/errors"
	csvdb "github.com/toku463ne/goCsvDb"
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
	minGapToRecord float64, maxBlocks, maxItemBlocks, linesInBlock, nTopRecords int) error {
	a.logPathRegex = logPathRegex
	a.filterRe = getRegex(filterStr)
	a.xFilterRe = getRegex(xFilterStr)
	a.nTopRareLogs = make([]*colLogRecords, nTopRecords)

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
			"maxItemBlocks", "filterRe", "xFilterRe", "minGapToRecord"},
		&a.logPathRegex, &a.linesInBlock, &a.maxBlocks, &a.maxItemBlocks,
		&filterReStr, &xFilterReStr, &a.minGapToRecord); err != nil {
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
	trans, err := newTrans(a.rootDir, a.maxItemBlocks, a.linesInBlock)
	if err != nil {
		return err
	}
	a.trans = trans

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
	if err := a.stats.load(); err != nil {
		return err
	}
	if err := a.logRecs.load(); err != nil {
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
	return nil
}

func (a *rarityAnalyzer) saveConfig() error {
	filterReStr := re2str(a.filterRe)
	xFilterReStr := re2str(a.xFilterRe)

	if err := a.configTable.Upsert(nil, map[string]interface{}{
		"logPathRegex":   a.logPathRegex,
		"linesInBlock":   a.linesInBlock,
		"maxBlocks":      a.maxBlocks,
		"maxItemBlocks":  a.maxItemBlocks,
		"filterRe":       filterReStr,
		"xFilterRe":      xFilterReStr,
		"minGapToRecord": a.minGapToRecord,
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

		lastEpoch = a.fp.currFileEpoch()
		a.trans.items.lastEpoch = lastEpoch
		a.logRecs.lastEpoch = lastEpoch

		tran, err := a.trans.tokenizeLine(te, a.filterRe, a.xFilterRe, true)
		if err != nil {
			return linesProcessed, err
		}
		if len(tran) == 0 {
			continue
		}

		score := a.trans.calcScore(tran)
		err = a.stats.registerScore(score, lastEpoch)
		if err != nil {
			return linesProcessed, err
		}
		if a.stats.lastGap >= a.minGapToRecord {
			if err := a.logRecs.insert(a.rowID, score, te, a.lastFileEpoch); err != nil {
				return linesProcessed, err
			}
		}
		if a.rootDir == "" {
			nTopRareLogs, m := registerNTopRareRec(a.nTopRareLogs, a.maxScore, a.rowID, score, te)
			a.nTopRareLogs = nTopRareLogs
			a.maxScore = m
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

func (a *rarityAnalyzer) scanAndGetNTops(recordsToShow int, startEpoch int64,
	filterReStr, xFilterReStr string) ([]*colLogRecords, error) {
	var conditionCheckFunc func(v []string) bool

	if startEpoch > 0 {
		lastEpochIdx := getColIdx("circuitDBStatus", "lastEpoch")
		conditionCheckFunc = func(v []string) bool {
			lastEpoch, _ := strconv.ParseInt(v[lastEpochIdx], 10, 64)
			return lastEpoch >= startEpoch
		}
	} else {
		conditionCheckFunc = nil
	}

	rows, err := a.logRecs.statusTable.SelectRows(conditionCheckFunc, []string{"blockNo"})
	if err != nil {
		return nil, errors.WithStack(err)
	}

	blockNos := make([]int, 0)
	for rows.Next() {
		blockNo := 0
		if err := rows.Scan(&blockNo); err != nil {
			return nil, err
		}
		blockNos = append(blockNos, blockNo)
	}

	r, err := a.logRecs.selectRows(nil, blockNos, []string{"rowID", "score", "record"})
	if err != nil {
		return nil, err
	}

	filterRe := getRegex(filterReStr)
	xFilterRe := getRegex(xFilterReStr)
	var rowID int64
	var score float64
	var record string
	nTopRareLogs := make([]*colLogRecords, recordsToShow)
	m := 0.0
	for r.next() {
		if err := r.scan(&rowID, &score, &record); err != nil {
			return nil, err
		}
		if filterRe != nil && !filterRe.Match([]byte(record)) {
			continue
		}
		if xFilterRe != nil && xFilterRe.Match([]byte(record)) {
			continue
		}
		nTopRareLogs, m = registerNTopRareRec(nTopRareLogs, m, rowID, score, record)
	}
	return nTopRareLogs, nil
}

func (a *rarityAnalyzer) showRarStats(rootDir string, histSize int) error {
	var g []int
	var err error
	if a.rootDir == "" {
		g = a.stats.countPerScore
	} else {
		g, err = a.stats.loadAllScorePerCount()
		if err != nil {
			return err
		}
	}

	s, err := a.stats.loadRecentStats(histSize)
	if err != nil {
		return err
	}

	if len(s) > 0 {
		lastStat := s[0]
		fmt.Printf("\n")
		fmt.Printf("Recent normal score border: %2.1f\n", lastStat.avg+lastStat.std*2)
	}
	fmt.Printf("\n")
	fmt.Printf("Counts per score\n")
	fmt.Printf(" score | count\n")
	fmt.Printf(" ------+--------------\n")
	for i := 0; i < cCountbyScoreLen; i++ {
		if g[i] > 0 {
			fmt.Printf("   %-2.1f | %d\n", float64(i), g[i])
		}
	}
	fmt.Println("")
	fmt.Println("")

	if a.rootDir == "" {
		fmt.Printf("statistics\n")
		fmt.Printf("average= %f\n", a.stats.lastAverage)
		fmt.Printf("std=     %f\n", a.stats.lastStd)
		fmt.Printf("max=     %f\n", a.stats.scoreMax)
		fmt.Printf("\n")
		return nil
	}

	fmt.Printf("score history\n")
	fmt.Printf(" last date           | average |     std |     max \n")
	fmt.Printf(" --------------------+---------+---------+---------\n")
	for _, rec := range s {
		if rec.lastFileEpoch == 0 {
			break
		}
		fmt.Printf(" %s |     %-3.1f |     %-3.1f |     %-3.1f \n",
			epochToString(rec.lastFileEpoch), rec.avg, rec.std, rec.max)
	}
	return nil
}

func (a *rarityAnalyzer) printNTops(msg string,
	recordsToShow int, startEpoch int64,
	filterReStr, xFilterReStr string,
) error {
	var err error
	var nTopRareLogs []*colLogRecords
	if a.rootDir == "" {
		nTopRareLogs = a.nTopRareLogs
	} else {
		nTopRareLogs, err = a.scanAndGetNTops(recordsToShow, startEpoch,
			filterReStr, xFilterReStr)
		if err != nil {
			return err
		}
	}

	fmt.Printf("%s\n", msg)
	fmt.Print("score     rowID      text\n")
	fmt.Print("---------+----------+-------\n")
	for i, logr := range nTopRareLogs {
		if logr == nil {
			break
		}
		fmt.Printf("   %-5.2f  %8d   %s\n", logr.score, logr.rowid, logr.record)
		if logr.score == 0 {
			break
		}
		if i+1 >= recordsToShow {
			break
		}
	}
	return nil
}
