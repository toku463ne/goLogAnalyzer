package analyzer

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/pkg/errors"
	csvdb "github.com/toku463ne/goLogAnalyzer/csvdb"
)

func newNTopRecords(n int,
	minScore float64, t *trans,
	isUniqMode bool) *nTopRecords {
	ntop := new(nTopRecords)
	ntop.n = n
	ntop.isUniqMode = isUniqMode
	ntop.minScore = minScore
	ntop.withDiff = false
	ntop.t = t
	ntop.initRecords()
	return ntop
}

func (ntop *nTopRecords) initRecords() {
	ntop.subN = cNTopMultiplier * ntop.n
	ntop.records = make([]*colLogRecord, ntop.subN)
}

func (ntop *nTopRecords) tokenizeLine(text string, registerItems bool) ([]int, time.Time) {
	_, tran, dt, _ := ntop.t.tokenizeLine(text, nil, nil, registerItems)
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
	sort.Slice(tran, func(i, j int) bool { return tran[i] < tran[j] })
	return tran, dt
}

func (ntop *nTopRecords) register(rowID int64, score float64, text string, registerItems bool) {
	if ntop.minScore > 0 && score <= ntop.minScore && ntop.memberCnt >= ntop.subN {
		return
	}
	newRecords := make([]*colLogRecord, ntop.subN)
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
		tran, dt := ntop.tokenizeLine(text, registerItems)
		if dt.Unix() > 0 {
			logr2.dates = append(logr2.dates, dt.Format("01-02T15:04:05"))
		}
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
		if ntop.withDiff {
			ntop.diff.register(logr2.rowid, logr2.score, logr2.record, registerItems)
		}
		return
	}

	if maxMatchIdx >= 0 {
		logr2.count += ntop.records[maxMatchIdx].count
		logr2.dates = append(logr2.dates, ntop.records[maxMatchIdx].dates...)
	}

	newi := 0
	for i, logr := range ntop.records {
		if newi >= ntop.subN {
			break
		}
		if logr == nil {
			newRecords[newi] = logr2
			if ntop.withDiff {
				ntop.diff.register(logr2.rowid, logr2.score, logr2.record, registerItems)
			}
			break
		}
		if score > logr.score {
			newRecords[newi] = logr2
			if ntop.withDiff {
				ntop.diff.register(logr2.rowid, logr2.score, logr2.record, registerItems)
			}
			newi++
			for j := i; j < ntop.subN; j++ {
				if newi >= ntop.subN {
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
	/*
		if !ntop.isUniqMode {
			return ntop.records
		}
			recs := make([]*colLogRecord, len(ntop.records))
			copy(recs, ntop.records)
			sort.Slice(recs, func(i, j int) bool {
					if recs[i] == nil {
						return false
					}
					if recs[j] == nil {
						return true
					}
					if recs[i].count == recs[j].count {
						return recs[i].score > recs[j].score
					}
					return recs[i].count < recs[j].count
				})
	*/

	cnt := 0
	for i := 0; i < ntop.n; i++ {
		if ntop.records[i] == nil {
			break
		}
		cnt++
	}
	return ntop.records[0:cnt]
}

func (ntop *nTopRecords) getDiffRecords() []*colLogRecord {
	minScore := ntop.getMinScore()
	cnt := 0
	if ntop == nil || ntop.diff == nil {
		return nil
	}
	if len(ntop.diff.records) == 0 {
		return nil
	}
	for _, r := range ntop.diff.records {
		if r == nil {
			break
		}
		if ntop.n <= cnt || r.score < minScore {
			break
		}
		cnt++
	}
	if cnt == 0 {
		return nil
	}
	return ntop.diff.records[0:cnt]
}

func (ntop *nTopRecords) save(rootDir string) error {
	db, err := csvdb.NewCsvDB(rootDir)
	if err != nil {
		return err
	}
	if db.TableExists("lastTopN") {
		if err := db.DropTable("lastTopN"); err != nil {
			return err
		}
	}
	ntopTable, err := db.CreateTableIfNotExists("lastTopN",
		tableDefs["lastTopN"], false, cDefaultBuffSize)
	if err != nil {
		return err
	}

	for _, r := range ntop.records {
		if r == nil {
			break
		}
		terms := make([]string, len(r.tran))
		for i, t := range r.tran {
			terms[i] = ntop.t.items.getWord(t)
		}
		if err := ntopTable.InsertRow([]string{"rowid", "score",
			"record", "terms", "count", "lastNdates"},
			r.rowid, r.score, r.record, terms, r.count,
			strings.Join(r.dates, ";")); err != nil {
			return err
		}
	}
	if err := ntopTable.Flush(); err != nil {
		return err
	}

	return nil
}

func (ntop *nTopRecords) load(rootDir string, lastRowID int64,
	maxRowIDs int, registerItems bool) error {
	db, err := csvdb.NewCsvDB(rootDir)
	if err != nil {
		return err
	}
	if !db.TableExists("lastTopN") {
		return nil
	}

	ntopTable, err := db.GetTable("lastTopN")
	if err != nil {
		return err
	}

	rows, err := ntopTable.SelectRows(nil,
		[]string{"rowid", "score", "record", "terms", "count", "lastNdates"})
	if err != nil {
		return err
	}

	lastNdatesStr := ""
	ntopIdx := 0
	for rows.Next() {
		r := new(colLogRecord)
		terms := make([]string, 0)
		if err := rows.Scan(&r.rowid, &r.score, &r.record, &terms,
			&r.count, &lastNdatesStr); err != nil {
			return errors.WithStack(err)
		}

		rowIdDiff := 0
		if r.rowid > lastRowID {
			rowIdDiff = int(cMaxRowID) - int(r.rowid) + int(lastRowID)
		} else {
			rowIdDiff = int(lastRowID) - int(r.rowid)
		}
		if rowIdDiff > maxRowIDs {
			continue
		}

		r.dates = strings.Split(lastNdatesStr, ";")
		tran, _ := ntop.tokenizeLine(r.record, registerItems)
		r.tran = tran
		ntop.records[ntopIdx] = r
		ntopIdx++
	}

	if ntopIdx >= 0 {
		ntop.withDiff = true
		ntop.diff = newNTopRecords(ntop.n, ntop.minScore, ntop.t, ntop.isUniqMode)
	}
	return nil
}

func (ntop *nTopRecords) getLen() int {
	cnt := 0
	if ntop == nil || ntop.records == nil {
		return 0
	}
	for _, l := range ntop.records {
		if l == nil {
			return cnt
		}
		cnt++
	}
	return cnt
}

func (ntop *nTopRecords) getMinScore() float64 {
	minScore := 0.0
	i := 0
	for _, l := range ntop.records {
		if i >= ntop.n {
			return minScore
		}
		if l == nil {
			return minScore
		}
		if minScore == 0.0 || (l.score > 0 && l.score < minScore) {
			minScore = l.score
		}
		i++
	}
	return minScore
}

func (ntop *nTopRecords) getMaxCount() int {
	maxCnt := 0
	for _, l := range ntop.records {
		if l == nil {
			return maxCnt
		}
		if maxCnt == 0 || l.count > maxCnt {
			maxCnt = l.count
		}
	}
	return maxCnt
}

func (ntop *nTopRecords) nTop2string(msg string, recordsToShow int) (string, float64, error) {
	out := fmt.Sprintf("%s\n", msg)
	out += " count | score   | rowID      | text\n"
	out += "-------+---------+------------+-------\n"
	topScore := 0.0
	for i, logr := range ntop.getRecords() {
		if logr == nil {
			break
		}
		if topScore == 0 {
			topScore = logr.score
		}
		te := ""
		if len(logr.record) > cMaxCharsToShowInTopN {
			te = string([]rune(logr.record)[:cMaxCharsToShowInTopN])
		} else {
			te = logr.record
		}
		outRec := fmt.Sprintf(" %5d |%8.2f | %10d | %s", logr.count, logr.score,
			logr.rowid, te)

		out += fmt.Sprintf("%s\n", outRec)

		if logr.score == 0 {
			break
		}

		if i+1 >= recordsToShow {
			break
		}
	}
	return out, topScore, nil
}

func (ntop *nTopRecords) nTop2html(msg string, recordsToShow int) (string, float64, error) {
	//println(a.trans.items.totalCount)
	out := ""
	topScore := 0.0
	out += fmt.Sprintf("<b>%s</b><br>", msg)
	out += "<table border=1 ~~~ style='table-layout:fixed;width:100%;'>"
	out += "<tr><td width=4%>count</td><td width=10%>dates</td><td width=6%>score</td><td width=6%>rowID</td><td>text</td></tr>"
	for i, logr := range ntop.getRecords() {
		if logr == nil {
			break
		}
		if topScore == 0 {
			topScore = logr.score
		}
		te := ""
		if len(logr.record) > cMaxCharsToShowInTopN {
			te = string([]rune(logr.record)[:cMaxCharsToShowInTopN])
		} else {
			te = logr.record
		}
		dates := logr.dates
		if len(dates) > 10 {
			dates = dates[len(dates)-10:]
		}
		dates2 := make([]string, len(dates))
		k := 0
		for j := len(dates) - 1; j >= 0; j-- {
			dates2[k] = dates[j]
			k++
		}

		out += fmt.Sprintf("<tr><td>%d</td><td>%s</td><td>%8.2f</td><td>%10d</td><td>%s</td></tr>",
			logr.count, strings.Join(dates2, ", "), logr.score, logr.rowid, te)

		if logr.score == 0 {
			break
		}

		if i+1 >= recordsToShow {
			break
		}
	}
	out += "</table><br>\n"

	return out, topScore, nil
}

func (ntop *nTopRecords) nTop2json(msg string, recordsToShow int) ([]*nTopOutRec,
	float64, error) {
	topScore := 0.0
	outRecs := make([]*nTopOutRec, 0)
	for i, logr := range ntop.getRecords() {
		if logr == nil {
			break
		}
		if topScore == 0 {
			topScore = logr.score
		}
		if logr.score == 0 {
			break
		}
		te := ""
		if len(logr.record) > cMaxCharsToShowInTopN {
			te = string([]rune(logr.record)[:cMaxCharsToShowInTopN])
		} else {
			te = logr.record
		}

		outRec := new(nTopOutRec)
		outRec.rowid = logr.rowid
		outRec.count = logr.count
		outRec.score = logr.score
		outRec.record = te
		outRecs = append(outRecs, outRec)

		if i+1 >= recordsToShow {
			break
		}
	}
	return outRecs, topScore, nil
}
