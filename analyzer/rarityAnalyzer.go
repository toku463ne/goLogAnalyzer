package analyzer

import (
	"github.com/pkg/errors"
)

func newRarityAnalyzer(rootDir string) *rarityAnalyzer {
	a := new(rarityAnalyzer)
	a.rootDir = rootDir
	return a
}

func (a *rarityAnalyzer) clear() error {
	for _, dbName := range []string{"main", "logRecords", "items"} {
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

	a.setDefaults()

	if minGapToRecord >= 0 {
		a.minGapToRecord = minGapToRecord
	}
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
		err := a.saveConfig()
		if err != nil {
			return err
		}
	}

	return nil
}

func (a *rarityAnalyzer) load() error {
	a.setDefaults()
	d, err := newDB(a.rootDir, "main")
	if err != nil {
		return err
	}
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
	lastFileEpoch, lastFileRow FROM lastStatus;`
	err = d.select1rec(sqlstr, &a.lastFileEpoch, &a.lastFileRow)
	if err != nil {
		return err
	}

	if err := a.openObjs(); err != nil {
		return err
	}

	return nil
}

func (a *rarityAnalyzer) setDefaults() {
	a.minGapToRecord = cMinGapToRecord
	a.maxBlocks = cDefaultMaxBlocks
	a.maxItemBlocks = cDefaultMaxItemBlocks
	a.linesInBlock = cDefaultBlockSize
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

	if err := d.createTable("lastStatus"); err != nil {
		return err
	}
	sqlstr = `REPLACE INTO lastStatus(lastFileEpoch, lastFileRow) VALUES (?,?)`
	stmt, err = d.conn.Prepare(sqlstr)
	if err != nil {
		return errors.WithStack(err)
	}
	_, err = stmt.Exec(0, 0)
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
}

func (a *rarityAnalyzer) commit(completed bool) error {
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
	return nil
}

func (a *rarityAnalyzer) registerItems(targetLinesCnt int) error {
	linesProcessed := 0

	if a.fp == nil || !a.fp.isOpen() {
		a.fp = newFilePointer(a.logPathRegex, a.lastFileEpoch, a.lastFileRow)
		if err := a.fp.open(); err != nil {
			return err
		}
	}
	for a.fp.next() {
		te := a.fp.text()
		if te == "" {
			continue
		}
		tran, err := a.trans.tokenizeLine(te, a.filterRe, a.xFilterRe, true)
		if err != nil {
			return err
		}
		if len(tran) == 0 {
			continue
		}

		linesProcessed++
		if targetLinesCnt > 0 && linesProcessed >= targetLinesCnt {
			break
		}
	}
	a.fp = nil
	return nil
}

func (a *rarityAnalyzer) calcRare(targetLinesCnt int, startEpoch int64) (int, error) {
	linesProcessed := 0

	if a.fp == nil || !a.fp.isOpen() {
		a.fp = newFilePointer(a.logPathRegex, startEpoch, 0)
		if err := a.fp.open(); err != nil {
			return 0, err
		}
	}
	for a.fp.next() {
		te := a.fp.text()
		if te == "" {
			continue
		}
		tran, err := a.trans.tokenizeLine(te, a.filterRe, a.xFilterRe, false)
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

		linesProcessed++
		if targetLinesCnt > 0 && linesProcessed >= targetLinesCnt {
			break
		}
	}
	if err := a.commit(false); err != nil {
		return linesProcessed, err
	}
	return linesProcessed, nil
}

func (a *rarityAnalyzer) run(targetLinesCnt int, startEpoch int64) (int, error) {
	linesProcessed := 0
	var err error

	if err := a.registerItems(targetLinesCnt); err != nil {
		return 0, err
	}
	linesProcessed, err = a.calcRare(targetLinesCnt, startEpoch)
	if err != nil {
		return linesProcessed, err
	}
	return linesProcessed, nil
}
