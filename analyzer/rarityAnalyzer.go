package analyzer

import (
	"fmt"

	"github.com/pkg/errors"
)

func newRarityAnalyzer(rootDir string) *rarityAnalyzer {
	a := new(rarityAnalyzer)
	a.rootDir = rootDir
	return a
}

func (a *rarityAnalyzer) clear() error {
	for _, dbName := range []string{"main", "logRecords", "items", "stats"} {
		if err := dropDB(a.rootDir, dbName); err != nil {
			return err
		}
	}
	return nil
}

func (a *rarityAnalyzer) init(logPathRegex, filterStr, xFilterStr string,
	minGapToRecord float64, maxBlocks, maxItemBlocks, linesInBlock int) error {
	a.logPathRegex = logPathRegex
	a.filterRe = getRegex(filterStr)
	a.xFilterRe = getRegex(xFilterStr)

	if err := ensureDir(a.rootDir); err != nil {
		return err
	}
	initLog(a.rootDir)
	logInfo("init() start")

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
		if err := a.saveConfig(); err != nil {
			return err
		}
		if err := a.saveLastStatus(); err != nil {
			return err
		}
	}
	logInfo("init() completed")

	return nil
}

func (a *rarityAnalyzer) load() error {
	initLog(a.rootDir)
	logInfo("load() start")

	d, err := newDB(a.rootDir, "main")
	if err != nil {
		return err
	}
	a.db = d
	sqlstr := `SELECT 
logPathRegex, linesInBlock, maxBlocks, maxItemBlocks, filterRe, xFilterRe, minGapToRecord
FROM config;`
	filterReStr := ""
	xFilterReStr := ""
	err = d.select1rec(sqlstr, &a.logPathRegex, &a.linesInBlock, &a.maxBlocks, &a.maxItemBlocks,
		&filterReStr, &xFilterReStr, &a.minGapToRecord)
	if err != nil {
		return err
	}
	a.filterRe = getRegex(filterReStr)
	a.xFilterRe = getRegex(xFilterReStr)

	sqlstr = `SELECT 
	lastRowID, lastFileEpoch, lastFileRow FROM lastStatus;`
	err = d.select1rec(sqlstr, &a.rowID, &a.lastFileEpoch, &a.lastFileRow)
	if err != nil {
		return err
	}

	if err := a.openObjs(); err != nil {
		return err
	}

	logInfo("load() completed")
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

func (a *rarityAnalyzer) saveConfig() error {
	d, err := newDB(a.rootDir, "main")
	if err != nil {
		return err
	}
	a.db = d
	if err := d.createTable("config"); err != nil {
		return err
	}
	sqlstr := `REPLACE INTO config(rootDir,
logPathRegex, linesInBlock, maxBlocks, maxItemBlocks, filterRe, xFilterRe, minGapToRecord)
VALUES (?,?,?,?,?,?,?,?)`
	stmt, err := d.conn.Prepare(sqlstr)
	if err != nil {
		return errors.WithStack(err)
	}

	filterReStr := re2str(a.filterRe)
	xFilterReStr := re2str(a.xFilterRe)

	_, err = stmt.Exec(a.rootDir, a.logPathRegex, a.linesInBlock, a.maxBlocks, a.maxItemBlocks,
		filterReStr, xFilterReStr, a.minGapToRecord)
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}

func (a *rarityAnalyzer) saveLastStatus() error {
	d := a.db
	if err := d.createTable("lastStatus"); err != nil {
		return err
	}
	if _, err := d.exec(`DELETE FROM lastStatus;`); err != nil {
		return err
	}

	sqlstr := `INSERT INTO lastStatus(lastRowID, lastFileEpoch, lastFileRow) VALUES (?,?,?)`
	stmt, err := d.conn.Prepare(sqlstr)
	if err != nil {
		return errors.WithStack(err)
	}
	if a.fp != nil {
		_, err = stmt.Exec(a.rowID, a.fp.currFileEpoch(), a.fp.row())
	} else {
		_, err = stmt.Exec(a.rowID, 0, 0)
	}
	if err != nil {
		return errors.WithStack(err)
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
		a.logRecs.close()
	}
	if a.trans != nil {
		a.trans.close()
	}
	logInfo("close() completed")
}

func (a *rarityAnalyzer) commit(completed bool) error {
	logInfo("commit() start")
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
	logInfo("commit() completed")
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
		err = a.stats.registerScore(score)
		if err != nil {
			return linesProcessed, err
		}
		if a.stats.lastGap >= a.minGapToRecord {
			if err := a.logRecs.insertRow(a.rowID, score, te); err != nil {
				return linesProcessed, err
			}
		}
		if a.fp.isEOF && !a.fp.isLastFile() {
			if err := a.saveLastStatus(); err != nil {
				return linesProcessed, err
			}
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
	wherestr := ""
	if startEpoch > 0 {
		wherestr = fmt.Sprintf("WHERE lastEpoch=%d", startEpoch)
	}
	rows, err := a.logRecs.query(`SELECT blockNo FROM circuitDBStatus` + wherestr)
	if err != nil {
		return nil, err
	}
	blockNos := make([]int, 0)
	for rows.Next() {
		blockNo := 0
		if err := rows.Scan(&blockNo); err != nil {
			return nil, err
		}
		blockNos = append(blockNos, blockNo)
	}

	r, err := a.logRecs.selectRows([]string{"rowId", "score", "record"}, "", blockNos)
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

func (a *rarityAnalyzer) printNTops(msg string,
	recordsToShow int, startEpoch int64,
	filterReStr, xFilterReStr string,
) error {
	var err error
	var nTopRareLogs []*colLogRecords
	nTopRareLogs, err = a.scanAndGetNTops(recordsToShow, startEpoch,
		filterReStr, xFilterReStr)
	if err != nil {
		return err
	}

	fmt.Printf("%s\n", msg)
	fmt.Print("score   rowID      text\n")
	fmt.Print("-------+----------+-------\n")
	for i, logr := range nTopRareLogs {
		if logr == nil {
			break
		}
		fmt.Printf(" %5.2f   %8d   %s\n", logr.score, logr.rowid, logr.record)
		if logr.score == 0 {
			break
		}
		if i+1 >= recordsToShow {
			break
		}
	}
	return nil
}
