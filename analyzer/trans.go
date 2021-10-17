package analyzer

import (
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"
)

func newTrans(dataDir string, maxBlocks, maxRowsInBlock int,
	datetimeStartPos int, datetimeLayout string, scoreStyle int) (*trans, error) {
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
	t.scoreStyle = scoreStyle
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

	switch t.scoreStyle {
	case cScoreSimpleAvg:
		for _, itemID := range tran {
			s := t.items.getScore(itemID)
			score += s
		}
		score /= float64(l)

	case cScoreNAvg, cScoreNDistAvg:
		tranSize := len(tran)
		scores := make([]float64, tranSize)
		for i, itemID := range tran {
			s := t.items.getScore(itemID)
			scores[i] = s
		}
		score = calcNAvgScore(scores, t.scoreStyle)
	}

	return score
}

func (t *trans) toTermList(line string, registerItem bool) ([]int, []int, time.Time, error) {
	var timeResult []int
	var dt time.Time
	var err error
	if t.datetimeLayout != "" && len(line) > t.datetimeEndPos {
		timeResult = make([]int, 5)
		dt, err = time.Parse(t.datetimeLayout, line[:t.datetimeEndPos])
		if err == nil {
			line = line[t.datetimeEndPos+1:]
			timeResult[0] = t.items.register(fmt.Sprint(dt.Month()), 1, registerItem)
			timeResult[1] = t.items.register(fmt.Sprintf("d-%02d", dt.Day()), 1, registerItem)
			timeResult[2] = t.items.register(fmt.Sprintf("H-%02d", dt.Hour()), 1, registerItem)
			timeResult[3] = t.items.register(fmt.Sprintf("M-%02d", dt.Minute()), 1, registerItem)
			timeResult[4] = t.items.register(fmt.Sprint(dt.Weekday()), 1, registerItem)
			//dtstr = dt.Format("2006-01-02T15:04:05")
			t.lastTimeResult = timeResult
		} else {
			if len(t.lastTimeResult) == 0 {
				return nil, nil, dt, err
			}
			log.Printf("%v\n", err)
			log.Println("Applying the last parsed time for this line")
			timeResult = t.lastTimeResult
		}
	}

	line = t.replacer.Replace(line)
	words := strings.Split(line, " ")
	result := make([]int, 0)
	for _, w := range words {
		if _, ok := enStopWords[w]; ok {
			continue
		}

		word := strings.ToLower(w)
		if len(word) > cWordMaxLen {
			word = word[:cWordMaxLen]
		}
		if len(word) > 2 {
			if isInt(word) && len(word) > cNumMaxDigits {
				continue
			}
			if registerItem {
				result = append(result, t.items.register(word, 1, true))
			} else {
				result = append(result, t.items.register(word, 0, false))
			}
		}
	}
	return timeResult, result, dt, nil
}

func (t *trans) tokenizeLine(line string,
	filterRe, xFilterRe *regexp.Regexp, registerItem bool) ([]int, []int, time.Time, error) {
	timeTran, tran, dt, err := t.toTermList(line, registerItem)
	if err != nil {
		return nil, nil, dt, err
	}

	if line == "" {
		return nil, nil, dt, err
	}
	b := []byte(line)

	if filterRe != nil && !filterRe.Match(b) {
		return nil, nil, dt, err
	}
	if xFilterRe != nil && xFilterRe.Match(b) {
		return nil, nil, dt, err
	}

	if registerItem {
		if err := t.items.next(); err != nil {
			return nil, nil, dt, err
		}
	}

	return timeTran, tran, dt, nil
}

func (t *trans) commit(completed bool) error {
	return t.items.commit(completed)
}
