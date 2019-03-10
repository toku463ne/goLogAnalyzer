package analyzer

type trans struct {
	tranList       *int2dArray
	maxTranID      int
	tranTimeStamps *stringArray
	doc            *stringArray
	mask           *intArray
}

func newTrans() *trans {
	t := new(trans)
	t.tranList = newInt2dArray()
	t.doc = newStringArray()
	t.tranTimeStamps = newStringArray()
	t.maxTranID = -1
	t.mask = newIntArray()
	return t
}

func (t *trans) add(timeStamp string, tran []int, d string) {
	t.maxTranID++
	t.tranTimeStamps.set(t.maxTranID, timeStamp)
	t.tranList.set(t.maxTranID, tran)
	t.doc.set(t.maxTranID, d)
}

func (t *trans) get(i int) []int {
	return t.tranList.get(i)
}

func (t *trans) getSlice() [][]int {
	return t.tranList.getSlice()
}

func (t *trans) getWordsAt(i int, items1 *items) []string {
	tran := t.get(i)
	tw := make([]string, len(tran))
	for j, itemID := range tran {
		tw[j] = items1.getWord(itemID)
	}
	return tw
}
