package analyzer

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

func newTrans(dataDir string, maxBlocks, maxRowsInBlock int,
	datetimeStartPos int, datetimeLayout string) (*trans, error) {
	t := new(trans)
	i, err := newItems(dataDir, "items", maxBlocks, maxRowsInBlock)
	if err != nil {
		return nil, err
	}
	t.items = i
	t.items.register("", 1, true)
	t.replacer = getDelimReplacer()
	t.maxRowsInBlock = maxRowsInBlock
	t.datetimeStartPos = datetimeStartPos
	t.datetimeLayout = datetimeLayout
	t.datetimeEndPos = datetimeStartPos + len(datetimeLayout)
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

func (t *trans) toTermList(line string, registerItem bool) ([]int, int64) {
	var timeResult []int
	var lastEpoch int64
	if t.datetimeLayout != "" && len(line) > t.datetimeEndPos {
		timeResult = make([]int, 5)
		if dt, err := time.Parse(t.datetimeLayout, line[:t.datetimeEndPos]); err == nil {
			line = line[t.datetimeEndPos+1:]
			timeResult[0] = t.items.register(fmt.Sprint(dt.Month()), 1, registerItem)
			timeResult[1] = t.items.register(fmt.Sprintf("d-%02d", dt.Day()), 1, registerItem)
			timeResult[2] = t.items.register(fmt.Sprintf("H-%02d", dt.Hour()), 1, registerItem)
			timeResult[3] = t.items.register(fmt.Sprintf("M-%02d", dt.Minute()), 1, registerItem)
			timeResult[4] = t.items.register(fmt.Sprint(dt.Weekday()), 1, registerItem)
			lastEpoch = dt.Unix()
		}
	}

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
	if len(timeResult) > 0 {
		result = append(timeResult, result...)
	}
	return result, lastEpoch
}

func (t *trans) tokenizeLine(line string,
	filterRe, xFilterRe *regexp.Regexp, registerItem bool) ([]int, int64, error) {
	tran, lastEpoch := t.toTermList(line, registerItem)

	if line == "" {
		return nil, -1, nil
	}
	b := []byte(line)

	if filterRe != nil && !filterRe.Match(b) {
		return nil, -1, nil
	}
	if xFilterRe != nil && xFilterRe.Match(b) {
		return nil, -1, nil
	}

	if registerItem {
		if err := t.items.next(); err != nil {
			return nil, -1, err
		}
	}

	return tran, lastEpoch, nil
}

func (t *trans) commit(completed bool) error {
	return t.items.commit(completed)
}
