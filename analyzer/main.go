package analyzer

import (
	"fmt"
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
	debug, verbose1, saveDb, silent bool) error {

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
		debug, verbose1, saveDb, silent)
	return err
}

// RunRarProc ... run rarity analyzer
func RunRarProc(logPathRegex,
	rootDir string,
	filterRe, xFilterRe string,
	rarityThreshold float64,
	maxLines, linesInBlock, maxBlocks int,
	debug, verbose1, saveDb, silent bool) error {

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
	logInfo(fmt.Sprintf("datadir=%s search=%s exclude=%s bsize=%d nblocks=%d",
		a.rootDir,
		a.filterRe, a.xFilterRe,
		a.linesInBlock, a.maxBlocks))

	defer a.close()

	var rowN int
	if rowN, err = a.run(maxLines); err != nil {
		return err
	}
	if silent == false {
		logInfo(fmt.Sprintf("Completed. row=%d items=%d", rowN, len(a.trans.items.counts)))

		if err := a.printNTops(fmt.Sprintf("%d top rare records", cNTopRareRecords),
			cNTopRareRecords,
			a.filterRe, a.xFilterRe, nil); err != nil {
			return err
		}

		//a.printCountPerScore(a.countPerScore, "Count per score gap")
	}
	return nil
}

// RarStats ... shows count per gap
func RarStats(rootDir string) error {
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

	a.printCountPerScore(a.countPerScore,
		fmt.Sprintf("Total count=%d items=%d\ncounts per gap",
			a.countTotal, len(a.trans.items.counts)))
	return nil
}

// RarTopN ... shows top N rare records
func RarTopN(rootDir string,
	recordsToShow int,
	filterRe, xFilterRe string,
	startEpoch, endEpoch int64,
) error {
	a, err := newRarityAnalyzer("",
		rootDir,
		"", "",
		0,
		-1, -1, 0)
	if err != nil {
		return err
	}

	var blockIDstrs []string
	if startEpoch > 0 || endEpoch > 0 {
		if a.useDB {
			blockIDstrs, err = a.getBlockIDsFromEpoch(startEpoch, endEpoch)
			if err != nil {
				return err
			}
		}
	}

	if recordsToShow == 0 {
		recordsToShow = cNTopRareRecords
	}

	a.printNTops(fmt.Sprintf("%d top rare records", recordsToShow),
		recordsToShow, filterRe, xFilterRe, blockIDstrs)

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

// RunReadTest .. To measure read speed
func RunReadTest(logPathRegex string, maxLines int) error {
	a, err := newRarityAnalyzer(logPathRegex,
		"",
		"", "",
		0,
		0, 0, 0)
	if err != nil {
		return err
	}
	var rowN int
	if rowN, err = a.runOnlyRead(maxLines); err != nil {
		return err
	}
	logInfo(fmt.Sprintf("Processed %d records.", rowN))
	return nil
}

// RunReadTokenizeTest .. To measure read speed
func RunReadTokenizeTest(logPathRegex string, maxLines int) error {
	a, err := newRarityAnalyzer(logPathRegex,
		"",
		"", "",
		0,
		0, 0, 0)
	if err != nil {
		return err
	}
	var rowN int
	if rowN, err = a.runReadTokenize(maxLines); err != nil {
		return err
	}
	logInfo(fmt.Sprintf("Processed %d records. items=%d", rowN, len(a.trans.items.counts)))
	return nil
}

// RunReadTokenizeNoregTest .. To measure read speed
func RunReadTokenizeNoregTest(logPathRegex string, maxLines int) error {
	a, err := newRarityAnalyzer(logPathRegex,
		"",
		"", "",
		0,
		0, 0, 0)
	if err != nil {
		return err
	}
	var rowN int
	if rowN, err = a.tokenizeLineNogeg(maxLines); err != nil {
		return err
	}
	logInfo(fmt.Sprintf("Processed %d records", rowN))
	return nil
}
