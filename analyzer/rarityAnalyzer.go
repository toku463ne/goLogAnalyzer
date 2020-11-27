package analyzer

import (
	"fmt"
	"io"
	"math"
	"strconv"
	"time"

	"github.com/pkg/errors"
)

type outputScoreFunc func(
	name string, rowID int64,
	scoreThreshold float64,
	score, scoreGap, scoreAvg, scoreStd float64,
	cnt int,
	text string)

type block struct {
	blockID     int
	blockCnt    int
	lastEpoch   int64
	scoreSum    float64
	scoreSqrSum float64
}

type rarityAnalyzerVars struct {
	name            string
	useDB           bool
	filterRe        string
	xFilterRe       string
	rootDir         string
	logPathRegex    string
	rarityThreshold float64
	linesInBlock    int
	maxBlocks       int
}

type rarityAnalyzer struct {
	name            string
	useDB           bool
	filterRe        string
	xFilterRe       string
	rowID           int64
	rowNum          int
	rootDir         string
	logPathRegex    string
	db              *csvDB
	items           *items
	trans           *trans
	blocks          []*block
	scoreCount      int
	scoreSum        float64
	scoreSqrSum     float64
	lastMsg         string
	currBlockID     int
	linesInBlock    int
	maxBlocks       int
	targetLinesCnt  int
	fp              *filePointer
	rarityThreshold float64
	lastFileEpoch   int64
	lastFileRow     int
	outputFunc      outputScoreFunc
	haveStatistics  bool
}

func newRarityAnalyzer() *rarityAnalyzer {
	a := new(rarityAnalyzer)
	a.name = "rarityAnal"
	//a.logPathRegex = v.logPathRegex
	a.useDB = true
	a.rootDir = fmt.Sprintf("./%s", a.name)
	a.rarityThreshold = 0.8
	a.linesInBlock = 10000
	a.maxBlocks = 1000
	a.haveStatistics = true
	a.outputFunc = outputScoreFunc(func(name string, rowID int64,
		scoreThreshold float64,
		score, scoreGap, scoreAvg, scoreStd float64,
		cnt int,
		text string) {
		if verbose || scoreGap > scoreThreshold {
			msg := fmt.Sprintf("%s %5d s=%5.2f g=%5.2f a=%5.2f d=%5.2f c=%5d | %s",
				name,
				rowID,
				score,
				scoreGap,
				scoreAvg,
				scoreStd,
				cnt,
				text,
			)
			println(msg)
		}
	})
	return a
}

func newRarityAnalyzerByVars(v *rarityAnalyzerVars) (*rarityAnalyzer, error) {
	a := newRarityAnalyzer()
	a.name = v.name
	a.useDB = v.useDB
	a.filterRe = v.filterRe
	a.xFilterRe = v.xFilterRe
	a.rootDir = v.rootDir
	a.logPathRegex = v.logPathRegex
	a.rarityThreshold = v.rarityThreshold
	a.linesInBlock = v.linesInBlock
	a.maxBlocks = v.maxBlocks

	if err := a.init(); err != nil {
		return nil, err
	}
	return a, nil
}

func (a *rarityAnalyzer) init() error {
	a.items = newItems()

	maxBlockNum := int(math.Pow10(maxBlockDitigs))
	if a.maxBlocks >= maxBlockNum {
		return fmt.Errorf("maxBlocks can't be %d or bigger", maxBlockNum)
	}
	a.blocks = make([]*block, a.maxBlocks)

	a.currBlockID = -1
	a.lastFileEpoch = 0.0
	a.lastFileRow = 0

	if err := ensureDir(a.rootDir); err != nil {
		return err
	}

	if a.useDB {
		db, err := getRarityAnalDB(a.rootDir, a.maxBlocks)
		if err != nil {
			return err
		}
		a.db = db
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

func (a *rarityAnalyzer) openFP(logPathRegex string,
	lastEpoch int64, lastRow int) error {
	a.fp = newFilePointer(logPathRegex, lastEpoch, lastRow)
	return a.fp.open()
}

func (a *rarityAnalyzer) run() error {
	targetLinesCnt, err := a.countTargetLines()
	if err != nil {
		return err
	}
	a.targetLinesCnt = targetLinesCnt
	if err := a.openFP(a.logPathRegex, a.lastFileEpoch, a.lastFileRow); err != nil {
		return err
	}
	for {
		a.currBlockID = a.nextBlockID(a.currBlockID)

		ok, err := a.runBlock(a.currBlockID)
		if err != nil {
			return err
		}
		if ok == false {
			break
		}
		if a.rowNum+a.linesInBlock > a.targetLinesCnt {
			break
		}
	}
	return nil
}

func (a *rarityAnalyzer) close() {
	if a.db != nil {
		a.db = nil
	}
	if a.fp != nil {
		a.fp.close()
	}
}

func (a *rarityAnalyzer) nextBlockID(blockID int) int {
	blockID++
	if blockID-a.maxBlocks >= 0 {
		blockID = 0
	}
	return blockID
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

		b := new(block)
		b.blockID = blockID
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

func (a *rarityAnalyzer) countTargetLines() (int, error) {
	targetLinesCntCnt := 0
	if err := a.openFP(a.logPathRegex, a.lastFileEpoch, a.lastFileRow); err != nil {
		return -1, err
	}
	defer a.fp.close()
	for a.fp.next() {
		targetLinesCntCnt++
	}
	if err := a.fp.err(); err != nil {
		return -1, err
	}
	return targetLinesCntCnt, nil
}

func (a *rarityAnalyzer) tokenizeLine(line string, row int) bool {
	isAdded := false

	if a.filterRe != "" && searchReg(line, a.filterRe) == false {
		return isAdded
	}
	if a.xFilterRe != "" && searchReg(line, a.xFilterRe) {
		return isAdded
	}

	a.trans.mask.append(row)
	bline := []byte(line)

	tran := getEnItems(bline, a.items, a.filterRe)

	if len(tran) > 0 {
		a.trans.add(tran, line, a.items)
		if verbose {
			tranID := a.trans.maxTranID
			a.trans.lastMsg = fmt.Sprintf("%s",
				a.trans.getSentenceAt(tranID, a.items))
		}
		isAdded = true
	}
	return isAdded
}

func (a *rarityAnalyzer) runBlock(blockID int) (bool, error) {
	a.trans = newTrans()
	a.items.clearNewCount()
	if blockID >= 0 {
		if err := a.deleteOld(blockID); err != nil {
			return false, err
		}
	}

	ok, b, err := a.tokenizeBlock(blockID)
	if ok == false || err != nil {
		return ok, err
	}
	if a.useDB && blockID >= 0 {
		if err := a.saveLastStatus(blockID); err != nil {
			return false, err
		}
		if err := a.saveBlock(b); err != nil {
			return false, err
		}
		if err := a.saveItems(blockID); err != nil {
			return false, err
		}
	}
	return true, nil
}

func (a *rarityAnalyzer) deleteOld(blockID int) error {
	b := a.blocks[blockID]
	blockIDstr := fmt.Sprint(blockID)
	if b != nil {
		a.scoreCount -= b.blockCnt
		a.scoreSum -= b.scoreSum
		a.scoreSqrSum -= b.scoreSqrSum
		if a.useDB {
			a.db.tables["logBlocks"].deletePartition(blockIDstr)
			a.db.tables["items"].deletePartition(blockIDstr)
		}
		a.blocks[blockID] = nil
	}
	return nil
}

func (a *rarityAnalyzer) tokenizeBlock(blockID int) (bool, *block, error) {
	linesProcessedInBlock := 0
	var blockScoreSum float64
	var blockScoreSqrSum float64
	var scoreAvg float64
	var scoreStd float64
	var scoreGap float64

	for a.fp.next() {
		a.rowNum++
		a.rowID++

		linesProcessedInBlock++
		te := a.fp.text()
		isAdded := a.tokenizeLine(te, linesProcessedInBlock)

		if isAdded && a.haveStatistics {
			score := a.trans.getLastScore()
			scoreSqr := score * score
			blockScoreSum += score
			blockScoreSqrSum += scoreSqr
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
			a.outputFunc(a.name, a.rowID, a.rarityThreshold,
				score, scoreGap, scoreAvg, scoreStd, cnt, te)
		}
		if a.linesInBlock > 0 && linesProcessedInBlock >= a.linesInBlock && blockID >= 0 {
			break
		}
	}
	if err := a.fp.err(); err == io.EOF {
		if a.linesInBlock > 0 && linesProcessedInBlock < a.linesInBlock {
			return false, nil, nil
		}
	} else if err != nil {
		return false, nil, err
	}
	b := new(block)
	b.blockID = blockID
	b.scoreSqrSum = blockScoreSqrSum
	b.scoreSum = blockScoreSum

	return true, b, nil
}

func (a *rarityAnalyzer) saveLastStatus(blockID int) error {
	// save last status
	epoch := a.fp.currFileEpoch()
	cond := map[string]string{}
	row := map[string]string{
		"lastRowID":       fmt.Sprint(a.rowID),
		"lastBlockID":     fmt.Sprint(a.currBlockID),
		"fileName":        a.fp.currFile(),
		"lastRow":         fmt.Sprint(a.fp.row()),
		"modifiedEpoch":   fmt.Sprint(epoch),
		"modifiedUtcTime": fmt.Sprint(epochToString(epoch)),
	}
	if err := a.db.tables["lastStatus"].update(
		cond, row, "", true); err != nil {
		return err
	}
	return nil
}

func (a *rarityAnalyzer) saveBlock(b *block) error {
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

func (a *rarityAnalyzer) destroy() error {
	return a.db.dropAllTables()
}
