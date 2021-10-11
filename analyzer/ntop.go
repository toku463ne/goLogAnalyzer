package analyzer

import (
	"sort"
	"time"

	"github.com/pkg/errors"
	csvdb "github.com/toku463ne/goCsvDb"
)

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
	logr2.dates = make([]string, 0)
	maxMatchRate := 0.0
	maxMatchIdx := -1
	maxGreaterMatchIdx := -1

	if ntop.isUniqMode {
		_, tran, dt, _ := ntop.t.tokenizeLine(text, nil, nil, isTest)
		if dt.Year() == 0 {
			y := time.Now().Year()
			m := dt.Month()
			d := dt.Day()
			if m == 0 {
				m = time.Now().Month()
			}
			if d == 0 {
				d = time.Now().Day()
			}
			dt = time.Date(y, m, d, dt.Hour(), dt.Minute(), dt.Second(), 0, dt.Location())
		}
		if dt.Unix() > 0 {
			logr2.dates = append(logr2.dates, dt.Format("01-02T15:04:05"))
		}
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
		ntop.records[maxMatchIdx].dates = append(ntop.records[maxMatchIdx].dates, logr2.dates...)
		return
	}

	if maxMatchIdx >= 0 {
		logr2.count += ntop.records[maxMatchIdx].count
		logr2.dates = append(logr2.dates, ntop.records[maxMatchIdx].dates...)
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

func (ntop *nTopRecords) save(rootDir string) error {
	db, err := csvdb.NewCsvDB(rootDir)
	if err != nil {
		return err
	}
	if err := db.DropTable("lastTopN"); err != nil {
		return err
	}

	ntopTable, err := db.CreateTableIfNotExists("lastTopN",
		tableDefs["lastTopN"], false, cDefaultBuffSize)
	if err != nil {
		return err
	}

	for _, r := range ntop.records {
		terms := make([]string, len(r.tran))
		for i, t := range r.tran {
			terms[i] = ntop.t.items.getWord(t)
		}
		if err := ntopTable.InsertRow([]string{"rowid", "score", "record", "terms", "count"},
			r.rowid, r.score, r.record, terms, r.count); err != nil {
			return err
		}
	}

	return nil
}

func (ntop *nTopRecords) load(rootDir string) error {
	db, err := csvdb.NewCsvDB(rootDir)
	if err != nil {
		return err
	}
	if !db.TaleExists("lastTopN") {
		return nil
	}

	ntopTable, err := db.GetTable("lastTopN")
	if err != nil {
		return err
	}

	rows, err := ntopTable.SelectRows(nil,
		[]string{"rowid", "score", "record", "terms", "count"})
	if err != nil {
		return err
	}
	records := make([]*colLogRecord, 0)
	for rows.Next() {
		r := new(colLogRecord)
		terms := make([]string, 0)
		if err := rows.Scan(&r.rowid, &r.score, &r.record, &terms, &r.count); err != nil {
			return errors.WithStack(err)
		}
		tran := make([]int, len(terms))
		for i, t := range terms {
			tran[i] = ntop.t.items.getItemID(t)
		}
		r.tran = tran
		records = append(records, r)
	}
	ntop.records = records

	return nil
}
