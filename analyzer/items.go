package analyzer

import (
	"math"
)

type items struct {
	maxItemID int
	//items      *strindex
	items      map[string]int
	itemMap    map[int]string
	counts     map[int]int
	currCounts map[int]int
	scores     map[int]float64
	totalCount int
}

func newItems() *items {
	i := new(items)
	i.itemMap = make(map[int]string, 10000)
	i.counts = make(map[int]int, 10000)
	//i.items = newStrIndex(0)
	i.items = make(map[string]int, 10000)
	i.currCounts = make(map[int]int, 10000)
	i.maxItemID = -1
	return i
}

func (i *items) register(item string, addCount int, isNew bool) int {
	if item == "" {
		return -1
	}
	itemID, ok := i.items[item]
	if ok == false {
		i.maxItemID++
		itemID = i.maxItemID
		i.items[item] = itemID
		i.itemMap[itemID] = item
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
	return i.itemMap[itemID]
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
		delete(i.items, item)
		delete(i.itemMap, itemID)
	} else {
		i.counts[itemID] = currCnt
	}
	i.totalCount -= cnt
}

func (i *items) getItemID(word string) int {
	itemID, ok := i.items[word]
	if ok == false {
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
