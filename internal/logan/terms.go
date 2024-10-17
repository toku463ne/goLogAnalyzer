package logan

import "math"

type terms struct {
	maxTermId  int
	term2Id    map[string]int
	id2term    map[int]string
	counts     map[int]int
	totalCount int
}

func newTerms() *terms {
	te := new(terms)
	te.maxTermId = 0
	te.term2Id = make(map[string]int, 10000)
	te.id2term = make(map[int]string, 10000)
	te.counts = make(map[int]int, 10000)
	return te
}

// add a new term or just count up the term count
func (te *terms) register(term string, addCnt int) int {
	termId, ok := te.term2Id[term]
	if !ok {
		te.maxTermId++
		termId = te.maxTermId
		te.id2term[termId] = term
		te.term2Id[term] = termId
		te.counts[termId] += addCnt
	}
	te.counts[termId] += addCnt
	te.totalCount += addCnt
	return termId
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
