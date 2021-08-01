package analyzer

import (
	"fmt"
	"log"
	"os"
)

func Clean(rootDir string) error {
	if pathExist(rootDir) {
		log.Printf("removing '%s'", rootDir)
		if err := os.RemoveAll(rootDir); err != nil {
			log.Printf("failed to remove the dir\n Try 'rm -rf %s'", rootDir)
			return err
		}
	} else {
		log.Printf("'%s' does not exist", rootDir)
	}

	return nil
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

func RarStats(rootDir string) error {
	a := newRarityAnalyzer(rootDir)
	if err := a.load(); err != nil {
		return err
	}

	g, err := a.stats.getAllScorePerCount()
	if err != nil {
		return err
	}

	fmt.Printf("\n")
	fmt.Printf("Counts per score\n")
	fmt.Printf(" score | count\n")
	fmt.Printf(" ------+--------------\n")
	for i := 0; i < cCountbyScoreLen; i++ {
		if g[i] > 0 {
			fmt.Printf("   %02.1f | %d\n", float64(i), g[i])
		}
	}

	return nil
}
