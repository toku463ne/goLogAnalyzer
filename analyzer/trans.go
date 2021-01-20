package analyzer

import (
	"bytes"
	"fmt"
	"regexp"

	"golang.org/x/text/unicode/norm"
)

type trans struct {
	tranList  *int2dArray
	maxTranID int
	items     *items
}

func newTrans(useTranList bool) *trans {
	t := new(trans)
	if useTranList {
		t.tranList = newInt2dArray()
	}
	t.items = newItems()
	t.maxTranID = -1
	return t
}

func (t *trans) add(tran []int) {
	t.maxTranID++
	t.tranList.set(t.maxTranID, tran)
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

func (t *trans) getSentenceAt(i int, items1 *items) string {
	tran := t.get(i)
	s := ""
	for _, itemID := range tran {
		if s == "" {
			s = items1.getWord(itemID)
		} else {
			s += " " + items1.getWord(itemID)
		}
	}
	return s
}

func (t *trans) toBitMatrix() *bitMatrix {
	matrix := newBitMatrix(t.maxTranID+1, t.items.maxItemID+1)
	for i := 0; i <= t.maxTranID; i++ {
		for _, itemID := range t.get(i) {
			matrix.set(i, itemID)
		}
	}
	return matrix
}

func (t *trans) calcScore(tran []int) float64 {
	l := len(tran)
	if l == 0 {
		return 0.0
	}
	score := 0.0
	for _, itemID := range tran {
		s := t.items.getScore(itemID)
		score += s
	}
	score /= float64(l)
	return score
}

func (t *trans) toTermList(content []byte) []int {
	content = bytes.ToLower(content)
	content = remTags.ReplaceAll(content, []byte(" "))
	content = norm.NFC.Bytes(content)
	words := regexp.MustCompile(fmt.Sprintf("%s", cWordReStr)).FindAll(content, -1)
	result := make([]int, 0)
	for _, w := range words {
		word := string(w)
		if len(word) <= 2 {
			continue
		}
		if _, ok := enStopWords[word]; !ok {
			itemID := t.items.register(word, 1, true)
			result = append(result, itemID)
		}
	}
	return result
}

func (t *trans) tokenizeLine(line, filterRe, xFilterRe string) []int {
	bline := []byte(line)
	tran := t.toTermList(bline)

	if line == "" {
		return nil
	}

	if filterRe != "" && searchReg(line, filterRe) == false {
		return nil
	}
	if xFilterRe != "" && searchReg(line, xFilterRe) {
		return nil
	}

	l := len(tran)
	if t.tranList != nil && l > 0 {
		t.add(tran)
	}
	return tran

}
