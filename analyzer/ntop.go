package analyzer

import "sort"

func newNTopRecords(n int,
	minScore float64, t *trans,
	isUniqMode bool) *nTopRecords {
	ntop := new(nTopRecords)
	ntop.n = n
	ntop.isUniqMode = isUniqMode
	ntop.minScore = minScore
	m := 1
	if isUniqMode {
		m = cNTopMultiplier
	}
	ntop.records = make([]*colLogRecord, n*m)
	ntop.t = t
	return ntop
}

func (ntop *nTopRecords) register(rowID int64, score float64, text string,
	isTest bool) {
	if ntop.minScore > 0 && score <= ntop.minScore && ntop.memberCnt >= ntop.n {
		return
	}
	m := 1
	if ntop.isUniqMode {
		m = cNTopMultiplier
	}
	newRecords := make([]*colLogRecord, ntop.n*m)
	logr2 := new(colLogRecord)
	logr2.rowid = rowID
	logr2.score = score
	logr2.record = text
	logr2.count = 1
	maxMatchRate := 0.0
	maxMatchIdx := -1
	maxGreaterMatchIdx := -1

	if ntop.isUniqMode {
		tran, _, _ := ntop.t.tokenizeLine(text, nil, nil, isTest)
		sort.Slice(tran, func(i, j int) bool { return tran[i] < tran[j] })
		logr2.tran = tran
		tranlen := len(logr2.tran)
		for i, logr := range ntop.records {
			if logr == nil {
				break
			}

			rate := checkMatchRate(logr.tran, logr2.tran)
			if rate > maxMatchRate {
				for j, l := range nTopBaseMTokens {
					if tranlen >= l {
						if nTopMatchRates[j] <= rate {
							maxMatchRate = rate
							maxMatchIdx = i
							if score > logr.score {
								maxGreaterMatchIdx = i
							}
							break
						}
					} else {
						break
					}
				}
			}
		}
	}
	// no change to ranking but increment the count
	if maxGreaterMatchIdx == -1 && maxMatchIdx >= 0 {
		ntop.records[maxMatchIdx].count++
		return
	}

	if maxMatchIdx >= 0 {
		logr2.count += ntop.records[maxMatchIdx].count
	}

	newi := 0
	for i, logr := range ntop.records {
		if newi >= ntop.n {
			break
		}
		if logr == nil {
			newRecords[newi] = logr2
			break
		}
		if score > logr.score {
			newRecords[newi] = logr2
			newi++
			for j := i; j < ntop.n; j++ {
				if newi >= ntop.n {
					break
				}
				if ntop.records[j] == nil {
					break
				}
				if j != maxGreaterMatchIdx {
					newRecords[newi] = ntop.records[j]
					newi++
				}
			}
			break
		} else {
			if i == maxMatchIdx {
				continue
			}
			newRecords[newi] = logr
		}
		newi++
	}

	ntop.records = newRecords
}

func (ntop *nTopRecords) getRecords() []*colLogRecord {
	if !ntop.isUniqMode {
		return ntop.records
	}

	sort.Slice(ntop.records, func(i, j int) bool {
		if ntop.records[i] == nil {
			return false
		}
		if ntop.records[j] == nil {
			return true
		}
		if ntop.records[i].count == ntop.records[j].count {
			return ntop.records[i].score > ntop.records[j].score
		}
		return ntop.records[i].count < ntop.records[j].count
	})
	return ntop.records[0:ntop.n]
}
