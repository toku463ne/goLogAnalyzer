package analyzer

import (
	"bufio"
	"fmt"
	"os"

	"github.com/pkg/errors"
)

type trans struct {
	tranList  *int2dArray
	maxTranID int
	doc       *stringArray
	mask      *intArray
	scores    *float64Array
	lastScore float64
	lastMsg   string
}

func newTrans() *trans {
	t := new(trans)
	t.tranList = newInt2dArray()
	t.doc = newStringArray()
	t.maxTranID = -1
	t.mask = newIntArray()
	t.scores = newFloat64Array()
	return t
}

func (t *trans) add(tran []int, d string, items1 *items) {
	t.maxTranID++
	t.tranList.set(t.maxTranID, tran)
	t.doc.set(t.maxTranID, d)
	score := calcScore(tran, items1)
	t.scores.set(t.maxTranID, score)
	t.lastScore = score
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

func calcScore(tran []int, items1 *items) float64 {
	l := len(tran)
	if l == 0 {
		return 0.0
	}
	score := 0.0
	for _, itemID := range tran {
		s := items1.getScore(itemID)
		score += s
	}
	score /= float64(l)
	return score
}

func (t *trans) getScore(i int, items1 *items) float64 {
	if t.scores.has(i) {
		return t.scores.get(i)
	}
	return 0.0
}

func (t *trans) getLastScore() float64 {
	return t.lastScore
}

func tokenizeFile(filepath string) (*trans, *items, error) {
	file, err := os.Open(filepath)
	defer file.Close()
	trans1 := newTrans()
	items1 := newItems()

	if err != nil {
		return nil, nil, errors.Wrapf(err, fmt.Sprintf("file open error: %s", filepath))
	}
	reader := bufio.NewReader(file)
	var line string
	eof := false
	i := 0
	for {
		line, err = reader.ReadString('\n')
		if err != nil {
			eof = true
		}
		if eof {
			break
		}

		trans1.mask.append(i)
		bline := []byte(line)

		tran := getEnItems(bline, items1, "")
		if len(tran) > 0 {
			trans1.add(tran, line, items1)
		}
		i++
	}
	return trans1, items1, nil
}

func (t *trans) outScore(outfile string) error {
	ou, err := os.Create(outfile)
	if err != nil {
		return err
	}
	w := bufio.NewWriter(ou)
	defer ou.Close()
	rowIDs := t.mask.getSlice()
	for i, rowID := range rowIDs {
		line := fmt.Sprintf("%v %v\n", rowID, t.scores.get(i))
		if _, err := fmt.Fprint(w, line); err != nil {
			return err
		}
	}
	w.Flush()
	return nil
}
