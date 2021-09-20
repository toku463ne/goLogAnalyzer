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
	if err := a.open(logPathRegex, filterStr, xFilterStr,
		minGapToRecord, maxBlocks, maxItemBlocks, linesInBlock, nTopRecords); err != nil {
		return 0, err
	}
	linesProcessed, err := a.analyze(linesToProcess)
	if err != nil {
		return linesProcessed, err
	}
	if rootDir == "" {
		msg := fmt.Sprintf("%d top rare records", nTopRecords)
		out, _, err := a.getNTopString(msg, nTopRecords, 0, 0, filterStr, xFilterStr, false, 0, 0)
		if err != nil {
			return linesProcessed, err
		}

		out2, err := a.getRarStatsString("", cDefaultHistSize)
		if err != nil {
			return linesProcessed, err
		}
		out += out2
		println(out)
	}
	return linesProcessed, nil
}

func PrintRarTopN(rootDir, msg string,
	recordsToShow int, startEpoch, endEpoch int64,
	filterReStr, xFilterReStr string, showItemScore bool,
	minScore float64, maxScore float64) error {
	a := newRarityAnalyzer(rootDir)
	if err := a.load(); err != nil {
		return err
	}

	if out, _, err := a.getNTopString(msg,
		recordsToShow, startEpoch, endEpoch,
		filterReStr, xFilterReStr, showItemScore, minScore, maxScore); err != nil {
		return err
	} else {
		println(out)
	}
	return nil
}

func RarStats(rootDir string, histSize int) error {
	a := newRarityAnalyzer(rootDir)
	if err := a.load(); err != nil {
		return err
	}

	out, err := a.getRarStatsString(rootDir, histSize)
	if err != nil {
		return err
	}
	println(out)
	return nil
}

func Report(jsonFile string, recentNdays int,
	defaultMinGapToRecord float64,
	defaultMaxBlocks, defaultMaxItemBlocks,
	defaultLinesInBlock, defaultNTopRecords, defaultHistSize int,
	defaultOutFormat string) error {

	ls, err := newLogSetInfo(jsonFile)
	if err != nil {
		return err
	}

	err = ls.run(recentNdays,
		defaultMinGapToRecord,
		defaultMaxBlocks, defaultMaxItemBlocks,
		defaultLinesInBlock, defaultNTopRecords, defaultHistSize,
		defaultOutFormat)
	if err != nil {
		return err
	}
	return nil
}
