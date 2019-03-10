package analyzer

type items struct {
	maxItemID int
	itemMap   map[string]int
	items     *stringArray
	idfScores *float64Array
	maxIDF    float64
}

func newItems() *items {
	i := new(items)
	i.itemMap = make(map[string]int, 10000)
	i.items = newStringArray()
	i.idfScores = newFloat64Array()
	i.maxItemID = -1
	return i
}

func (i *items) regist(item string) int {
	if itemID, ok := i.itemMap[item]; ok {
		return itemID
	}
	i.maxItemID++
	i.itemMap[item] = i.maxItemID
	i.items.set(i.maxItemID, item)
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
