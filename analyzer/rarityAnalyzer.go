package analyzer

import (
	"fmt"
	"math"
	"strconv"
	"time"

	"github.com/pkg/errors"
)

type block struct {
	blockID     int
	blockCnt    int
	lastEpoch   int64
	scoreSum    float64
	scoreSqrSum float64
	scoreGapMax float64
	scoreGapMin float64
	completed   bool
}

type rarityAnalyzer struct {
	// fields
	name                  string
	rowID                 int64
	rootDir               string
	useDB                 bool
	db                    *csvDB
	items                 *items
	trans                 *trans
	blocks                []*block
	scoreCount            int
	scoreSum              float64
	scoreSqrSum           float64
	currBlock             *block
	currBlockID           int
	oldBlockID            int
	rowNum                int
	targetLinesCnt        int
	maxTargetLinesCnt     int
	haveStatistics        bool
	linesProcessedInBlock int
	lastFileEpoch         int64
	lastFileRow           int

	filterRe        string
	xFilterRe       string
	linesInBlock    int
	maxBlocks       int
	rarityThreshold float64

	// functions
	outputFunc func(
		name string, rowID int64,
		scoreThreshold float64,
		score, scoreGap, scoreAvg, scoreStd float64,
		cnt int,
		text []string)
	setTargetLinesCnt func(int) error
	countTargetLines  func() (int, error)

	postDeleteOldFunc func(blockID int) error

	pointerNext      func() bool
	pointerText      func() []string
	pointerOpen      func() error
	pointerClose     func()
	pointerCurrEpoch func() int64
	pointerCurrName  func() string
	pointerPos       func() int
	pointerErr       func() error
}

func newBlock(blockID int) *block {
	b := new(block)
	b.blockID = blockID
	b.completed = false
	return b
}

func (a *rarityAnalyzer) init() error {
	a.items = newItems()
	maxBlockNum := int(math.Pow10(maxBlockDitigs))
	if a.maxBlocks >= maxBlockNum {
		return fmt.Errorf("maxBlocks can't be %d or bigger", maxBlockNum)
	}

	a.currBlockID = -1
	a.oldBlockID = -1

	a.blocks = make([]*block, a.maxBlocks)
	if a.useDB {
		if err := ensureDir(a.rootDir); err != nil {
			return err
		}
		db, err := getRarityAnalDB(a.rootDir, a.maxBlocks)
		if err != nil {
			return err
		}
		a.db = db
	}

	a.postDeleteOldFunc = func(blockID int) error {
		return nil
	}
	return nil
}

func (a *rarityAnalyzer) loadDB() error {
	lastBlockID := -1
	var lastEpoch int64
	var lastRow int

	tmpBlockID, tmpRow, tmpEpoch, err := a.loadDBLastStatus()
	if err != nil {
		return err
	}
	lastRow = tmpRow
	lastEpoch = tmpEpoch
	lastBlockID = tmpBlockID

	if a.haveStatistics {
		if err := a.loadDBBlocks(); err != nil {
			return err
		}
	} else {
	}

	if err := a.loadDBItems(); err != nil {
		return err
	}

	a.currBlockID = lastBlockID
	a.lastFileEpoch = lastEpoch
	a.lastFileRow = lastRow

	return nil
}

func (a *rarityAnalyzer) close() {
	if a.db != nil {
		a.db = nil
	}
	a.pointerClose()
}

func (a *rarityAnalyzer) loadDBItems() error {
	rows, err := a.db.tables["items"].query(nil, "*")
	if err != nil {
		return err
	}
	if rows == nil || len(rows) == 0 {
		return nil
	}
	i := newItems()
	for _, v := range rows {
		cnt, err := strconv.Atoi(v[1])
		if err != nil {
			return errors.WithStack(err)
		}
		i.regist(v[0], cnt, false)

	}
	a.items = i
	return nil
}

func (a *rarityAnalyzer) loadDBLastStatus() (int, int, int64, error) {
	v, err := a.db.tables["lastStatus"].select1rec(nil, "")
	if err != nil {
		return -1, 0, 0, err
	}
	if v == nil {
		return -1, 0, 0, nil
	}

	cols := a.db.tables["lastStatus"].colMap
	idxLastRowID := cols["lastRowID"]
	idxLastBlockID := cols["lastBlockID"]
	idxLastRow := cols["lastRow"]
	idxModifiedEpoch := cols["modifiedEpoch"]
	var lastEpoch int64
	var lastRow int
	lastBlockID := -1
	lastRowID, err := strconv.ParseInt(v[idxLastRowID], 10, 64)
	if err != nil {
		return -1, 0, 0, errors.WithStack(err)
	}
	a.rowID = lastRowID

	lastBlockID, err = strconv.Atoi(v[idxLastBlockID])
	if err != nil {
		return -1, 0, 0, errors.WithStack(err)
	}
	tmpEpoch1, err := strconv.Atoi(v[idxModifiedEpoch])
	if err != nil {
		return -1, 0, 0, errors.WithStack(err)
	}
	tmpEpoch2 := int64(tmpEpoch1)
	lastEpoch = tmpEpoch2
	tmpRow, err := strconv.Atoi(v[idxLastRow])
	if err != nil {
		return -1, 0, 0, errors.WithStack(err)
	}
	lastRow = tmpRow

	return lastBlockID, lastRow, lastEpoch, nil
}

func (a *rarityAnalyzer) loadDBBlocks() error {
	rows, err := a.db.tables["logBlocks"].query(nil, "*")
	if err != nil {
		return err
	}
	if rows == nil || len(rows) == 0 {
		return nil
	}
	cols := a.db.tables["logBlocks"].colMap
	idxBlockID := cols["blockID"]
	idxBlockCnt := cols["blockCnt"]
	idxScoreSum := cols["scoreSum"]
	idxScoreSqrSum := cols["scoreSqrSum"]
	idxLastEpoch := cols["lastEpoch"]
	for _, v := range rows {
		blockID, err := strconv.Atoi(v[idxBlockID])
		if err != nil {
			return errors.WithStack(err)
		}
		blockCnt, err := strconv.Atoi(v[idxBlockCnt])
		if err != nil {
			return errors.WithStack(err)
		}
		scoreSum, err := strconv.ParseFloat(v[idxScoreSum], 64)
		if err != nil {
			return errors.WithStack(err)
		}
		scoreSqrSum, err := strconv.ParseFloat(v[idxScoreSqrSum], 64)
		if err != nil {
			return errors.WithStack(err)
		}
		tmpEpoch1, err := strconv.Atoi(v[idxLastEpoch])
		if err != nil {
			return errors.WithStack(err)
		}
		tmpEpoch2 := int64(tmpEpoch1)
		if err != nil {
			return errors.WithStack(err)
		}

		b := newBlock(blockID)
		b.blockCnt = blockCnt
		b.lastEpoch = tmpEpoch2
		b.scoreSum = scoreSum
		b.scoreSqrSum = scoreSqrSum
		a.blocks[blockID] = b

		//a.rowNum += blockCnt
		a.scoreCount += blockCnt
		a.scoreSum += scoreSum
		a.scoreSqrSum += scoreSqrSum

	}

	return nil
}

func (a *rarityAnalyzer) deleteOld(blockID int) error {
	if blockID < 0 {
		return nil
	}
	b := a.blocks[blockID]
	if b == nil {
		return nil
	}

	blockIDstr := fmt.Sprint(blockID)

	a.scoreCount -= b.blockCnt
	a.scoreSum -= b.scoreSum
	a.scoreSqrSum -= b.scoreSqrSum
	if a.useDB {
		rows, err := a.db.tables["items"].query(nil, fmt.Sprint(blockID))
		if err != nil {
			return err
		}
		if rows == nil || len(rows) == 0 {
			return nil
		}
		for _, v := range rows {
			itemID, ok := a.items.getItemID(v[0])
			if !ok {
				continue
			}

			cnt, err := strconv.Atoi(v[1])
			if err != nil {
				return errors.WithStack(err)
			}
			a.items.addCount(itemID, -cnt)

		}

		a.db.tables["logBlocks"].dropPartition(blockIDstr)
		a.db.tables["items"].dropPartition(blockIDstr)
	}

	if err := a.postDeleteOldFunc(blockID); err != nil {
		return err
	}

	a.blocks[blockID] = nil

	return nil
}

func (a *rarityAnalyzer) saveLastStatus(blockID int) error {
	// save last status
	epoch := a.pointerCurrEpoch()
	cond := map[string]string{}
	row := map[string]string{
		"lastRowID":       fmt.Sprint(a.rowID),
		"lastBlockID":     fmt.Sprint(a.currBlockID),
		"fileName":        a.pointerCurrName(),
		"lastRow":         fmt.Sprint(a.pointerPos()),
		"modifiedEpoch":   fmt.Sprint(epoch),
		"modifiedUtcTime": fmt.Sprint(epochToString(epoch)),
	}
	if err := a.db.tables["lastStatus"].update(
		cond, row, "", true); err != nil {
		return err
	}
	return nil
}

func (a *rarityAnalyzer) saveBlock() error {
	b := a.currBlock
	if b == nil {
		return nil
	}
	blockID := b.blockID
	// save block
	cond := map[string]string{
		"blockID": fmt.Sprint(a.currBlockID),
	}
	row := map[string]string{
		"lastRowID":   fmt.Sprint(a.rowID),
		"blockCnt":    fmt.Sprint(a.linesInBlock),
		"scoreSum":    fmt.Sprint(b.scoreSum),
		"scoreSqrSum": fmt.Sprint(b.scoreSqrSum),
		"createdAt":   timeToString(time.Now()),
	}
	if err := a.db.tables["logBlocks"].update(
		cond, row, fmt.Sprint(blockID), true); err != nil {
		return err
	}
	return nil
}

func (a *rarityAnalyzer) saveItems(blockID int) error {
	// save item
	items := a.items
	rows := [][]string{}
	table := a.db.tables["items"]
	cols := table.colMap
	for itemID, cnt := range items.newCounts.getSlice() {
		if cnt > 0 {
			row := make([]string, len(cols))
			row[cols["word"]] = string(items.items.get(itemID))
			row[cols["cnt"]] = fmt.Sprint(cnt)
			rows = append(rows, row)
		}
	}
	if err := a.db.tables["items"].insertRows(rows, fmt.Sprint(blockID)); err != nil {
		return err
	}
	return nil
}

func (a *rarityAnalyzer) clean() error {
	return a.db.dropAllTables()
}

func (a *rarityAnalyzer) initCurrBlock(blockID int) {
	a.trans = newTrans()
	a.items.clearNewCount()
	if a.blocks[blockID] != nil {
		a.items = a.items.getNonZeroItems()
	}
	a.currBlock = newBlock(blockID)
}

func (a *rarityAnalyzer) nextBlock() {
	if a.currBlock != nil {
		msg := fmt.Sprintf("block=%d finished: gapMax=%f gapMin=%f",
			a.currBlock.blockID,
			a.currBlock.scoreGapMax,
			a.currBlock.scoreGapMin)
		logDebug(msg)
	}

	a.oldBlockID = a.currBlockID
	a.currBlockID++
	if a.currBlockID-a.maxBlocks >= 0 {
		a.currBlockID = 0
	}
	a.initCurrBlock(a.currBlockID)
}

func (a *rarityAnalyzer) postBlock() error {
	blockID := a.currBlockID
	if blockID >= 0 && a.currBlock != nil {
		if err := a.deleteOld(blockID); err != nil {
			return err
		}

		if a.useDB {
			if err := a.saveLastStatus(blockID); err != nil {
				return err
			}
			if err := a.saveBlock(); err != nil {
				return err
			}
			if err := a.saveItems(blockID); err != nil {
				return err
			}
		}
		a.blocks[a.currBlockID] = a.currBlock
	}
	return nil
}

func (a *rarityAnalyzer) run(targetLinesCnt int, forcedBlockID int) (int, error) {
	var scoreAvg float64
	var scoreStd float64
	var scoreGap float64
	linesProcessed := 0
	logInfo(fmt.Sprintf("data=%s search=%s exclude=%s gap=%f bsize=%d nblocks=%d",
		a.rootDir,
		a.filterRe, a.xFilterRe,
		a.rarityThreshold,
		a.linesInBlock, a.maxBlocks))

	a.setTargetLinesCnt(targetLinesCnt)
	if a.targetLinesCnt > 0 && a.rowNum >= a.targetLinesCnt {
		return 0, nil
	}

	if a.rowNum == 0 && forcedBlockID < 0 {
		a.nextBlock()
	}

	if forcedBlockID >= 0 {
		a.initCurrBlock(forcedBlockID)
		a.currBlockID = forcedBlockID
	}

	if err := a.pointerOpen(); err != nil {
		return 0, err
	}

	for a.pointerNext() {
		if a.currBlock.completed && forcedBlockID < 0 {
			a.nextBlock()
		}

		a.linesProcessedInBlock++
		a.rowNum++
		if a.rowNum > int(maxRowID) {
			a.rowNum = 0
		}
		a.rowID++
		if a.rowID > maxRowID {
			a.rowID = 0
		}

		te := a.pointerText()
		isAdded := tokenizeLine(te[0], a.trans,
			a.items, a.filterRe, a.xFilterRe, a.linesProcessedInBlock)

		if isAdded && a.haveStatistics {
			score := a.trans.getLastScore()
			scoreSqr := score * score
			a.currBlock.scoreSum += score
			a.currBlock.scoreSqrSum += scoreSqr
			a.currBlock.blockCnt++
			a.scoreSum += score
			a.scoreSqrSum += scoreSqr
			a.scoreCount++
			cnt := a.scoreCount
			scoreAvg = a.scoreSum / float64(cnt)
			scoreStd = math.Sqrt(a.scoreSqrSum / float64(cnt))
			if scoreStd == 0 {
				scoreGap = 0
			} else {
				scoreGap = (score - scoreAvg) / scoreStd
			}
			if a.currBlock.scoreGapMax == 0 || a.currBlock.scoreGapMax < scoreGap {
				a.currBlock.scoreGapMax = scoreGap
			}
			if a.currBlock.scoreGapMin == 0 || a.currBlock.scoreGapMin > scoreGap {
				a.currBlock.scoreGapMin = scoreGap
			}
			a.outputFunc(a.name, a.rowID, a.rarityThreshold,
				score, scoreGap, scoreAvg, scoreStd, cnt, te)
		}

		linesProcessed++
		if a.linesInBlock > 0 && a.linesProcessedInBlock >= a.linesInBlock {
			a.currBlock.completed = true
			if err := a.postBlock(); err != nil {
				return linesProcessed, err
			}
			a.linesProcessedInBlock = 0
		}
		if a.targetLinesCnt > 0 && a.rowNum >= a.targetLinesCnt {
			break
		}
	}
	if a.linesInBlock == 0 {
		if err := a.postBlock(); err != nil {
			return linesProcessed, err
		}
		a.linesProcessedInBlock = 0
	}
	return linesProcessed, nil
}

func (a *rarityAnalyzer) runBeforeNextBlock(targetLinesCnt int) (int, error) {
	if a.linesInBlock <= 0 {
		return a.run(targetLinesCnt, -1)
	}
	linesToProcess := a.linesInBlock - a.linesProcessedInBlock
	if targetLinesCnt > 0 {
		if targetLinesCnt < linesToProcess {
			linesToProcess = targetLinesCnt
		}
	}
	return a.run(linesToProcess, -1)
}
