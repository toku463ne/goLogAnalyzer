package analyzer

func Clean(rootDir string) error {
	a := newRarityAnalyzer(rootDir)
	return a.clear()
}

func AnalyzeRarity(rootDir, logPathRegex, filterStr, xFilterStr string,
	minGapToRecord float64, maxBlocks, maxItemBlocks, linesInBlock int,
	linesToProcess int) (int, error) {

	a := newRarityAnalyzer(rootDir)
	if pathExist(rootDir) {
		if err := a.load(); err != nil {
			return 0, err
		}
	} else {
		if err := a.init(logPathRegex, filterStr, xFilterStr,
			minGapToRecord, maxBlocks, maxItemBlocks, linesInBlock); err != nil {
			return 0, err
		}
	}
	return a.analyze(linesToProcess)
}

func PrintRarTopN(rootDir, msg string,
	recordsToShow int, startEpoch int64,
	filterReStr, xFilterReStr string) error {
	a := newRarityAnalyzer(rootDir)
	if err := a.load(); err != nil {
		return err
	}

	return a.printNTops(msg,
		recordsToShow, startEpoch,
		filterReStr, xFilterReStr)
}
