package analyzer

import (
	"fmt"
	"io"
)

type fileRarityAnalyzerVars struct {
	name            string
	useDB           bool
	filterRe        string
	xFilterRe       string
	rootDir         string
	logPathRegex    string
	rarityThreshold float64
	linesInBlock    int
	maxBlocks       int
}

type fileRarityAnalyzer struct {
	rarityAnalyzer
	logPathRegex string
	fp           *filePointer
}

func newFileRarityAnalyzer() *fileRarityAnalyzer {
	a := new(fileRarityAnalyzer)
	a.name = "fileRarityAnal"
	a.rootDir = fmt.Sprintf("./%s", a.name)
	a.rarityThreshold = 0.8
	a.linesInBlock = 10000
	a.maxBlocks = 1000
	a.useDB = false
	a.haveStatistics = true

	a.outputFunc = func(name string, rowID int64,
		scoreThreshold float64,
		score, scoreGap, scoreAvg, scoreStd float64,
		cnt int,
		text string) {
		if verbose || scoreGap > scoreThreshold {
			msg := fmt.Sprintf("%s s=%5.2f g=%5.2f a=%5.2f | %s",
				name,
				score,
				scoreGap,
				scoreAvg,
				text,
			)
			logInfo(msg)
		}
	}

	a.countTargetLines = func() (int, error) {
		targetLinesCnt := 0
		if err := a.pointerOpen(); err != nil {
			return -1, err
		}
		defer a.fp.close()
		for a.fp.next() {
			targetLinesCnt++
		}
		if err := a.fp.err(); err != nil && err != io.EOF {
			return -1, err
		}
		return targetLinesCnt, nil
	}

	a.setTargetLinesCnt = func(targetLinesCnt int) error {
		if a.maxTargetLinesCnt == 0 {
			maxTargetLinesCnt, err := a.countTargetLines()
			if err != nil {
				return err
			}
			a.maxTargetLinesCnt = maxTargetLinesCnt
		}

		if targetLinesCnt <= 0 {
			a.targetLinesCnt = a.maxTargetLinesCnt
			return nil
		}
		a.targetLinesCnt += targetLinesCnt
		if a.targetLinesCnt > a.maxTargetLinesCnt {
			a.targetLinesCnt = a.maxTargetLinesCnt
		}
		return nil
	}

	a.pointerNext = func() bool {
		return a.fp.next()
	}
	a.pointerText = func() string {
		return a.fp.text()
	}
	a.pointerOpen = func() error {
		if a.fp == nil || !a.fp.isOpen() {
			a.fp = newFilePointer(a.logPathRegex, a.lastFileEpoch, a.lastFileRow)
			if err := a.fp.open(); err != nil {
				return err
			}
		}
		return nil
	}
	a.pointerClose = func() {
		if a.fp != nil {
			a.fp.close()
		}
	}
	a.pointerCurrEpoch = func() int64 {
		return a.fp.currFileEpoch()
	}
	a.pointerCurrName = func() string {
		return a.fp.currFile()
	}
	a.pointerPos = func() int {
		return a.fp.row()
	}
	a.pointerErr = func() error {
		return a.fp.err()
	}

	return a
}

func newFileRarityAnalyzerByVars(v *fileRarityAnalyzerVars) (*fileRarityAnalyzer, error) {
	a := newFileRarityAnalyzer()
	a.name = v.name
	a.useDB = v.useDB
	a.filterRe = v.filterRe
	a.xFilterRe = v.xFilterRe
	a.rootDir = v.rootDir
	a.logPathRegex = v.logPathRegex
	a.rarityThreshold = v.rarityThreshold
	a.linesInBlock = v.linesInBlock
	a.maxBlocks = v.maxBlocks

	if err := a.init(); err != nil {
		return nil, err
	}
	return a, nil
}
