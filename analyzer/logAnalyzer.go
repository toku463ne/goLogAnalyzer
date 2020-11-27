package analyzer

import (
	"fmt"
	"math"
	"strings"

	"github.com/go-ini/ini"
)

// mainly for test
type logAnalyzerVars struct {
	name               string
	useDB              bool
	filterRe           string
	xFilterRe          string
	rootDir            string
	logPathRegex       string
	rarityThreshold    float64
	frequencyThreshold float64
	absencyThreshold   float64
	linesInBlock       int
	maxBlocks          int
	frequencyCheck     bool
	minSupportPerBlock float64
}

type logAnalyzer struct {
	rarAnal *rarityAnalyzer
	frqAnal *rarityAnalyzer
	absAnal *rarityAnalyzer

	name               string
	useDB              bool
	filterRe           string
	xFilterRe          string
	rootDir            string
	logPathRegex       string
	rarityThreshold    float64
	frequencyThreshold float64
	absencyThreshold   float64
	linesInBlock       int
	maxBlocks          int

	frequencyCheck     bool
	minSupportPerBlock float64
	freqClosedItems    map[string]string

	dciDB *csvDB
}

func newLogAnalyzer() *logAnalyzer {
	a := new(logAnalyzer)
	a.name = "loganal"
	//a.logPathRegex = v.logPathRegex
	a.useDB = true
	a.rootDir = fmt.Sprintf("./%s", a.name)
	a.rarityThreshold = 0.8
	a.absencyThreshold = 0.7
	a.linesInBlock = 10000
	a.maxBlocks = 1000
	a.frequencyCheck = true
	a.frequencyThreshold = 0.5
	a.minSupportPerBlock = 0.1
	a.freqClosedItems = make(map[string]string, 1000)

	return a
}

func newLogAnalyzerByVars(v *logAnalyzerVars) (*logAnalyzer, error) {
	a := newLogAnalyzer()
	a.name = v.name
	a.useDB = v.useDB
	a.filterRe = v.filterRe
	a.xFilterRe = v.xFilterRe
	a.rootDir = v.rootDir
	a.logPathRegex = v.logPathRegex
	a.rarityThreshold = v.rarityThreshold
	a.frequencyThreshold = v.frequencyThreshold
	a.linesInBlock = v.linesInBlock
	a.maxBlocks = v.maxBlocks

	if err := a.init(); err != nil {
		return nil, err
	}
	return a, nil
}

func newLogAnalyzerByIni(iniFile string) (*logAnalyzer, error) {
	a := newLogAnalyzer()
	if err := a.loadIni(iniFile); err != nil {
		return nil, err
	}
	err := a.init()
	return a, err
}

func (a *logAnalyzer) init() error {
	ra := newRarityAnalyzer()
	ra.name = fmt.Sprintf("%s.rare", a.name)
	ra.useDB = a.useDB
	ra.filterRe = a.filterRe
	ra.xFilterRe = a.xFilterRe
	ra.rootDir = fmt.Sprintf("%s/rarityDB", a.rootDir)
	ra.logPathRegex = a.logPathRegex
	ra.rarityThreshold = a.rarityThreshold
	ra.linesInBlock = a.linesInBlock
	ra.maxBlocks = a.maxBlocks
	a.rarAnal = ra

	if err := ra.init(); err != nil {
		return err
	}

	if a.frequencyCheck {
		// collect frequency
		fa := newRarityAnalyzer()
		fa.name = fmt.Sprintf("%s.freq", a.name)
		fa.useDB = true
		fa.rootDir = fmt.Sprintf("%s/freqDB", a.rootDir)
		fa.logPathRegex = ""
		//da.rarityThreshold = a.frequencyThreshold //frequencyThreshold is not a mistake
		fa.linesInBlock = 0
		fa.maxBlocks = a.maxBlocks
		fa.haveStatistics = false
		fa.outputFunc = outputScoreFunc(func(name string, rowID int64,
			scoreThreshold float64,
			score, scoreGap, scoreAvg, scoreStd float64,
			cnt int,
			text string) {
			return
		})
		if err := fa.init(); err != nil {
			return err
		}
		a.frqAnal = fa
		db, err := getDCIClosedDB(fmt.Sprintf("%s/dciDB", a.rootDir), a.maxBlocks)
		if err != nil {
			return err
		}
		a.dciDB = db

		// collect absence
		aa := newRarityAnalyzer()
		aa.name = fmt.Sprintf("%s.absn", a.name)
		aa.useDB = true
		aa.rootDir = fmt.Sprintf("%s/absentDB", a.rootDir)
		aa.logPathRegex = ""
		aa.rarityThreshold = a.absencyThreshold //absencyThreshold is not a mistake
		aa.linesInBlock = 0
		aa.maxBlocks = a.maxBlocks
		aa.haveStatistics = true
		aa.outputFunc = outputScoreFunc(func(name string, rowID int64,
			scoreThreshold float64,
			score, scoreGap, scoreAvg, scoreStd float64,
			cnt int,
			text string) {
			text = strings.Replace(text, ",", "", -1)
			text2 := a.freqClosedItems[text]
			if verbose || scoreGap > scoreThreshold {
				msg := fmt.Sprintf("%s %5d s=%5.2f g=%5.2f a=%5.2f d=%5.2f c=%5d | %s",
					name,
					rowID,
					score,
					scoreGap,
					scoreAvg,
					scoreStd,
					cnt,
					text2,
				)
				println(msg)
			}
		})
		if err := aa.init(); err != nil {
			return err
		}
		a.absAnal = aa

	}
	return nil
}

func (a *logAnalyzer) loadIni(iniFile string) error {
	cfg, err := ini.Load(iniFile)
	if err != nil {
		return err
	}
	for _, k := range cfg.Section("LogFile").Keys() {
		switch k.Name() {
		case "logName":
			a.name = k.MustString(a.name)
		case "rootDir":
			a.rootDir = k.MustString(a.rootDir)
		case "logPathRegex":
			a.logPathRegex = k.MustString(a.logPathRegex)
		case "linesInBlock":
			a.linesInBlock = k.MustInt(a.linesInBlock)
		case "maxBlocks":
			a.maxBlocks = k.MustInt(a.maxBlocks)
		case "rarityThreshold":
			a.rarityThreshold = k.MustFloat64(a.rarityThreshold)
		case "absencyThreshold":
			a.absencyThreshold = k.MustFloat64(a.absencyThreshold)
		case "frequencyCheck":
			a.frequencyCheck = k.MustBool(a.frequencyCheck)
		case "minSupportPerBlock":
			a.minSupportPerBlock = k.MustFloat64(a.minSupportPerBlock)
		}
	}
	return nil
}

func (a *logAnalyzer) loadDB() error {
	if err := a.rarAnal.loadDB(); err != nil {
		return err
	}
	if err := a.frqAnal.loadDB(); err != nil {
		return err
	}
	if err := a.absAnal.loadDB(); err != nil {
		return err
	}
	if err := a.loadDBFreqItemSets(); err != nil {
		return err
	}
	return nil
}

func (a *logAnalyzer) loadDBFreqItemSets() error {
	rows, err := a.dciDB.tables["frequentItemSets"].query(nil, "*")
	if err != nil {
		return err
	}
	if rows == nil || len(rows) == 0 {
		return nil
	}
	for _, v := range rows {
		a.freqClosedItems[v[0]] = v[1]
	}
	return nil
}

func (a *logAnalyzer) close() {
	a.frqAnal.close()
	a.rarAnal.close()
	a.absAnal.close()
}

func (a *logAnalyzer) destroy() error {
	if err := a.frqAnal.destroy(); err != nil {
		return err
	}
	if err := a.rarAnal.destroy(); err != nil {
		return err
	}
	if err := a.dciDB.dropAllTables(); err != nil {
		return err
	}
	if err := a.absAnal.destroy(); err != nil {
		return err
	}

	return nil
}

func (a *logAnalyzer) run() error {
	ra := a.rarAnal
	targetLinesCnt, err := ra.countTargetLines()
	if err != nil {
		return err
	}
	ra.targetLinesCnt = targetLinesCnt
	if err := ra.openFP(ra.logPathRegex, ra.lastFileEpoch, ra.lastFileRow); err != nil {
		return err
	}
	for {
		ra.currBlockID = ra.nextBlockID(ra.currBlockID)
		ok, err := ra.runBlock(ra.currBlockID)
		if err != nil {
			return err
		}
		if ok == false {
			break
		}
		blockID := ra.currBlockID
		if err := a.dropPartitions(blockID); err != nil {
			return err
		}
		if err := a.collectFrequency(blockID, ra.trans, ra.items); err != nil {
			return err
		}

		if err := a.collectAbsence(blockID); err != nil {
			return err
		}

		if ra.rowNum+ra.linesInBlock > ra.targetLinesCnt {
			break
		}
	}
	return nil
}

func (a *logAnalyzer) dropPartitions(blockID int) error {
	blockIDstr := fmt.Sprint(blockID)
	if err := a.dciDB.tables["frequentItemSets"].dropPartition(blockIDstr); err != nil {
		return err
	}
	if err := a.dciDB.tables["frequentItemFirstLines"].dropPartition(blockIDstr); err != nil {
		return err
	}
	if err := a.dciDB.tables["frequentItemLastLines"].dropPartition(blockIDstr); err != nil {
		return err
	}
	if err := a.dciDB.tables["frequentItemSetsDotted"].dropPartition(blockIDstr); err != nil {
		return err
	}
	if err := a.dciDB.tables["frequentItemSetsAbsent"].dropPartition(blockIDstr); err != nil {
		return err
	}
	return nil
}

func (a *logAnalyzer) getFreqItemString(fi []string) string {
	return strings.Join(fi, " ")
}

func (a *logAnalyzer) getFreqItemKey(fi []string) string {
	return strings.Join(fi, ".")
}

func (a *logAnalyzer) collectFrequency(blockID int, trans1 *trans, items1 *items) error {
	blockIDstr := fmt.Sprint(blockID)
	matrix := tran2BitMatrix(trans1, items1)
	minSup := int(math.Round(float64(a.linesInBlock) * a.minSupportPerBlock))
	if minSup == 0 {
		minSup = 1
	}
	dci, err := newDCIClosed(matrix, minSup, true)
	if err != nil {
		return err
	}

	err = dci.run()
	if err != nil {
		return err
	}

	closedSets, supps, firstTIDs, lastTIDs := dci.getClosedWordsSorted(items1)
	rows1 := make([][]string, len(supps))
	rows2 := make([][]string, len(supps))
	rows3 := make([][]string, len(supps))
	rows4 := make([][]string, len(supps))
	table := a.dciDB.tables["frequentItemSets"]
	cols := table.colMap
	for i, cs := range closedSets {
		tev := a.getFreqItemString(cs)
		tek := a.getFreqItemKey(cs)
		row1 := make([]string, 3)
		row2 := make([]string, 1)
		row3 := make([]string, 1)
		row4 := make([]string, 1)
		row1[cols["support"]] = fmt.Sprint(supps[i])
		row1[cols["key"]] = tek
		row1[cols["itemSets"]] = tev
		row2[cols["line"]] = trans1.doc.get(firstTIDs[i])
		row3[cols["line"]] = trans1.doc.get(lastTIDs[i])
		row4[cols["line"]] = tek
		a.freqClosedItems[tek] = tev
		rows1[i] = row1
		rows2[i] = row2
		rows3[i] = row3
		rows4[i] = row4
	}
	if err := a.dciDB.tables["frequentItemSets"].insertRows(rows1,
		blockIDstr); err != nil {
		return err
	}
	if err := a.dciDB.tables["frequentItemFirstLines"].insertRows(rows2,
		blockIDstr); err != nil {
		return err
	}
	if err := a.dciDB.tables["frequentItemLastLines"].insertRows(rows3,
		blockIDstr); err != nil {
		return err
	}
	if err := a.dciDB.tables["frequentItemSetsDotted"].insertRows(rows4,
		blockIDstr); err != nil {
		return err
	}

	fa := a.frqAnal
	freqPath := a.dciDB.tables["frequentItemSetsDotted"].getPath(fmt.Sprint(blockID))
	if err := fa.openFP(freqPath, 0, 0); err != nil {
		return err
	}
	fa.currBlockID = blockID
	if _, err := fa.runBlock(blockID); err != nil {
		return err
	}
	return nil
}

func (a *logAnalyzer) collectAbsence(blockID int) error {
	fa := a.frqAnal
	// save item
	items := fa.items
	rows := [][]string{}
	table := a.dciDB.tables["frequentItemSetsAbsent"]
	cols := table.colMap
	for itemID, cnt := range items.newCounts.getSlice() {
		if cnt == 0 {
			row := make([]string, len(cols))
			row[cols["line"]] = string(items.items.get(itemID))
			rows = append(rows, row)
		}
	}
	if err := a.dciDB.tables["frequentItemSetsAbsent"].insertRows(rows,
		fmt.Sprint(blockID)); err != nil {
		return err
	}
	aa := a.absAnal
	absenPath := a.dciDB.tables["frequentItemSetsAbsent"].getPath(fmt.Sprint(blockID))
	if err := aa.openFP(absenPath, 0, 0); err != nil {
		return err
	}
	aa.currBlockID = blockID
	if _, err := aa.runBlock(blockID); err != nil {
		return err
	}

	return nil

}
