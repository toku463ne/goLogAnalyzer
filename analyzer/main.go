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
	linesToProcess, nTopRecords int) (int, error) {

	a := newRarityAnalyzer(rootDir)
	if pathExist(rootDir) {
		if err := a.load(); err != nil {
			return 0, err
		}
	} else {
		if err := a.init(logPathRegex, filterStr, xFilterStr,
			minGapToRecord, maxBlocks, maxItemBlocks, linesInBlock, nTopRecords); err != nil {
			return 0, err
		}
	}
	linesProcessed, err := a.analyze(linesToProcess)
	if err != nil {
		return linesProcessed, err
	}
	if rootDir == "" {
		msg := fmt.Sprintf("%d top rare records", nTopRecords)
		a.printNTops(msg, nTopRecords, 0, filterStr, xFilterStr)

		if err := a.showRarStats("", cDefaultHistSize); err != nil {
			return linesProcessed, err
		}
	}
	return linesProcessed, nil
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

func RarStats(rootDir string, histSize int) error {
	a := newRarityAnalyzer(rootDir)
	if err := a.load(); err != nil {
		return err
	}

	return a.showRarStats(rootDir, histSize)
}
