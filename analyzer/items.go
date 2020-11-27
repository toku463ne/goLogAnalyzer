package analyzer

import (
	"math"
)

type items struct {
	maxItemID  int
	itemMap    map[string]int
	items      *stringArray
	counts     *intArray
	newCounts  *intArray
	totalCount int
}

func newItems() *items {
	i := new(items)
	i.itemMap = make(map[string]int, 10000)
	i.items = newStringArray()
	i.counts = newIntArray()
	i.newCounts = newIntArray()
	i.maxItemID = -1
	return i
}

func (i *items) regist(item string, count int, isNew bool) int {
	if itemID, ok := i.itemMap[item]; ok {
		i.counts.set(itemID, i.counts.get(itemID)+count)
		if isNew {
			i.newCounts.set(itemID, i.newCounts.get(itemID)+count)
		}
		i.totalCount += count
		return itemID
	}
	i.maxItemID++
	i.itemMap[item] = i.maxItemID
	i.items.set(i.maxItemID, item)
	i.counts.set(i.maxItemID, count)
	if isNew {
		i.newCounts.set(i.maxItemID, count)
	}
	i.totalCount += count
	return i.maxItemID
}

func (i *items) getSlice() []string {
	return i.items.getSlice()
}

func (i *items) getWord(itemID int) string {
	if itemID < 0 {
		return "-"
	}
	if itemID > i.items.pos {
		return "-"
	}
	return i.items.get(itemID)
}

func (i *items) getScore(itemID int) float64 {
	if i.totalCount == 0 {
		return 0
	}
	count := i.counts.get(itemID)
	if count == 0 {
		return 0
	}
	score := math.Log(float64(i.totalCount)/float64(count)) + 1
	return score
}

func (i *items) getCount(itemID int) int {
	return i.counts.get(itemID)
}

func (i *items) getNewCount(itemID int) int {
	return i.newCounts.get(itemID)
}

func (i *items) clearNewCount() {
	i.newCounts = newIntArray()
}
