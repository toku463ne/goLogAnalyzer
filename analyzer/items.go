package analyzer

import (
	"fmt"
	"math"

	"github.com/pkg/errors"
)

func newItems(dataDir, tableName string, maxBlocks, maxRowsInBlock int) (*items, error) {
	i := new(items)
	d, err := newCircuitDB(dataDir, "items", tableDefs["items"], maxBlocks, 0)
	if err != nil {
		return nil, err
	}
	i.circuitDB = d

	i.maxRowsInItemBlock = maxRowsInBlock
	i.termMap = make(map[int]string, 10000)
	i.counts = make(map[int]int, 10000)
	i.terms = make(map[string]int, 10000)
	i.tranScoreAvg = make(map[int]float64, 10000)
	i.currCounts = make(map[int]int, 10000)
	i.maxItemID = 0

	if IsDebug {
		msg := "items.newItems(): "
		msg += fmt.Sprintf("dataDir=%s maxBlocks=%d maxRowsInBlock=%d",
			dataDir, maxBlocks, maxRowsInBlock)
		ShowDebug(msg)
	}
	return i, nil
}

func (i *items) load() error {
	if i.dataDir != "" {
		if err := i.loadDB(); err != nil {
			return err
		}
	}
	return nil
}

func (i *items) register(item string, addCount int, tranScore float64, isNew bool) int {
	if item == "" {
		return -1
	}
	itemID, ok := i.terms[item]
	if !ok {
		i.maxItemID++
		itemID = i.maxItemID
		i.terms[item] = itemID
		i.termMap[itemID] = item
	}
	if addCount == 0 {
		return itemID
	}

	if tranScore > 0 {
		i.tranScoreAvg[itemID] = tranScore
	}
	i.counts[itemID] += addCount
	if isNew {
		i.currCounts[itemID] += addCount
	}
	i.totalCount += addCount
	return itemID
}

func (i *items) getWord(itemID int) string {
	if itemID < 0 {
		return "-"
	}
	return i.termMap[itemID]
}

func (i *items) getScore(itemID int) float64 {
	if i.totalCount == 0 {
		return 0
	}
	count := i.counts[itemID]
	if count == 0 {
		return 0
	}
	score := math.Log(float64(i.totalCount)/float64(count)) + 1
	return score
}

func (i *items) getCount(itemID int) int {
	return i.counts[itemID]
}

func (i *items) getFrequency(itemID int) float64 {
	if i.totalCount == 0 {
		return 0.0
	}
	return float64(i.counts[itemID]) / float64(i.totalCount)
}

func (i *items) getItemID(term string) int {
	itemID, ok := i.terms[term]
	if !ok {
		return -1
	}
	return itemID
}

func (i *items) clearCurrCount() {
	i.currCounts = make(map[int]int, 10000)
}

func (i *items) loadDB() error {
	if i.dataDir == "" {
		return nil
	}
	cnt := i.statusTable.Count(nil)
	if cnt <= 0 {
		return nil
	}

	if err := i.loadCircuitDBStatus(); err != nil {
		return err
	}

	rows, err := i.selectRows(nil, nil, []string{"item", "itemCount", "tranScoreAvg"})
	if err != nil {
		return err
	}
	if rows == nil {
		return nil
	}

	for rows.next() {
		var item string
		var itemCount int
		var tranScoreAvg float64
		err = rows.scan(&item, &itemCount, &tranScoreAvg)
		if err != nil {
			return err
		}
		i.register(item, itemCount, tranScoreAvg, !rows.blockCompleted)
	}
	return nil
}

func (i *items) next() error {
	i.rowNo++
	if i.maxRowsInItemBlock > 0 && i.rowNo >= i.maxRowsInItemBlock {
		if i.dataDir != "" {
			if err := i.flush(); err != nil {
				return err
			}
		}
		i.clearCurrCount()
		i.nextBlock()
	}
	return nil
}

func (i *items) commit(completed bool) error {
	if i.dataDir == "" {
		return nil
	}
	if err := i.flush(); err != nil {
		return err
	}
	if err := i.updateBlockStatus(completed); err != nil {
		return err
	}
	return nil
}

func (i *items) flush() error {
	if i.dataDir == "" {
		return nil
	}
	for itemID, cnt := range i.currCounts {
		term := i.getWord(itemID)
		avg := i.tranScoreAvg[itemID]
		if err := i.insertRow([]string{"item", "itemCount", "tranScoreAvg"},
			term, cnt, avg); err != nil {
			return err
		}
	}
	if err := i.currTable.FlushOverwrite(); err != nil {
		return errors.WithStack(err)
	}
	return nil
}

func (i *items) calcAdjScore(itemID int) float64 {
	s := i.getScore(itemID)
	m := i.tranScoreAvg[itemID]
	if m == 0 {
		return s // this score is suspicious
	} else {
		return (s + m) / 2
		//return m
	}
}
