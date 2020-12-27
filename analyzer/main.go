package analyzer

import "fmt"

// CleanupDb ... Drop all tables
func CleanupDb(rootDir string, debug bool) error {
	if debug {
		curLogLevel = cLogLevelDebug
	}

	a, err := newFileRarityAnalyzerByVars("",
		rootDir,
		"", "",
		0,
		0, 0)
	if err != nil {
		return err
	}
	if err := a.clean(); err != nil {
		return err
	}
	return nil
}

// Rar ...
func Rar(logPathRegex,
	rootDir string,
	filterRe, xFilterRe string,
	rarityThreshold float64,
	linesInBlock, maxBlocks int,
	debug bool, verbose1 bool) error {

	verbose = verbose1

	if debug {
		curLogLevel = cLogLevelDebug
	}

	a, err := newFileRarityAnalyzerByVars(logPathRegex,
		rootDir,
		filterRe, xFilterRe,
		rarityThreshold,
		linesInBlock, maxBlocks)
	if err != nil {
		return err
	}

	if a.useDB {
		if err := a.loadDB(); err != nil {
			return err
		}
	}

	defer a.close()
	var rowN int
	if rowN, err = a.run(0, -1); err != nil {
		return err
	}
	logInfo(fmt.Sprintf("Processed %d rows", rowN))

	if a.useDB {
		if err := a.SaveIni(); err != nil {
			logError(fmt.Sprintf("Failed to save config"))
		}
	}

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

	a := newFileAnalyzer(path, filterRe, xFilterRe)
	if err := a.tokenizeFile(); err != nil {
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
