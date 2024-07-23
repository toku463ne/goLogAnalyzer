package analyzer

import (
	"fmt"
	"log"
	"math"
	"regexp"
	"strings"
	"time"
)

func newTrans(dataDir string, maxBlocks, blockSize int,
	datetimeStartPos int, datetimeLayout string,
	scoreStyle, scoreNSize int, ignoreCount int) (*trans, error) {
	t := new(trans)
	i, err := newItems(dataDir, "items", maxBlocks, blockSize)
	if err != nil {
		return nil, err
	}
	p, err := newPhrases(dataDir, maxBlocks, blockSize)
	if err != nil {
		return nil, err
	}

	t.items = i
	t.items.register("", 1, 0, true)
	t.items.register(" ", 1, 0, true)
	t.phrases = p
	t.replacer = getDelimReplacer()
	t.blockSize = blockSize
	t.datetimeStartPos = datetimeStartPos
	t.datetimeLayout = datetimeLayout
	t.datetimeEndPos = datetimeStartPos + len(datetimeLayout)
	t.scoreStyle = scoreStyle
	if t.scoreStyle == 0 {
		t.scoreStyle = CDefaultScoreStyle
	}
	t.scoreNSize = scoreNSize
	if t.scoreNSize == 0 {
		t.scoreNSize = CDefaultScoreNSize
	}
	t.ignoreCount = ignoreCount
	if IsDebug {
		msg := "trans.newTrans(): "
		msg += fmt.Sprintf("dataDir=%s maxBlocks=%d blockSize=%d datetimeStartPos=%d datetimeLayout=%s scoreStyle=%d scoreNSize=%d",
			dataDir, maxBlocks, blockSize, datetimeStartPos, datetimeLayout, scoreStyle, scoreNSize)
		ShowDebug(msg)
	}

	return t, nil
}

func (t *trans) close() {
	if t.items.circuitDB != nil {
		t.items = nil
	}
}

func (t *trans) load() error {
	err := t.items.load()
	if err != nil {
		return err
	}
	return t.phrases.load()
}

func (t *trans) calcScore(prh []int, tran []int) float64 {
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
			//adjScore := t.items.calcAdjScore(itemID)
			//scores[i] = adjScore
			scores[i] = t.items.getScore(itemID)
			//t.items.topNItems.register(itemID, adjScore)
		}
		score = calcNAvgScore(scores, t.scoreStyle, t.scoreNSize)

	case cScoreConstSizeAvg:
		pscores := make([]float64, len(prh))
		tscores := make([]float64, len(tran))
		//blankItemID := t.items.getItemID(" ")
		//blankScore := t.items.getScore(blankItemID)
		for i, phraseID := range prh {
			pscores[i] = t.phrases.getScore(phraseID, t.items.totalCount)
		}
		for i, itemID := range tran {
			tscores[i] = t.items.getScore(itemID)
		}
		score = calcConstSizeAvgScore(pscores, tscores, t.scoreNSize)
	}
	for _, itemID := range tran {
		c := t.items.counts[itemID]
		if c <= 0 {
			continue
		}
		s := t.items.tranScoreAvg[itemID]
		if c == 1 || s == 0 {
			t.items.tranScoreAvg[itemID] = score
		} else {
			s = (s*(float64(c)-1) + score) / float64(c)
			t.items.tranScoreAvg[itemID] = s
		}
	}
	if math.IsNaN(score) {
		return 0
	}
	return score
}

func (t *trans) toTermList(line string, registerItem bool) ([]int, []int, []int, time.Time, error) {
	var timeResult []int
	var dt time.Time
	var err error

	if t.datetimeLayout != "" && len(line) > t.datetimeEndPos {
		timeResult = make([]int, 5)
		dt, err = time.Parse(t.datetimeLayout, line[:t.datetimeEndPos])
		if err == nil {
			line = line[t.datetimeEndPos+1:]
			timeResult[0] = t.items.register(fmt.Sprint(dt.Month()), 1, 0, registerItem)
			timeResult[1] = t.items.register(fmt.Sprintf("d-%02d", dt.Day()), 1, 0, registerItem)
			timeResult[2] = t.items.register(fmt.Sprintf("H-%02d", dt.Hour()), 1, 0, registerItem)
			timeResult[3] = t.items.register(fmt.Sprintf("M-%02d", dt.Minute()), 1, 0, registerItem)
			timeResult[4] = t.items.register(fmt.Sprint(dt.Weekday()), 1, 0, registerItem)
			//dtstr = dt.Format("2006-01-02T15:04:05")
			t.lastTimeResult = timeResult
		} else {
			if len(t.lastTimeResult) == 0 {
				return nil, nil, nil, dt, err
			}
			log.Printf("%v\n", err)
			log.Println("Applying the last parsed time for this line")
			timeResult = t.lastTimeResult
		}
	}

	line = t.replacer.Replace(line)
	words := strings.Split(line, " ")

	//for len(words) < t.scoreNSize {
	//	words = append(words, " ")
	//}

	var phraseIDs []int
	tokens := make([]int, 0)
	itemID := -1
	itemCnt := 0
	phrases := make([]string, len(termAppearenceInPhrases))
	phraseID := -1
	itemCnts := make([]int, len(termAppearenceInPhrases))
	phrase := ""

	var registerPhrase = func(itemID int, word string) {
		cnt := 0
		if itemID >= 0 {
			cnt = t.items.getCount(itemID)
		}
		for i, minCnt := range termAppearenceInPhrases {
			if cnt >= minCnt {
				itemCnts[i]++
				phrases[i] += " " + word
			} else {
				if itemCnt >= cMinTermInPhrases {
					phrase = strings.TrimSpace(phrases[i])
					if registerItem {
						phraseID = t.phrases.register(phrase, 1, 0, true)
					} else {
						phraseID = t.phrases.register(phrase, 0, 0, false)
					}
					phraseIDs = append(phraseIDs, phraseID)
				}
				itemCnts[i] = 0
				phrases[i] = ""
			}
		}
	}

	for _, w := range words {
		if _, ok := enStopWords[w]; ok {
			continue
		}

		word := strings.ToLower(w)
		lenw := len(word)
		if lenw > cWordMaxLen {
			word = word[:cWordMaxLen]
			lenw = cWordMaxLen
		}
		//remove '.' in the end
		if lenw > 1 && string(word[lenw-1]) == "." {
			word = word[:lenw-1]
		}

		if len(word) > 2 || word == " " {
			if isInt(word) && len(word) > cNumMaxDigits {
				continue
			}
			if registerItem {
				itemID = t.items.register(word, 1, 0, true)
			} else {
				itemID = t.items.register(word, 0, 0, false)
			}
			tokens = append(tokens, itemID)

			registerPhrase(itemID, word)
		}
	}
	registerPhrase(-1, "")
	return timeResult, tokens, phraseIDs, dt, nil
}

func (t *trans) tokenizeLine(line string,
	filterRe, xFilterRe *regexp.Regexp, registerItem bool) ([]int, []int, []int, time.Time, error) {
	timeTran, tran, phr, dt, err := t.toTermList(line, registerItem)
	if err != nil {
		return nil, nil, nil, dt, err
	}

	if line == "" {
		return nil, nil, nil, dt, err
	}
	b := []byte(line)

	if filterRe != nil && !filterRe.Match(b) {
		return nil, nil, nil, dt, err
	}
	if xFilterRe != nil && xFilterRe.Match(b) {
		return nil, nil, nil, dt, err
	}

	if registerItem {
		if err := t.items.next(); err != nil {
			return nil, nil, nil, dt, err
		}
	}

	return timeTran, tran, phr, dt, nil
}

func (t *trans) commit(completed bool) error {
	err := t.items.commit(completed)
	if err != nil {
		return err
	}
	return t.phrases.commit(completed)
}
