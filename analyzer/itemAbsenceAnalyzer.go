package analyzer

import (
	"fmt"
	"io"
	"time"

	"github.com/pkg/errors"
)

type itemAbsenceAnalyzer struct {
	rarityAnalyzer
	absItems *items
	pos      int
	err      error
}

func newItemAbsenceAnalyzer() *itemAbsenceAnalyzer {
	a := new(itemAbsenceAnalyzer)
	a.name = "itemRarityAnal"
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
		text []string) {
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
		if a.absItems == nil {
			a.err = errors.New("No items to read")
			return false
		}
		for a.pos+1 < a.absItems.items.len() {
			a.err = io.EOF

			a.err = nil
			a.pos++

			if a.absItems.counts.get(a.pos) > 0 && a.absItems.newCounts.get(a.pos) == 0 {
				a.err = nil
				return true
			}
		}
		a.err = io.EOF
		return false
	}
	a.pointerText = func() []string {
		return []string{a.absItems.getWord(a.pos)}
	}
	a.pointerOpen = func() error {
		return nil
	}
	a.pointerClose = func() {
		a.items = nil
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

func (a *itemAbsenceAnalyzer) setAbsItems(absItems *items) {
	a.absItems = absItems
	for itemID, v := range a.items.items.getSlice() {
		_, ok := absItems.getItemID(v)
		if !ok {
			a.items.clearCount(itemID)
		}
	}
	a.items.totalCount = absItems.totalCount
	a.pos = -1
}
