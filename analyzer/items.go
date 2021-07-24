package analyzer

import (
	"fmt"
	"math"

	"github.com/pkg/errors"
)

func newItems(dataDir string, maxBlocks, maxRowsInBlock int) (*items, error) {
	d, err := newCircuitDB(dataDir, "items", maxBlocks, maxRowsInBlock)
	if err != nil {
		return nil, err
	}
	i := new(items)
	i.circuitDB = d
	i.termMap = make(map[int]string, 10000)
	i.counts = make(map[int]int, 10000)
	i.terms = make(map[string]int, 10000)
	i.currCounts = make(map[int]int, 10000)
	i.maxItemID = -1
	if dataDir != "" {
		if err := i.loadDB(); err != nil {
			return nil, err
		}
	}
	return i, nil
}

func (i *items) register(item string, addCount int, isNew bool) int {
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

func (i *items) addCount(itemID, cnt int) {
	i.counts[itemID] += cnt
}

func (i *items) subCount(item string, cnt int) {
	itemID := i.getItemID(item)
	currCnt := i.counts[itemID]
	currCnt -= cnt
	if currCnt == 0 {
		delete(i.counts, itemID)
		delete(i.terms, item)
		delete(i.termMap, itemID)
	} else {
		i.counts[itemID] = currCnt
	}
	i.totalCount -= cnt
}

func (i *items) getItemID(term string) int {
	itemID, ok := i.terms[term]
	if !ok {
		return -1
	}
	return itemID
}

func (i *items) getCurrCount(itemID int) int {
	return i.currCounts[itemID]
}

func (i *items) clearCurrCount() {
	i.currCounts = make(map[int]int, 10000)
}

func (i *items) next() error {
	i.rowNo++
	if i.maxRowsInBlock > 0 && i.rowNo >= i.maxRowsInBlock {
		if i.dataDir != "" {
			if err := i.commit(true); err != nil {
				return err
			}
		}
		i.clearCurrCount()
		i.nextBlock()
	}
	return nil
}

func (i *items) commit(completed bool) error {
	if err := i.dropBlock(i.blockNo); err != nil {
		return err
	}
	if err := i.createBlock(i.blockNo); err != nil {
		return err
	}

	sqlstr := fmt.Sprintf("INSERT INTO %s(item, itemCount) VALUES(?,?) \n",
		i.getBlockTableName(i.blockNo))
	stmt, err := i.conn.Prepare(sqlstr)
	if err != nil {
		return errors.WithStack(err)
	}

	for itemID, cnt := range i.currCounts {
		term := i.getWord(itemID)
		_, err := stmt.Exec(term, cnt)
		if err != nil {
			return errors.WithStack(err)
		}
	}

	err = i.updateBlockStatus(completed)
	return err
}

func (i *items) loadDB() error {
	rows, err := i.query([]string{"item", "itemCount"}, "")
	if err != nil {
		return err
	}
	if rows == nil {
		return nil
	}

	for rows.next() {
		var item string
		var itemCount int
		err = rows.scan(&item, &itemCount)
		if err != nil {
			return err
		}
		i.register(item, itemCount, !rows.blockCompleted)
	}
	return nil
}
