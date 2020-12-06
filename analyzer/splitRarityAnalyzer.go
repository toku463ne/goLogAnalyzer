package analyzer

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/pkg/errors"
)

type splitRarityAnalyzer struct {
	rarityAnalyzer
	lines [][]string
	pos   int
	err   error
}

func newSplitRarityAnalyzer() *splitRarityAnalyzer {
	a := new(splitRarityAnalyzer)
	a.name = "splitRarityAnal"
	a.rootDir = fmt.Sprintf("./%s", a.name)
	a.rarityThreshold = 0.8
	a.linesInBlock = 10000
	a.maxBlocks = 1000
	a.useDB = false
	a.haveStatistics = true
	a.pos = -1

	a.outputFunc = func(name string, rowID int64,
		scoreThreshold float64,
		score, scoreGap, scoreAvg, scoreStd float64,
		cnt int,
		text string) {
		return
	}

	a.countTargetLines = func() (int, error) {
		return 0, nil
	}

	a.setTargetLinesCnt = func(targetLinesCnt int) error {
		a.targetLinesCnt += targetLinesCnt
		return nil
	}

	a.pointerNext = func() bool {
		if a.lines == nil {
			a.err = errors.New("No lines to read")
			return false
		}
		if a.pos+1 >= len(a.lines) {
			a.err = io.EOF
			return false
		}
		a.err = nil
		a.pos++
		return true
	}
	a.pointerText = func() string {
		return strings.Join(a.lines[a.pos], " ")
	}
	a.pointerOpen = func() error {
		return nil
	}
	a.pointerClose = func() {
		a.lines = nil
	}
	a.pointerCurrEpoch = func() int64 {
		return time.Now().Unix()
	}
	a.pointerCurrName = func() string {
		return fmt.Sprint(a.currBlockID)
	}
	a.pointerPos = func() int {
		return a.pos
	}
	a.pointerErr = func() error {
		return a.err
	}

	return a
}

func (a *splitRarityAnalyzer) setLines(lines [][]string) {
	a.pos = -1
	a.lines = lines
}
