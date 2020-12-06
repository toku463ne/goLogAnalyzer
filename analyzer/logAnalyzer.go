package analyzer

import (
	"fmt"
	"math"
	"strings"

	"github.com/go-ini/ini"
)

type logAnalyzerVars struct {
	name               string
	useDB              bool
	filterRe           string
	xFilterRe          string
	rootDir            string
	logPathRegex       string
	rarityThreshold    float64
	absenceThreshold   float64
	linesInBlock       int
	maxBlocks          int
	absenceCheck       bool
	minSupportPerBlock float64
}

type closedItemSet struct {
	text     string
	lastLine string
}

type logAnalyzer struct {
	rarAnal *fileRarityAnalyzer
	frqAnal *splitRarityAnalyzer
	absAnal *itemAbsenceAnalyzer

	name             string
	useDB            bool
	filterRe         string
	xFilterRe        string
	rootDir          string
	logPathRegex     string
	rarityThreshold  float64
	absenceThreshold float64
	linesInBlock     int
	maxBlocks        int
	closedItemSets   map[string]closedItemSet

	absenceCheck       bool
	minSupportPerBlock float64

	dciDB *csvDB
}

func newLogAnalyzer() *logAnalyzer {
	a := new(logAnalyzer)
	a.name = "loganal"
	//a.logPathRegex = v.logPathRegex
	a.useDB = true
	a.rootDir = fmt.Sprintf("./%s", a.name)
	a.rarityThreshold = 0.8
	a.absenceThreshold = 0.7
	a.linesInBlock = 10000
	a.maxBlocks = 1000
	a.absenceCheck = true
	a.minSupportPerBlock = 0.1
	a.closedItemSets = make(map[string]closedItemSet, 1000)

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
	a.absenceThreshold = v.absenceThreshold
	a.linesInBlock = v.linesInBlock
	a.maxBlocks = v.maxBlocks
	a.absenceCheck = v.absenceCheck

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

func newLogAnalyzerByDefaults(pathRegex string) (*logAnalyzer, error) {
	tmp := strings.Split(pathRegex, "/")
	if len(tmp) <= 1 {
		tmp = strings.Split(pathRegex, "\\")
	}
	name := tmp[len(tmp)-1]

	a := newLogAnalyzer()
	a.name = name
	a.useDB = false
	a.filterRe = ""
	a.xFilterRe = ""
	a.rootDir = "."
	a.logPathRegex = pathRegex
	a.rarityThreshold = 0.8
	a.absenceThreshold = 0.0
	a.linesInBlock = 0
	a.maxBlocks = 100
	a.absenceCheck = false

	if err := a.init(); err != nil {
		return nil, err
	}
	return a, nil
}

func (a *logAnalyzer) init() error {
	ra := newFileRarityAnalyzer()
	ra.name = fmt.Sprintf("%s.rare", a.name)
	ra.useDB = a.useDB
	ra.filterRe = a.filterRe
	ra.xFilterRe = a.xFilterRe
	ra.rootDir = fmt.Sprintf("%s/rarityDB", a.rootDir)
	ra.logPathRegex = a.logPathRegex
	ra.rarityThreshold = a.rarityThreshold
	ra.linesInBlock = a.linesInBlock
	ra.maxBlocks = a.maxBlocks
	if err := ra.init(); err != nil {
		return err
	}
	a.rarAnal = ra

	if a.absenceCheck {
		a.useDB = true
		ra.useDB = true
		db, err := getClosedItemsDB(fmt.Sprintf("%s/closedItemsDB", a.rootDir), a.maxBlocks)
		if err != nil {
			return err
		}
		a.dciDB = db

		// collect frequency
		fa := newSplitRarityAnalyzer()
		fa.name = fmt.Sprintf("%s.freq", a.name)
		fa.useDB = true
		fa.rootDir = fmt.Sprintf("%s/frequencyDB", a.rootDir)
		fa.linesInBlock = 0
		fa.maxBlocks = a.maxBlocks
		fa.haveStatistics = false
		if err := fa.init(); err != nil {
			return err
		}
		a.frqAnal = fa

		aa := newItemAbsenceAnalyzer()
		aa.name = fmt.Sprintf("%s.absn", a.name)
		aa.useDB = true
		aa.rootDir = fmt.Sprintf("%s/absenceDB", a.rootDir)
		aa.rarityThreshold = a.absenceThreshold //absencyThreshold is not a mistake
		aa.linesInBlock = 0
		aa.maxBlocks = a.maxBlocks
		aa.haveStatistics = true
		aa.outputFunc = func(name string, rowID int64,
			scoreThreshold float64,
			score, scoreGap, scoreAvg, scoreStd float64,
			cnt int,
			text string) {
			text = strings.Replace(text, ",", "", -1)
			text2 := a.closedItemSets[text].text
			text3 := a.closedItemSets[text].lastLine
			if verbose || scoreGap >= scoreThreshold {
				msg := fmt.Sprintf("%s s=%5.2f g=%5.2f a=%5.2f | %s | %s",
					name,
					score,
					scoreGap,
					scoreAvg,
					text2,
					text3,
				)
				if verbose {
					logDebug(msg)
				} else {
					logInfo(msg)
				}
			}
		}

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
		case "absenceThreshold":
			a.absenceThreshold = k.MustFloat64(a.absenceThreshold)
		case "absenceCheck":
			a.absenceCheck = k.MustBool(a.absenceCheck)
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
	if err := a.loadDBclosedItemSets(); err != nil {
		return err
	}
	return nil
}

func (a *logAnalyzer) loadDBclosedItemSets() error {
	rows, err := a.dciDB.tables["closedItemSets"].query(nil, "*")
	if err != nil {
		return err
	}
	if rows == nil || len(rows) == 0 {
		return nil
	}
	for _, v := range rows {
		a.closedItemSets[v[0]] = closedItemSet{v[1], v[3]}
	}
	return nil
}

func (a *logAnalyzer) close() {
	a.rarAnal.close()
	if a.absenceCheck {
		a.frqAnal.close()
		a.absAnal.close()
	}
}

func (a *logAnalyzer) destroy() error {
	if err := a.rarAnal.destroy(); err != nil {
		return err
	}
	if err := a.dciDB.dropAllTables(); err != nil {
		return err
	}
	if err := a.frqAnal.destroy(); err != nil {
		return err
	}
	if err := a.absAnal.destroy(); err != nil {
		return err
	}

	return nil
}

func (a *logAnalyzer) dropPartitions(blockID int) error {
	blockIDstr := fmt.Sprint(blockID)
	if err := a.dciDB.tables["closedItemSets"].dropPartition(blockIDstr); err != nil {
		return err
	}
	if err := a.dciDB.tables["closedItemFirstLines"].dropPartition(blockIDstr); err != nil {
		return err
	}
	if err := a.dciDB.tables["closedItemLastLines"].dropPartition(blockIDstr); err != nil {
		return err
	}
	if err := a.dciDB.tables["closedItemSetsKeys"].dropPartition(blockIDstr); err != nil {
		return err
	}
	if err := a.dciDB.tables["closedItemSetsAbsent"].dropPartition(blockIDstr); err != nil {
		return err
	}
	return nil
}

func (a *logAnalyzer) getClosedItemString(fi []string) string {
	return strings.Join(fi, " ")
}

func (a *logAnalyzer) getClosedItemKey(fi []string) string {
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

	closedSets, supps, _, lastTIDs := dci.getClosedWordsSorted(items1)
	rows1 := make([][]string, len(supps))
	rows2 := make([][]string, len(supps))
	table := a.dciDB.tables["closedItemSets"]
	cols := table.colMap
	table2 := a.dciDB.tables["closedItemKeys"]
	cols2 := table2.colMap
	for i, cs := range closedSets {
		tek := a.getClosedItemKey(cs)
		tev := a.getClosedItemString(cs)
		fis := new(closedItemSet)
		fis.text = tev
		fis.lastLine = trans1.doc.get(lastTIDs[i])
		row1 := make([]string, 4)
		row2 := make([]string, 1)
		row1[cols["key"]] = tek
		row1[cols["itemSets"]] = tev
		row1[cols["support"]] = fmt.Sprint(supps[i])
		row1[cols["lastLine"]] = fis.lastLine
		row2[cols2["key"]] = tek
		a.closedItemSets[tek] = *fis
		rows1[i] = row1
		rows2[i] = row2
	}
	if err := a.dciDB.tables["closedItemSets"].insertRows(rows1,
		blockIDstr); err != nil {
		return err
	}
	if err := a.dciDB.tables["closedItemKeys"].insertRows(rows2,
		blockIDstr); err != nil {
		return err
	}
	fa := a.frqAnal
	fa.setLines(rows2)
	if _, err := fa.run(0, blockID); err != nil {
		return err
	}

	return nil
}

func (a *logAnalyzer) collectAbsence(blockID int) error {
	aa := a.absAnal
	aa.setAbsItems(a.frqAnal.items)
	if _, err := aa.run(0, blockID); err != nil {
		return err
	}
	return nil
}

func (a *logAnalyzer) run(targetLinesCnt int) error {
	linesProcessed := 0
	linesToProcess := 0
	for {
		if targetLinesCnt > 0 {
			linesToProcess = targetLinesCnt - linesProcessed
			if linesToProcess <= 0 {
				break
			}
		}
		cnt, err := a.rarAnal.runBeforeNextBlock(linesToProcess)
		if err != nil {
			return err
		}
		linesProcessed += cnt
		if cnt <= 0 {
			break
		}

		if a.absenceCheck && a.rarAnal.currBlock != nil && a.rarAnal.currBlock.completed {
			ra := a.rarAnal
			if err := a.collectFrequency(ra.currBlockID, ra.trans, ra.items); err != nil {
				return err
			}
			if err := a.collectAbsence(ra.currBlockID); err != nil {
				return err
			}
		}
		if cnt < a.linesInBlock {
			break
		}
		if targetLinesCnt > 0 && linesProcessed >= targetLinesCnt {
			break
		}
	}
	return nil
}
