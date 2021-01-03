package analyzer

import (
	"fmt"
)

// CleanupDb ... Drop all tables
func CleanupDb(rootDir string, debug bool) error {
	if debug {
		curLogLevel = cLogLevelDebug
	}

	a, err := newFileRarityAnalyzerByVars("",
		rootDir,
		"", "",
		0, 0,
		-1, -1)
	if err != nil {
		return err
	}
	if err := a.gainLock(); err != nil {
		fmt.Printf("%v : Probably another process is updating '%s'\n", err, rootDir)
		return nil
	}
	defer a.unLock()

	if err := a.clean(); err != nil {
		return err
	}
	a.unLock()
	return nil
}

// Rar ...
func Rar(logPathRegex,
	rootDir string,
	filterRe, xFilterRe string,
	rarityThreshold, rarityCountRate float64,
	linesInBlock, maxBlocks int,
	debug bool, verbose1 bool, saveDb bool) error {

	verbose = verbose1

	if debug {
		curLogLevel = cLogLevelDebug
	}

	a, err := newFileRarityAnalyzerByVars(logPathRegex,
		rootDir,
		filterRe, xFilterRe,
		rarityThreshold, rarityCountRate,
		linesInBlock, maxBlocks)
	if err != nil {
		return err
	}

	if a.useDB {
		if err := a.loadDB(); err != nil {
			return err
		}
		a.useDB = saveDb
		if err := a.gainLock(); err != nil {
			fmt.Printf("%v : Probably another process is updating '%s'\n", err, rootDir)
			return nil
		}
	}

	defer a.close()
	var rowN int
	if rowN, err = a.run(0, -1); err != nil {
		return err
	}
	logInfo(fmt.Sprintf("Processed %d rows", rowN))

	if a.useDB && saveDb {
		if err := a.SaveIni(); err != nil {
			logError(fmt.Sprintf("Failed to save config\n"))
		}
	}
	a.unLock()
	a.printCountPerGap(a.countPerGap, "Count per score gap")
	return nil
}

// Stats ... shows count per gap
func Stats(rootDir string) error {
	a, err := newFileRarityAnalyzerByVars("",
		rootDir,
		"", "",
		0, 0,
		-1, -1)
	if err != nil {
		return err
	}
	if err := a.loadDB(); err != nil {
		return err
	}
	a.printCountPerGap(a.countPerGap,
		fmt.Sprintf("Total count %d\ncounts per gap", a.countTotal))
	return nil
}

// Frq ... Get Closed Frequent Item Sets by DCI Closed Algorithm
func Frq(path string,
	minSupport int,
	filterRe, xFilterRe string,
	debug bool) error {

	if debug {
		curLogLevel = cLogLevelDebug
	}

	logInfo(fmt.Sprintf("Calculating closed frq itemsets. Max recs=%d", cMaxRecsToProcessFrq))
	a := newFileAnalyzer(path, filterRe, xFilterRe)
	if err := a.tokenizeFile(cMaxRecsToProcessFrq); err != nil {
		return err
	}
	if minSupport <= 0 {
		minSupport = a.rowNum / 10
		if minSupport <= 0 {
			minSupport = 10
		}
	}
	matrix := tran2BitMatrix(a.trans, a.items)
	dci, err := newDCIClosed(matrix, minSupport, true)
	if err != nil {
		return err
	}
	if err = dci.run(); err != nil {
		return err
	}
	dci.output(a.items, len(a.trans.getSlice()), a.trans.mask)

	return nil
}
