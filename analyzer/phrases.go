package analyzer

import (
	"fmt"
	"math"

	"github.com/pkg/errors"
)

func newPhrases(dataDir string, maxBlocks, maxRowsInBlock int) (*phrases, error) {
	p := new(phrases)
	d, err := newCircuitDB(dataDir, "phrases", tableDefs["phrases"], maxBlocks, 0)
	if err != nil {
		return nil, err
	}
	p.circuitDB = d

	p.maxRowsInBlock = maxRowsInBlock
	p.phraseMap = make(map[int]string, 10000)
	p.counts = make(map[int]int, 10000)
	p.phrases = make(map[string]int, 10000)
	p.tranScoreAvg = make(map[int]float64, 10000)
	p.currCounts = make(map[int]int, 10000)
	p.maxPhraseID = 0

	if IsDebug {
		msg := "phrases.newPhrases(): "
		msg += fmt.Sprintf("dataDir=%s maxBlocks=%d maxRowsInBlock=%d",
			dataDir, maxBlocks, maxRowsInBlock)
		ShowDebug(msg)
	}
	return p, nil
}

func (p *phrases) load() error {
	if p.dataDir != "" {
		if err := p.loadDB(); err != nil {
			return err
		}
	}
	return nil
}

func (p *phrases) register(phrase string, addCount int, tranScore float64, isNew bool) int {
	if phrase == "" {
		return -1
	}
	phraseID, ok := p.phrases[phrase]
	if !ok {
		p.maxPhraseID++
		phraseID = p.maxPhraseID
		p.phrases[phrase] = phraseID
		p.phraseMap[phraseID] = phrase
	}
	if addCount == 0 {
		return phraseID
	}

	if tranScore > 0 {
		p.tranScoreAvg[phraseID] = tranScore
	}
	p.counts[phraseID] += addCount
	if isNew {
		p.currCounts[phraseID] += addCount
	}
	return phraseID
}

func (p *phrases) getPhrase(phraseID int) string {
	if phraseID < 0 {
		return "-"
	}
	return p.phraseMap[phraseID]
}

func (p *phrases) getScore(phraseID, totalCount int) float64 {
	if totalCount == 0 {
		return 0
	}
	count := p.counts[phraseID]
	if count == 0 {
		return 0
	}
	score := math.Log(float64(totalCount)/float64(count)) + 1
	return score
}

func (p *phrases) getCount(phraseID int) int {
	return p.counts[phraseID]
}

func (p *phrases) getPhraseID(phrase string) int {
	phraseID, ok := p.phrases[phrase]
	if !ok {
		return -1
	}
	return phraseID
}

func (p *phrases) clearCurrCount() {
	p.currCounts = make(map[int]int, 10000)
}

func (p *phrases) loadDB() error {
	if p.dataDir == "" {
		return nil
	}
	cnt := p.statusTable.Count(nil)
	if cnt <= 0 {
		return nil
	}

	if err := p.loadCircuitDBStatus(); err != nil {
		return err
	}

	rows, err := p.selectRows(nil, nil, []string{"phrase", "phraseCount", "tranScoreAvg"})
	if err != nil {
		return err
	}
	if rows == nil {
		return nil
	}

	for rows.next() {
		var phrase string
		var phraseCount int
		var tranScoreAvg float64
		err = rows.scan(&phrase, &phraseCount, &tranScoreAvg)
		if err != nil {
			return err
		}
		p.register(phrase, phraseCount, tranScoreAvg, !rows.blockCompleted)
	}
	return nil
}

func (p *phrases) next() error {
	p.rowNo++
	if p.maxRowsInBlock > 0 && p.rowNo >= p.maxRowsInBlock {
		if p.dataDir != "" {
			if err := p.flush(); err != nil {
				return err
			}
		}
		p.clearCurrCount()
		p.nextBlock()
	}
	return nil
}

func (p *phrases) commit(completed bool) error {
	if p.dataDir == "" {
		return nil
	}
	if err := p.flush(); err != nil {
		return err
	}
	if err := p.updateBlockStatus(completed); err != nil {
		return err
	}
	return nil
}

func (p *phrases) flush() error {
	if p.dataDir == "" {
		return nil
	}
	for phraseID, cnt := range p.currCounts {
		term := p.getPhrase(phraseID)
		avg := p.tranScoreAvg[phraseID]
		if err := p.insertRow([]string{"item", "itemCount", "tranScoreAvg"},
			term, cnt, avg); err != nil {
			return err
		}
	}
	if err := p.currTable.FlushOverwrite(); err != nil {
		return errors.WithStack(err)
	}
	return nil
}
