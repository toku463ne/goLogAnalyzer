package logan

import (
	"goLogAnalyzer/pkg/csvdb"
	"math"
	"sort"
)

type terms struct {
	*csvdb.CircuitDB
	maxTermId  int
	term2Id    map[string]int
	id2term    map[int]string
	counts     map[int]int
	currCounts map[int]int
	totalCount int
	testMode   bool
}

func newTerms(dataDir string,
	maxBlocks int,
	unitSecs, keepPeriod int64,
	useGzip, testMode bool) (*terms, error) {
	te := new(terms)
	te.maxTermId = 0
	te.term2Id = make(map[string]int, 10000)
	te.id2term = make(map[int]string, 10000)
	te.counts = make(map[int]int, 10000)
	te.currCounts = make(map[int]int, 10000)
	te.testMode = testMode
	if testMode {
		return te, nil
	}
	tedb, err := csvdb.NewCircuitDB(dataDir, "terms",
		tableDefs["terms"], maxBlocks, 0, keepPeriod, unitSecs, useGzip)
	if err != nil {
		return nil, err
	}
	te.CircuitDB = tedb

	return te, nil
}

// add a new term or just count up the term count
func (te *terms) register(term string) int {
	termId, ok := te.term2Id[term]
	if !ok {
		te.maxTermId++
		termId = te.maxTermId
		te.id2term[termId] = term
		te.term2Id[term] = termId
		//te.counts[termId] += addCnt
	}
	//te.counts[termId] += addCnt
	//te.totalCount += addCnt

	return termId
}

func (te *terms) addCount(termId int, addCnt int, isNew bool) {
	te.counts[termId] += addCnt
	te.totalCount += addCnt
	if isNew {
		te.currCounts[termId] += addCnt
	}
}

func (te *terms) addCurrCount(termId int, addCnt int) {
	te.currCounts[termId] += addCnt
}

func (te *terms) getIdf(termId int) float64 {
	if te.totalCount == 0 {
		return 0
	}
	count := te.counts[termId]
	if count == 0 {
		return 0
	}
	score := math.Log(float64(te.totalCount)/float64(count)) + 1
	return score
}

func (te *terms) getCount(word string) int {
	termId, ok := te.term2Id[word]
	if !ok {
		return -1
	}
	return te.counts[termId]
}

func (te *terms) flush() error {
	if te.DataDir == "" || te.testMode {
		return nil
	}
	for termId, count := range te.currCounts {
		if count == 0 {
			continue
		}
		// inserts to buffer only
		if err := te.InsertRow(tableDefs["terms"],
			te.id2term[termId], count); err != nil {
			return err
		}
	}

	// flush buffer to the block table file with WRITE mode (not APPEND)
	if err := te.FlushOverwriteCurrentTable(); err != nil {
		return err
	}
	te.currCounts = make(map[int]int, 10000)
	return nil
}

func (te *terms) next(updated int64) error {
	// write current block to the block table
	if err := te.flush(); err != nil {
		return err
	}
	if err := te.NextBlock(updated); err != nil {
		return err
	}

	// If the new Block have data, subtruct the counts
	rows, err := te.SelectFromCurrentTable(nil, tableDefs["terms"])
	if err != nil {
		return err
	}
	if rows == nil {
		return nil
	}
	for rows.Next() {
		var term string
		var count int
		err = rows.Scan(&term, &count)
		if err != nil {
			return err
		}
		if termId, ok := te.term2Id[term]; ok {
			te.addCount(termId, -count, false)
		}
	}

	return nil
}

func (te *terms) commit(completed bool) error {
	if te.DataDir == "" {
		return nil
	}
	if err := te.flush(); err != nil {
		return err
	}

	if err := te.UpdateBlockStatus(completed); err != nil {
		return err
	}
	return nil
}

func (te *terms) load() error {
	cnt := te.CountFromStatusTable(nil)
	if cnt <= 0 {
		return nil
	}

	if err := te.LoadCircuitDBStatus(); err != nil {
		return err
	}

	rows, err := te.SelectCompletedRows(nil, nil, tableDefs["terms"])
	if err != nil {
		return err
	}
	if rows == nil {
		return nil
	}

	for rows.Next() {
		var term string
		var count int
		err = rows.Scan(&term, &count)
		if err != nil {
			return err
		}
		termId := te.register(term)
		te.addCount(termId, count, false)
	}

	trows, err := te.SelectFromCurrentTable(nil, tableDefs["terms"])
	if err != nil {
		return err
	}
	if trows == nil {
		return nil
	}
	for trows.Next() {
		var term string
		var count int
		err = trows.Scan(&term, &count)
		if err != nil {
			return err
		}
		termId := te.register(term)
		te.addCount(termId, count, true)
	}

	return nil
}

// read from specified block into a map[string]int
// mainly for testing
func (te *terms) getBlockData(blockNo int) (map[string]int, error) {
	table, err := te.GetBlockTable(blockNo)
	if err != nil {
		return nil, err
	}
	rows, err := table.SelectRows(nil, nil)
	if err != nil {
		return nil, err
	}
	if rows == nil {
		return nil, nil
	}

	counts := make(map[string]int)
	for rows.Next() {
		var term string
		var count int
		err = rows.Scan(&term, &count)
		if err != nil {
			return nil, err
		}
		counts[term] = count
	}

	return counts, nil
}

func (te *terms) getCountBorder(rate float64) int {
	n := len(te.counts)
	counts := make([]int, n)
	j := 0
	for _, cnt := range te.counts {
		counts[j] = cnt
		j++
	}
	sort.Slice(counts, func(i, j int) bool {
		return counts[i] > counts[j]
	})
	total := te.totalCount
	sum := 0
	cnt := 0
	preCnt := 0
	oldCnt := 0
	for _, cnt = range counts {
		sum += cnt
		if float64(sum)/float64(total) >= rate {
			break
		}
		if cnt < oldCnt {
			preCnt = oldCnt
		}
		oldCnt = cnt
	}
	if preCnt == 0 {
		if oldCnt > 0 {
			preCnt = oldCnt
		} else {
			preCnt = cnt
		}
	}

	return preCnt
}
