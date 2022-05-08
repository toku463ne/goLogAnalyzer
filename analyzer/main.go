package analyzer

import (
	"fmt"
	"log"
	"os"

	"github.com/pkg/errors"
)

func Clean(rootDir string) error {
	SetNamespace(rootDir)
	if PathExist(rootDir) {
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

func Run(c *AnalConf) error {
	SetNamespace(c.LogPathRegex)
	a, err := newRarityAnalyzer(c)
	if err != nil {
		return err
	}
	log.Printf("blockSize=%d maxBlocks=%d maxItemBlocks=%d minGap=%1.1f",
		a.BlockSize, a.MaxBlocks, a.MaxItemBlocks, a.MinGapToRecord)
	if err := a.analyze(0); err != nil {
		return err
	}
	if c.RootDir == "" {
		filterReStr := re2str(a.filterRe)
		xFilterReStr := re2str(a.xFilterRe)
		ntop, err := a.getNTop("ntop", a.NTopRecordsCount, 0, 0,
			filterReStr, xFilterReStr, 0, 0, a.NRareTerms)
		if err != nil {
			return err
		}
		msg := fmt.Sprintf("%d top rare records", a.NTopRecordsCount)
		out, _, err := ntop.getString(msg, a.NTopRecordsCount, a.NRareTerms)
		if err != nil {
			return err
		}
		println(out)
	}

	return nil
}

func PrintTopN(rootDir string, n int,
	filterRe, xFilterRe string,
	startEpoch, endEpoch int64,
	minScore, maxScore float64, nRareTerms int) error {
	SetNamespace(rootDir)
	if rootDir == "" {
		return errors.New("rootDir cannot be empty")
	}
	if !PathExist(rootDir) {
		return errors.New("Run analyzation first.")
	}
	c := NewAnalConf(rootDir)
	c.NRareTerms = nRareTerms
	a, err := newRarityAnalyzer(c)
	if err != nil {
		return err
	}

	ntop, err := a.getNTop("ntop", n, startEpoch, endEpoch,
		filterRe, xFilterRe, minScore, maxScore, nRareTerms)
	if err != nil {
		return err
	}

	if out, _, err := ntop.getString(fmt.Sprintf("Top %d rare messages", n), n, nRareTerms); err != nil {
		return err
	} else {
		println(out)
	}
	return nil
}

func RarStats(rootDir string, histSize int) error {
	c := NewAnalConf(rootDir)
	a, err := newRarityAnalyzer(c)
	if err != nil {
		return err
	}
	SetNamespace(rootDir)
	out, err := a.getRarStatsString(rootDir, histSize)
	if err != nil {
		return err
	}
	println(out)
	return nil
}

func Report(jsonFile string, nDays int) error {
	r, err := newReport(jsonFile, nDays)
	if err != nil {
		return err
	}
	if err := r.run(); err != nil {
		return err
	}
	return nil
}
