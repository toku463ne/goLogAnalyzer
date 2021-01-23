package analyzer

import (
	"fmt"
	"os"
)

// CleanupDB .. cleanup data with lock
func CleanupDB(rootDir string, debug bool) error {
	lock, err := newLock(rootDir)
	if err != nil {
		fmt.Printf("%v : Probably another process is updating '%s'\n", err, rootDir)
		return err
	}
	defer lock.unLock()
	err = CleanupDBProc(rootDir, debug)
	return err
}

// CleanupDBProc .. cleanup data : call this directly when debugging
func CleanupDBProc(rootDir string, debug bool) error {
	if debug {
		curLogLevel = cLogLevelDebug
	}

	a, err := newRarityAnalyzer("",
		rootDir,
		"", "",
		0,
		-1, -1, 0)
	if err != nil {
		return err
	}

	if err := a.clean(); err != nil {
		return err
	}
	return nil
}

// RunRar ... run rarity analyzer with lock
func RunRar(logPathRegex,
	rootDir string,
	filterRe, xFilterRe string,
	rarityThreshold float64,
	maxLines, linesInBlock, maxBlocks int,
	debug, verbose1, saveDb bool) error {

	lock, err := newLock(rootDir)
	if err != nil {
		fmt.Printf("%v : Probably another process is updating '%s'\n", err, rootDir)
		return err
	}
	defer lock.unLock()
	err = RunRarProc(logPathRegex,
		rootDir,
		filterRe, xFilterRe,
		rarityThreshold,
		maxLines, linesInBlock, maxBlocks,
		debug, verbose1, saveDb)
	return err
}

// RunRarProc ... run rarity analyzer
func RunRarProc(logPathRegex,
	rootDir string,
	filterRe, xFilterRe string,
	rarityThreshold float64,
	maxLines, linesInBlock, maxBlocks int,
	debug, verbose1, saveDb bool) error {

	verbose = verbose1

	if debug {
		curLogLevel = cLogLevelDebug
	}

	a, err := newRarityAnalyzer(logPathRegex,
		rootDir,
		filterRe, xFilterRe,
		rarityThreshold,
		linesInBlock, maxBlocks, 0)
	if err != nil {
		return err
	}

	if a.useDB {
		if err := a.loadDB(); err != nil {
			return err
		}
		a.useDB = saveDb
	}
	logInfo(fmt.Sprintf("[%d] datadir=%s search=%s exclude=%s bsize=%d nblocks=%d",
		os.Getpid(),
		a.rootDir,
		a.filterRe, a.xFilterRe,
		a.linesInBlock, a.maxBlocks))

	defer a.close()

	var rowN int
	if rowN, err = a.run(maxLines); err != nil {
		return err
	}
	logInfo(fmt.Sprintf("row=%d items=%d", rowN, len(a.trans.items.counts)))

	if err := a.printNTops(fmt.Sprintf("%d top rare records", cNTopRareRecords),
		cNTopRareRecords,
		a.filterRe, a.xFilterRe); err != nil {
		return err
	}
	a.printCountPerGap(a.countPerGap, "Count per score gap")
	return nil
}

// RarStats ... shows top N rare records and count per gap
func RarStats(rootDir string,
	recordsToShow int,
	filterRe, xFilterRe string,
) error {
	a, err := newRarityAnalyzer("",
		rootDir,
		"", "",
		0,
		-1, -1, 0)
	if err != nil {
		return err
	}
	if err := a.loadDB(); err != nil {
		return err
	}
	if recordsToShow == 0 {
		recordsToShow = cNTopRareRecords
	}
	a.printNTops(fmt.Sprintf("%d top rare records", recordsToShow),
		recordsToShow, filterRe, xFilterRe)
	a.printCountPerGap(a.countPerGap,
		fmt.Sprintf("Total count %d\ncounts per gap", a.countTotal))
	return nil
}

// RunFrq ... Get Closed Frequent Item Sets by DCI Closed Algorithm
func RunFrq(path string,
	minSupport int,
	filterRe, xFilterRe string,
	debug bool) error {

	if debug {
		curLogLevel = cLogLevelDebug
	}

	logInfo(fmt.Sprintf("Calculating closed frq itemsets. Max recs=%d", cMaxRecsToProcessFrq))

	dci, trans1, err := runDCIClosed(path,
		minSupport,
		filterRe, xFilterRe)
	if err != nil {
		return err
	}
	dci.output(trans1.items, trans1.maxTranID)

	return nil
}
