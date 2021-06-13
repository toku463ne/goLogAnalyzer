package analyzer

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/go-ini/ini"
	"github.com/pkg/errors"
)

type rarityAnalyzer struct {
	// fields
	name    string
	rowID   int64
	rootDir string
	useDB   bool
	db      *csvDB
	trans   *trans
	blocks  []*block

	scoreCount  int
	scoreSum    float64
	scoreSqrSum float64
	countTotal  int

	currBlock    *block
	currBlockID  int
	linesInBlock int
	maxBlocks    int

	lastFileEpoch int64
	lastFileRow   int

	logPathRegex string
	filterRe     string
	xFilterRe    string

	countPerScore     []int
	logRecordsBuff    [][]string
	logRecordsBuffPos int
	recordsToShow     int
	nTopRareLogs      []*logRec
	minTopRareScore   float64
	minGapToRecord    float64

	fp *filePointer

	outputRes func(rowID int64, score, scoreGap float64, text string) error
}

func newRarityAnalyzer(logPathRegex,
	rootDir string,
	filterRe, xFilterRe string,
	minGapToRecord float64,
	linesInBlock, maxBlocks, recordsToShow int) (*rarityAnalyzer, error) {

	a := new(rarityAnalyzer)
	if err := a.init(logPathRegex,
		rootDir,
		filterRe, xFilterRe,
		minGapToRecord,
		linesInBlock, maxBlocks, recordsToShow); err != nil {
		return nil, err
	}

	a.outputRes = func(rowID int64, score, scoreGap float64, text string) error {
		if verbose || scoreGap >= a.minGapToRecord {

			if a.useDB {
				a.logRecordsBuffPos++
				a.logRecordsBuff[a.logRecordsBuffPos] = []string{
					fmt.Sprint(rowID),
					fmt.Sprintf("%3.2f", score),
					text,
				}
			}
		}
		return nil
	}

	return a, nil
}

func loadAnalyzer(rootDir string) (*rarityAnalyzer, error) {
	a, err := newRarityAnalyzer("",
		rootDir,
		"", "",
		0,
		-1, -1, 0)
	if err != nil {
		return nil, err
	}

	if err := a.loadDB(); err != nil {
		return nil, err
	}
	return a, err
}

func (a *rarityAnalyzer) init(logPathRegex,
	rootDir string,
	filterRe, xFilterRe string,
	minGapToRecord float64,
	linesInBlock, maxBlocks, recordsToShow int) error {

	a.useDB = false
	if rootDir != "" {
		a.rootDir = rootDir
		a.name = filepath.Base(rootDir)
		a.useDB = true
	}

	// default values
	a.linesInBlock = cDefaultBlockSize
	a.maxBlocks = cDefaultMaxBlocks
	a.recordsToShow = cNTopRareRecords
	a.minGapToRecord = cMinGapToRecord

	if a.useDB {
		if err := a.loadIni(); err != nil {
			return err
		}
	}

	if logPathRegex != "" {
		a.logPathRegex = logPathRegex
	}
	if filterRe != "" {
		a.filterRe = filterRe
	}
	if xFilterRe != "" {
		a.xFilterRe = xFilterRe
	}

	if linesInBlock > 0 {
		a.linesInBlock = linesInBlock
	}
	if maxBlocks > 0 {
		a.maxBlocks = maxBlocks
	}
	if recordsToShow > 0 {
		a.recordsToShow = recordsToShow
	}

	if minGapToRecord > 0 {
		a.minGapToRecord = minGapToRecord
	}

	maxBlockNum := int(math.Pow10(cMaxBlockDitigs))
	if a.maxBlocks >= maxBlockNum {
		return fmt.Errorf("maxBlocks can't be %d or bigger", maxBlockNum)
	}

	a.blocks = make([]*block, a.maxBlocks)
	a.logRecordsBuff = make([][]string, a.linesInBlock)
	a.logRecordsBuffPos = -1
	a.countPerScore = make([]int, cCountbyScoreLen)
	a.nTopRareLogs = make([]*logRec, a.recordsToShow)

	if a.useDB {
		db, err := getRarityAnalDB(a.rootDir, a.maxBlocks)
		if err != nil {
			return err
		}
		a.db = db
	}

	a.currBlockID = -1
	a.trans = newTrans(false)

	return nil
}

func (a *rarityAnalyzer) getIniPath() string {
	return fmt.Sprintf("%s/cfg.ini", a.rootDir)
}

func (a *rarityAnalyzer) loadIni() error {
	iniFile := a.getIniPath()
	if !PathExist(iniFile) {
		return nil
	}

	cfg, err := ini.Load(iniFile)
	if err != nil {
		return err
	}
	for _, k := range cfg.Section("LogFile").Keys() {
		switch k.Name() {
		case "logPathRegex":
			a.logPathRegex = k.MustString(a.logPathRegex)

		case "linesInBlock":
			a.linesInBlock = k.MustInt(a.linesInBlock)

		case "maxBlocks":
			a.maxBlocks = k.MustInt(a.maxBlocks)

		case "filterRe":
			a.filterRe = k.MustString(a.filterRe)

		case "xFilterRe":
			a.xFilterRe = k.MustString(a.xFilterRe)
		}
	}
	return nil
}

func (a *rarityAnalyzer) saveIni() error {
	iniFile := a.getIniPath()
	if !PathExist(iniFile) {
		file, err := os.OpenFile(iniFile, os.O_CREATE, 0640)
		if err != nil {
			return err
		}
		defer file.Close()
	}
	cfg, err := ini.Load(iniFile)
	if err != nil {
		return err
	}
	cfg.Section("LogFile").Key("logPathRegex").SetValue(a.logPathRegex)
	cfg.Section("LogFile").Key("linesInBlock").SetValue(fmt.Sprint(a.linesInBlock))
	cfg.Section("LogFile").Key("maxBlocks").SetValue(fmt.Sprint(a.maxBlocks))
	cfg.Section("LogFile").Key("filterRe").SetValue(a.filterRe)
	cfg.Section("LogFile").Key("xFilterRe").SetValue(a.xFilterRe)

	return cfg.SaveTo(iniFile)
}

func (a *rarityAnalyzer) loadDB() error {
	lastBlockID := -1
	var lastEpoch int64
	var lastRow int

	if a.useDB == false {
		return nil
	}

	tmpBlockID, tmpRow, tmpEpoch, err := a.loadDBLastStatus()
	if err != nil {
		return err
	}
	lastRow = tmpRow
	lastEpoch = tmpEpoch
	lastBlockID = tmpBlockID

	if err := a.loadDBBlocks(); err != nil {
		return err
	}

	if err := a.loadDBItems(); err != nil {
		return err
	}

	if err := a.loadLogRecords(); err != nil {
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
	if a.fp != nil {
		a.fp.close()
	}
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
	for blockID := range a.blocks {
		b, err := a.loadDBBlock(blockID)
		if err != nil {
			return err
		}
		a.blocks[blockID] = b
	}
	b, err := a.loadDBBlock(cLastTmpBlockID)
	if err != nil {
		return err
	}
	if b != nil {
		a.currBlock = b
		a.currBlockID = b.blockID
		//a.linesProcessedInBlock = b.
	}
	return nil
}

func (a *rarityAnalyzer) loadDBBlock(blockID int) (*block, error) {
	b := newBlock(blockID)
	blockIDstr := a.blockID2Str(blockID)

	rows, err := a.db.tables["logBlocks"].query(nil, blockIDstr)
	if err != nil {
		return nil, err
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
			return nil, errors.WithStack(err)
		}

		blockCnt, err := strconv.Atoi(v[idxBlockCnt])
		if err != nil {
			return nil, errors.WithStack(err)
		}
		scoreSum, err := strconv.ParseFloat(v[idxScoreSum], 64)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		scoreSqrSum, err := strconv.ParseFloat(v[idxScoreSqrSum], 64)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		tmpEpoch1, err := strconv.Atoi(v[idxLastEpoch])
		if err != nil {
			return nil, errors.WithStack(err)
		}
		tmpEpoch2 := int64(tmpEpoch1)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		b.blockID = blockID
		b.blockCnt = blockCnt
		b.lastEpoch = tmpEpoch2
		b.scoreSum = scoreSum
		b.scoreSqrSum = scoreSqrSum

		//a.rowNum += blockCnt
		a.scoreCount += blockCnt
		a.scoreSum += scoreSum
		a.scoreSqrSum += scoreSqrSum
		a.countTotal += blockCnt
	}

	rows, err = a.db.tables["countPerScore"].query(nil, blockIDstr)
	if err != nil {
		return nil, err
	}
	cols = a.db.tables["countPerScore"].colMap
	for _, v := range rows {
		countPerScore := make([]int, cCountbyScoreLen)
		for i := 0; i < cCountbyScoreLen; i++ {
			cnt, err := strconv.Atoi(v[i])
			if err != nil {
				break
			}
			countPerScore[i] = cnt
			a.countPerScore[i] += cnt
		}
		b.countPerScore = countPerScore
	}

	rows, err = a.db.tables["nTopRareLogs"].query(nil, blockIDstr)
	if err != nil {
		return nil, err
	}
	cols = a.db.tables["nTopRareLogs"].colMap
	idxRowID := cols["rowID"]
	idxScore := cols["score"]
	idxText := cols["text"]
	nTopRareLogs := make([]*logRec, a.recordsToShow)
	for i, v := range rows {
		if i >= a.recordsToShow {
			break
		}
		r := new(logRec)
		rowID, err := strconv.ParseInt(v[idxRowID], 10, 64)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		score, err := strconv.ParseFloat(v[idxScore], 64)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		text := v[idxText]
		r.rowID = rowID
		r.score = score
		r.text = text
		nTopRareLogs[i] = r
	}

	return b, nil
}

func (a *rarityAnalyzer) loadDBItems() error {
	var err error
	items1 := newItems()
	for blockID := range a.blocks {
		items1, err = a.loadDBItem(blockID, items1)
		if err != nil {
			return err
		}
	}
	items1, err = a.loadDBItem(cLastTmpBlockID, items1)
	if err != nil {
		return err
	}
	a.trans.items = items1
	return nil
}

func (a *rarityAnalyzer) loadDBItem(blockID int, items1 *items) (*items, error) {
	blockIDstr := a.blockID2Str(blockID)
	isNewItem := false
	if blockID == cLastTmpBlockID {
		isNewItem = true
	}

	rows, err := a.db.tables["items"].query(nil, blockIDstr)
	if err != nil {
		return items1, err
	}
	if rows == nil || len(rows) == 0 {
		return items1, nil
	}
	for _, v := range rows {
		cnt, err := strconv.Atoi(v[1])
		if err != nil {
			return items1, errors.WithStack(err)
		}
		items1.register(v[0], cnt, isNewItem)
	}
	return items1, nil
}

func (a *rarityAnalyzer) loadLogRecords() error {
	rows, err := a.db.tables["logRecords"].query(nil, cLastTmpBlockStr)
	if err != nil {
		return err
	}
	for i, v := range rows {
		buff := make([]string, len(v))
		for j, w := range v {
			buff[j] = w
		}
		a.logRecordsBuff[i] = buff
		a.logRecordsBuffPos = i
	}
	return nil
}

func (a *rarityAnalyzer) deleteItemBlock(blockIDstr string) error {
	rows, err := a.db.tables["items"].query(nil, blockIDstr)
	if err != nil {
		return err
	}
	if rows == nil || len(rows) == 0 {
		return nil
	}
	for _, v := range rows {
		cnt, err := strconv.Atoi(v[1])
		if err != nil {
			return errors.WithStack(err)
		}
		a.trans.items.subCount(v[0], cnt)
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
	for i := 0; i < cCountbyScoreLen; i++ {
		cnt := b.countPerScore[i]
		a.countPerScore[i] -= cnt
	}
	a.countTotal -= b.blockCnt

	if a.useDB {
		a.deleteItemBlock(blockIDstr)
		a.db.tables["logBlocks"].dropPartition(blockIDstr)
		a.db.tables["items"].dropPartition(blockIDstr)
		a.db.tables["logRecords"].dropPartition(blockIDstr)
		a.db.tables["countPerScore"].dropPartition(blockIDstr)
		a.db.tables["nTopRareLogs"].dropPartition(blockIDstr)
	}

	a.blocks[blockID] = nil
	return nil
}

func (a *rarityAnalyzer) blockID2Str(blockID int) string {
	if blockID == cLastTmpBlockID {
		return cLastTmpBlockStr
	}
	return fmt.Sprint(blockID)
}

func (a *rarityAnalyzer) saveLastStatus() error {
	// save last status
	blockID := a.currBlockID
	epoch := a.fp.currFileEpoch()
	cond := map[string]string{}
	row := map[string]string{
		"lastRowID":       fmt.Sprint(a.rowID),
		"lastBlockID":     a.blockID2Str(blockID),
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

func (a *rarityAnalyzer) saveBlock(blockID int) error {
	b := a.currBlock
	if b == nil {
		return nil
	}
	blockIDstr := a.blockID2Str(blockID)
	if blockID < 0 {
		a.db.tables["logBlocks"].dropPartition(blockIDstr)
		//a.db.tables["countPerScore"].dropPartition(blockIDstr)
	}

	//blockID := b.blockID
	// save block
	cond := map[string]string{
		"blockID": fmt.Sprint(a.currBlockID),
	}
	completed := "1"
	if b.completed == false {
		completed = "0"
	}
	row := map[string]string{
		"lastRowID":   fmt.Sprint(a.rowID),
		"blockCnt":    fmt.Sprint(b.blockCnt),
		"scoreSum":    fmt.Sprint(b.scoreSum),
		"scoreSqrSum": fmt.Sprint(b.scoreSqrSum),
		"lastEpoch":   fmt.Sprint(a.fp.currFileEpoch()),
		"createdAt":   timeToString(time.Now()),
		"completed":   completed,
	}
	if err := a.db.tables["logBlocks"].update(
		cond, row, blockIDstr, true); err != nil {
		return err
	}
	return nil
}

func (a *rarityAnalyzer) saveCountPerScore(blockID int) error {
	b := a.currBlock
	if b == nil {
		return nil
	}
	blockIDstr := a.blockID2Str(blockID)
	row := make([]string, cCountbyScoreLen)
	for i := 0; i < cCountbyScoreLen; i++ {
		row[i] = fmt.Sprint(b.countPerScore[i])
	}
	if err := a.db.tables["countPerScore"].insertRows(
		[][]string{row}, blockIDstr, 0); err != nil {
		return err
	}
	return nil
}

func (a *rarityAnalyzer) saveItems(blockID int) error {
	// save item
	//blockID := a.currBlockID
	blockIDstr := a.blockID2Str(blockID)
	items := a.trans.items
	rows := [][]string{}
	table := a.db.tables["items"]
	cols := table.colMap
	for itemID, cnt := range items.currCounts {
		if cnt > 0 {
			row := make([]string, len(cols))
			row[cols["word"]] = items.getWord(itemID)
			row[cols["cnt"]] = fmt.Sprint(cnt)
			rows = append(rows, row)
		}
	}
	if blockID < 0 {
		table.dropPartition(blockIDstr)
	}

	if err := table.insertRows(rows, blockIDstr, 0); err != nil {
		return err
	}
	return nil
}

func (a *rarityAnalyzer) saveLogRecords(blockID int) error {
	blockIDstr := a.blockID2Str(blockID)

	table := a.db.tables["logRecords"]
	if blockID < 0 {
		table.dropPartition(blockIDstr)
	}
	if a.logRecordsBuffPos >= 0 {
		if err := table.insertRows(a.logRecordsBuff,
			blockIDstr,
			a.logRecordsBuffPos+1); err != nil {
			return err
		}
	}

	return nil
}

func (a *rarityAnalyzer) clean() error {
	return a.db.dropAllTables()
}

func (a *rarityAnalyzer) initCurrBlock(blockID int) {
	a.currBlock = newBlock(blockID)
	nTopRareLogs := make([]*logRec, a.recordsToShow)
	for i := range nTopRareLogs {
		nTopRareLogs[i] = new(logRec)
	}
}

func (a *rarityAnalyzer) nextBlock() {
	if a.currBlock != nil {
		if curLogLevel == cLogLevelDebug {
			a.printCountPerScore(a.currBlock.countPerScore,
				fmt.Sprintf("Finished blockID=%d\ncounts per score",
					a.currBlockID))
		}
	}
	a.logRecordsBuff = make([][]string, a.linesInBlock)
	a.logRecordsBuffPos = -1

	a.currBlockID++
	if a.currBlockID-a.maxBlocks >= 0 {
		a.currBlockID = 0
	}
	a.initCurrBlock(a.currBlockID)
	if curLogLevel == cLogLevelDebug {
		fmt.Printf("\nStarting blockID=%d\n", a.currBlockID)
	}
	a.trans.items.clearCurrCount()
}

func (a *rarityAnalyzer) postBlock(blockID int) error {
	//blockID := a.currBlockID
	if (blockID == cLastTmpBlockID || blockID >= 0) && a.currBlock != nil {
		if err := a.deleteOld(blockID); err != nil {
			return err
		}

		if a.useDB {
			if err := a.saveLogRecords(blockID); err != nil {
				return err
			}
			if err := a.saveCountPerScore(blockID); err != nil {
				return err
			}

			if err := a.saveLastStatus(); err != nil {
				return err
			}
			if err := a.saveBlock(blockID); err != nil {
				return err
			}

			if err := a.saveItems(blockID); err != nil {
				return err
			}
			if err := a.saveIni(); err != nil {
				return err
			}
		}
		if blockID != cLastTmpBlockID {
			a.blocks[a.currBlockID] = a.currBlock
		}
	}
	return nil
}

func (a *rarityAnalyzer) printCountPerScore(g []int, msg string) {
	//fmt.Printf("\n")
	fmt.Printf("%s\n", msg)
	fmt.Printf(" score | count\n")
	fmt.Printf(" ------+--------------\n")
	for i := 0; i < cCountbyScoreLen; i++ {
		if g[i] > 0 {
			fmt.Printf("  %4.1f | %d\n", float64(i), g[i])
		}
	}
}

func (a *rarityAnalyzer) scanAndGetNTops(recordsToShow int,
	filterRe, xFilterRe string, blockIDstrs []string,
	topNMaxScore float64) ([]*logRec, error) {

	if recordsToShow == 0 {
		recordsToShow = a.recordsToShow
	}

	nTopRareLogs := make([]*logRec, recordsToShow)
	minTopRareScore := 0.0
	var rows [][]string
	var err error
	if blockIDstrs == nil {
		rows, err = a.db.tables["logRecords"].query(nil, "*")
		if err != nil {
			return nil, err
		}
	} else {
		for _, blockIDstr := range blockIDstrs {
			if blockIDstr == "" {
				break
			}
			rows1, err := a.db.tables["logRecords"].query(nil, blockIDstr)
			if err != nil {
				return nil, err
			}
			rows = append(rows, rows1...)
		}
	}
	cols := a.db.tables["logRecords"].colMap
	idxRowID := cols["rowID"]
	idxScore := cols["score"]
	idxText := cols["text"]

	for _, row := range rows {
		te := row[idxText]
		if filterRe != "" && searchReg(te, filterRe) == false {
			continue
		}
		if xFilterRe != "" && searchReg(te, xFilterRe) {
			continue
		}

		rowID, err := strconv.ParseInt(row[idxRowID], 10, 64)
		if err != nil {
			return nil, err
		}
		score, err := strconv.ParseFloat(row[idxScore], 64)
		if err != nil {
			return nil, err
		}

		if topNMaxScore > 0 && score > topNMaxScore {
			continue
		}

		nTopRareLogs, minTopRareScore = registerNTopRareRec(nTopRareLogs,
			minTopRareScore, rowID, score, te)
	}
	return nTopRareLogs, nil
}

func (a *rarityAnalyzer) getBlockIDsFromEpoch(startEpoch, endEpoch int64) ([]string, error) {
	rows, err := a.db.tables["logBlocks"].query(nil, "*")
	if err != nil {
		return nil, err
	}
	blockIDstrs := make([]string, len(rows))

	cols := a.db.tables["logBlocks"].colMap
	idxBlockID := cols["blockID"]
	idxLastEpoch := cols["lastEpoch"]
	idxCompleted := cols["completed"]

	i := 0
	for _, v := range rows {
		blockIDstr := v[idxBlockID]
		lastEpoch, err := strconv.ParseInt(v[idxLastEpoch], 10, 64)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		completed := v[idxCompleted]
		if completed == "0" {
			blockIDstr = cLastTmpBlockStr
		}
		if startEpoch > 0 && startEpoch <= lastEpoch {
			if endEpoch > 0 {
				if endEpoch >= lastEpoch {
					blockIDstrs[i] = blockIDstr
					i++
				}
			} else {
				blockIDstrs[i] = blockIDstr
				i++
			}
		} else if startEpoch <= 0 && endEpoch > 0 && endEpoch >= lastEpoch {
			blockIDstrs[i] = blockIDstr
			i++
		}
	}
	return blockIDstrs, nil
}

func (a *rarityAnalyzer) printNTops(msg string,
	recordsToShow int,
	filterRe, xFilterRe string, blockIDstrs []string,
	topNMaxScore float64,
) error {
	var err error
	var nTopRareLogs []*logRec
	if recordsToShow > 0 && recordsToShow != a.recordsToShow || a.nTopRareLogs == nil || a.nTopRareLogs[0] == nil {
		if recordsToShow == 0 {
			recordsToShow = cNTopRareRecords
		}
		nTopRareLogs, err = a.scanAndGetNTops(recordsToShow, filterRe, xFilterRe, blockIDstrs, topNMaxScore)
		if err != nil {
			return err
		}
	} else {
		nTopRareLogs = a.nTopRareLogs
	}

	fmt.Printf("%s\n", msg)
	fmt.Print("score   rowID      text\n")
	fmt.Print("-------+----------+-------\n")
	for i, logr := range nTopRareLogs {
		if logr == nil {
			break
		}
		fmt.Printf(" %5.2f   %8d   %s\n", logr.score, logr.rowID, logr.text)
		if logr.score == 0 {
			break
		}
		if i+1 >= recordsToShow {
			break
		}
	}
	return nil
}

func getScoreStage(score float64) int {
	scoreStage := int(math.Floor(score))
	if scoreStage < 0 {
		scoreStage = 0
	}
	if scoreStage >= cCountbyScoreLen {
		scoreStage = cCountbyScoreLen - 1
	}
	return scoreStage
}

func (a *rarityAnalyzer) registerScore(score float64) {
	scoreStage := getScoreStage(score)

	a.countPerScore[scoreStage]++
	a.currBlock.countPerScore[scoreStage]++
}

func (a *rarityAnalyzer) run(targetLinesCnt int) (int, error) {
	linesProcessed := 0

	if a.currBlockID < 0 {
		a.nextBlock()
	}

	if a.fp == nil || !a.fp.isOpen() {
		a.fp = newFilePointer(a.logPathRegex, a.lastFileEpoch, a.lastFileRow)
		if err := a.fp.open(); err != nil {
			return 0, err
		}
	}
	for a.fp.next() {
		if a.currBlock == nil || a.currBlock.completed {
			a.nextBlock()
		}

		a.rowID++
		if a.rowID > cMaxRowID {
			a.rowID = 0
		}

		te := a.fp.text()
		if te == "" {
			continue
		}

		tran := a.trans.tokenizeLineLight(te, a.filterRe, a.xFilterRe)
		if len(tran) == 0 {
			continue
		}
		score := a.trans.calcScore(tran)
		scoreSqr := score * score
		a.currBlock.scoreSum += score
		a.currBlock.scoreSqrSum += scoreSqr
		a.currBlock.blockCnt++
		a.scoreSum += score
		a.scoreSqrSum += scoreSqr
		a.scoreCount++
		cnt := float64(a.scoreCount)
		//score avg
		sa := a.scoreSum / cnt
		//score std
		ss := math.Sqrt((a.scoreSqrSum - 2*a.scoreSum*sa + cnt*sa*sa) / cnt)

		var scoreGap float64
		if ss > 0 {
			scoreGap = (score - sa) / (ss)
		}
		a.registerScore(score)

		a.countTotal++
		linesProcessed++

		if err := a.outputRes(a.rowID, score, scoreGap, te); err != nil {
			return linesProcessed, err
		}
		a.nTopRareLogs, a.minTopRareScore = registerNTopRareRec(a.nTopRareLogs,
			a.minTopRareScore, a.rowID, score, te)

		if a.linesInBlock > 0 && (a.currBlock.blockCnt >= a.linesInBlock || (a.fp.isEOF && !a.fp.isLastFile())) {
			a.currBlock.completed = true
			if err := a.postBlock(a.currBlockID); err != nil {
				return linesProcessed, err
			}

		}

		if targetLinesCnt > 0 && linesProcessed >= targetLinesCnt {
			break
		}
	}

	if a.currBlock.blockCnt > 0 && !a.currBlock.completed {
		if err := a.postBlock(cLastTmpBlockID); err != nil {
			return linesProcessed, err
		}
	}
	return linesProcessed, nil
}

func (a *rarityAnalyzer) updateLogScore() error {
	//a.initCurrBlock(0)
	a.logRecordsBuff = make([][]string, a.linesInBlock)
	a.logRecordsBuffPos = -1
	//a.nextBlock()

	cols := a.db.tables["logRecords"].colMap
	idxRowID := cols["rowID"]
	idxText := cols["text"]
	idxScore := cols["score"]

	procBlock := func(blockID int) error {
		blockIDstr := a.blockID2Str(blockID)
		rows, err := a.db.tables["logRecords"].query(nil, blockIDstr)
		if err != nil {
			return err
		}
		if len(rows) == 0 {
			return nil
		}
		for _, v := range rows {
			te := v[idxText]
			rowID, _ := strconv.ParseInt(v[idxRowID], 10, 64)
			tran := a.trans.toTermListLight(te, false)
			oldscore, _ := strconv.ParseFloat(v[idxScore], 64)
			scoreStorage := getScoreStage(oldscore)
			a.countPerScore[scoreStorage]--
			a.blocks[blockID].countPerScore[scoreStorage]--
			score := a.trans.calcScore(tran)
			scoreStorage = getScoreStage(score)
			a.countPerScore[scoreStorage]++
			a.blocks[blockID].countPerScore[scoreStorage]++

			if err := a.outputRes(rowID, score, 1000, te); err != nil {
				return err
			}
		}

		a.currBlock = a.blocks[blockID]
		a.db.tables["logRecords"].dropPartition(blockIDstr)
		a.saveLogRecords(blockID)
		a.db.tables["countPerScore"].dropPartition(blockIDstr)
		a.saveCountPerScore(blockID)
		a.nextBlock()
		return nil
	}

	for i := 0; i < a.maxBlocks; i++ {
		if a.blocks[i] == nil {
			break
		}
		if err := procBlock(i); err != nil {
			return err
		}
	}
	return nil
}

// To test read speed
func (a *rarityAnalyzer) runOnlyRead(targetLinesCnt int) (int, error) {
	linesProcessed := 0

	if a.fp == nil || !a.fp.isOpen() {
		a.fp = newFilePointer(a.logPathRegex, a.lastFileEpoch, a.lastFileRow)
		if err := a.fp.open(); err != nil {
			return 0, err
		}
	}
	for a.fp.next() {
		linesProcessed++
		if targetLinesCnt > 0 && linesProcessed >= targetLinesCnt {
			break
		}
	}
	a.fp.close()

	return linesProcessed, nil
}

// To test read/tokenize speed
func (a *rarityAnalyzer) runReadTokenize(targetLinesCnt int) (int, error) {
	linesProcessed := 0

	if a.fp == nil || !a.fp.isOpen() {
		a.fp = newFilePointer(a.logPathRegex, a.lastFileEpoch, a.lastFileRow)
		if err := a.fp.open(); err != nil {
			return 0, err
		}
	}
	for a.fp.next() {
		te := a.fp.text()
		tran := a.trans.tokenizeLine(te, a.filterRe, a.xFilterRe)
		if tran == nil || len(tran) == 0 {
			continue
		}

		linesProcessed++
		if linesProcessed >= targetLinesCnt {
			break
		}
	}
	a.fp.close()

	return linesProcessed, nil
}

// To test read/tokenize speed
func (a *rarityAnalyzer) tokenizeLineNogeg(targetLinesCnt int) (int, error) {
	linesProcessed := 0

	if a.fp == nil || !a.fp.isOpen() {
		a.fp = newFilePointer(a.logPathRegex, a.lastFileEpoch, a.lastFileRow)
		if err := a.fp.open(); err != nil {
			return 0, err
		}
	}
	for a.fp.next() {
		te := a.fp.text()
		a.trans.tokenizeLineNogeg(te)

		linesProcessed++
		if linesProcessed >= targetLinesCnt {
			break
		}
	}
	a.fp.close()

	return linesProcessed, nil
}
