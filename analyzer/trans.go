package analyzer

import (
	"regexp"
	"strings"
)

func newTrans(dataDir string, maxBlocks, maxRowsInBlock int) (*trans, error) {
	t := new(trans)
	i, err := newItems(dataDir, maxBlocks, maxRowsInBlock)
	if err != nil {
		return nil, err
	}
	t.items = i
	t.items.register("", 1, true)
	t.replacer = getDelimReplacer()
	t.maxRowsInBlock = maxRowsInBlock
	return t, nil
}

func (t *trans) close() {
	if t.items.circuitDB != nil {
		t.items = nil
	}
}

func (t *trans) load() error {
	return t.items.load()
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

func (t *trans) toTermList(line string, registerItem bool) []int {
	line = t.replacer.Replace(line)
	words := strings.Split(line, " ")
	result := make([]int, len(words))
	for i, w := range words {
		if _, ok := enStopWords[w]; ok {
			continue
		}

		word := strings.ToLower(w)
		if len(word) > cWordMaxLen {
			word = word[:cWordMaxLen]
		}
		result[i] = 0
		if len(word) > 2 {
			if isInt(word) && len(word) > cNumMaxDigits {
				continue
			}
			if registerItem {
				result[i] = t.items.register(word, 1, true)
			} else {
				result[i] = t.items.register(word, 0, false)
			}
		}
	}
	return result
}

func (t *trans) tokenizeLine(line string,
	filterRe, xFilterRe *regexp.Regexp, registerItem bool) ([]int, error) {
	tran := t.toTermList(line, registerItem)

	if line == "" {
		return nil, nil
	}
	b := []byte(line)

	if filterRe != nil && !filterRe.Match(b) {
		return nil, nil
	}
	if xFilterRe != nil && xFilterRe.Match(b) {
		return nil, nil
	}

	if registerItem {
		if err := t.items.next(); err != nil {
			return nil, err
		}
	}

	return tran, nil
}

func (t *trans) commit(completed bool) error {
	return t.items.commit(completed)
}
