package analyzer

import (
	"fmt"
	"sort"
	"time"

	csvdb "goLogAnalyzer/csvdb"

	"github.com/pkg/errors"
)

func newNTopRecords(name string,
	n int, minScore float64, t *trans,
	isUniqMode bool, rootDir string) (*nTopRecords, error) {
	ntop := new(nTopRecords)
	ntop.n = n
	ntop.isUniqMode = isUniqMode
	ntop.minScore = minScore
	ntop.rootDir = rootDir
	ntop.name = name
	if t == nil {
		ntop.t, _ = newTrans("", 0, 0, 0, "", 1, 0)
	} else {
		ntop.t = t
	}
	ntop.initRecords()
	if rootDir != "" {
		if err := ensureDir(rootDir); err != nil {
			return nil, err
		}
		db, err := csvdb.NewCsvDB(rootDir)
		if err != nil {
			return nil, err
		}
		ntop.CsvDB = db
		if err := ntop.prepareTables(); err != nil {
			return nil, err
		}
		//if err := ntop.load(); err != nil {
		//	return nil, err
		//}
	}
	return ntop, nil
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

	if rowID > 0 && rowID <= ntop.lastRowId {
		return
	}

	ntop.lastRowId = rowID
	logr2 := new(colLogRecord)
	logr2.rowid = rowID
	logr2.score = score
	logr2.maxScore = score
	logr2.record = text
	logr2.count = 1
	maxMatchRate := 0.0
	maxMatchIdx := -1
	tran, dt := ntop.tokenizeLine(text, registerItems)
	if dt.Unix() > 0 {
		logr2.lastDate = dt.Format("01/02 15:04")
	}
	logr2.tran = tran

	if ntop.isUniqMode {
		tranlen := len(logr2.tran)
		for i, logr := range ntop.records {
			if logr == nil {
				break
			}

			rate := checkMatchRate(logr.tran, logr2.tran)
			if rate == 1.0 {
				maxMatchIdx = i
				break
			} else if rate > maxMatchRate {
				for _, tmr := range tranMatchRates {
					if tranlen >= tmr.matchLen && tmr.matchRate <= rate {
						maxMatchRate = rate
						maxMatchIdx = i
						break
					}
				}
			}
		}
	}

	if maxMatchIdx >= 0 {
		logr2.count += ntop.records[maxMatchIdx].count
		maxScore := ntop.records[maxMatchIdx].maxScore
		if logr2.maxScore < maxScore {
			logr2.maxScore = maxScore
		}
		tmpRec := new(colLogRecord)
		tmpRec.rowid = -1
		ntop.records[maxMatchIdx] = tmpRec
	}

	newRecords := make([]*colLogRecord, ntop.subN)
	newi := 0
	for i, logr := range ntop.records {
		if newi >= ntop.subN {
			break
		}
		if logr == nil {
			newRecords[newi] = logr2
			break
		}
		if logr.rowid == -1 {
			continue
		}
		if score >= logr.score {
			newRecords[newi] = logr2
			for j := i; j < len(ntop.records); j++ {
				if ntop.records[j] == nil {
					break
				}
				if ntop.records[j].rowid != -1 {
					newi++
					if newi >= len(newRecords) {
						break
					}
					newRecords[newi] = ntop.records[j]
				}
			}
			break
		}
		newRecords[newi] = logr
		newi++
	}

	ntop.records = newRecords
}

func (ntop *nTopRecords) getRareTerms(nItems int, records []*colLogRecord) []string {
	ntopi := newTopNItems(nItems)
	items := ntop.t.items
	for _, record := range records {
		for _, itemID := range record.tran {
			cnt := items.counts[itemID]
			if cnt < cMinNTopItemCount {
				continue
			}
			term := items.getWord(itemID)
			if len(term) < cMinNTopItemTermLen || isNumeric(term) {
				continue
			}
			score := items.calcAdjScore(itemID)
			ntopi.register(itemID, score)
		}
	}
	terms := make([]string, nItems)
	for j, itemID := range ntopi.itemIDs {
		terms[j] = items.getWord(itemID)
	}
	return terms
}

func (ntop *nTopRecords) getRecords() []*colLogRecord {
	cnt := 0
	for i := 0; i < ntop.n; i++ {
		if ntop.records[i] == nil {
			break
		}
		cnt++
	}
	return ntop.records[0:cnt]
}

// get records sorted by count
func (ntop *nTopRecords) getRecords2() []*colLogRecord {
	cnt := 0
	records := make([]*colLogRecord, len(ntop.records))
	for i := 0; i < ntop.n; i++ {
		if ntop.records[i] == nil {
			break
		}
		records[i] = ntop.records[i]
		cnt++
	}
	records = records[0:cnt]
	sort.Slice(records,
		func(i, j int) bool {
			return records[i].count < records[j].count || (records[i].count == records[j].count && records[i].score > records[j].score)
		})
	return records[0:cnt]
}

func (ntop *nTopRecords) getTableName() string {
	return fmt.Sprintf("topn_%s", ntop.name)
}

func (ntop *nTopRecords) prepareTables() error {
	ntopTable, err := ntop.CreateTableIfNotExists(ntop.getTableName(),
		tableDefs["lastTopN"], false, cDefaultBuffSize)
	if err != nil {
		return err
	}
	ntop.ntopTable = ntopTable
	return nil
}

func (ntop *nTopRecords) save() error {
	if err := ntop.ntopTable.Truncate(); err != nil {
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
		if err := ntop.ntopTable.InsertRow([]string{"rowid", "score", "maxScore",
			"record", "terms", "count", "lastDate"},
			r.rowid, r.score, r.maxScore, r.record, terms, r.count,
			r.lastDate); err != nil {
			return err
		}
	}
	if err := ntop.ntopTable.Flush(); err != nil {
		return err
	}

	return nil
}

func (ntop *nTopRecords) load(lastRowID int64,
	maxRowIDs int, registerItems bool) error {
	if !ntop.TableExists(ntop.getTableName()) {
		return nil
	}

	ntopTable, err := ntop.GetTable(ntop.getTableName())
	if err != nil {
		return err
	}
	defer ntopTable.Close()

	rows, err := ntopTable.SelectRows(nil,
		[]string{"rowid", "score", "maxScore", "record", "terms", "count", "lastDate"})
	if err != nil {
		return err
	}

	lastDate := ""
	ntopIdx := 0
	for rows.Next() {
		r := new(colLogRecord)
		terms := make([]string, 0)
		if err := rows.Scan(&r.rowid, &r.score, &r.maxScore, &r.record, &terms,
			&r.count, &lastDate); err != nil {
			return errors.WithStack(err)
		}

		r.lastDate = lastDate
		tran, _ := ntop.tokenizeLine(r.record, registerItems)
		r.tran = tran
		ntop.records[ntopIdx] = r
		ntopIdx++
		if ntop.lastRowId < r.rowid {
			ntop.lastRowId = r.rowid
		}
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

func (ntop *nTopRecords) getString(msg string, recordsToShow, nRareTerms int) (string, float64, error) {
	out := fmt.Sprintf("%s\n", msg)
	out += " count | score   | maxScore | rowID      | text\n"
	out += "-------+---------+----------+------------+-------\n"
	topScore := 0.0
	records := ntop.getRecords2()
	for i, logr := range records {
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
		outRec := fmt.Sprintf(" %5d |%8.2f |%8.2f  | %10d | %s", logr.count,
			logr.score, logr.maxScore, logr.rowid, te)

		out += fmt.Sprintf("%s\n", outRec)

		if logr.score == 0 {
			break
		}

		if i+1 >= recordsToShow {
			break
		}
	}

	out += "\nRare words:\n"
	for i, term := range ntop.getRareTerms(nRareTerms, records) {
		if term == "" {
			break
		}
		if i == 0 {
			out += term
		} else {
			out += fmt.Sprintf(" %s", term)
		}
	}
	out += "\n\n"
	return out, topScore, nil
}

func (ntop *nTopRecords) getHtmlTable(recordsToShow int) (string, float64, error) {
	//println(a.trans.items.totalCount)
	out := ""
	out += "<table border=1 ~~~ style='table-layout:fixed;width:100%;'>"
	out += "<tr><td width=4%>count</td><td width=10%>lastUpdate</td><td width=6%>score</td><td width=10%>rowID</td><td>text</td></tr>"
	topScore := 0.0
	for i, logr := range ntop.getRecords2() {
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

		out += fmt.Sprintf("<tr><td>%d</td><td>%s</td><td>%8.2f</td><td>%10d</td><td>%s</td></tr>",
			logr.count, logr.lastDate, logr.score, logr.rowid, te)

		if logr.score == 0 {
			break
		}

		if i+1 >= recordsToShow {
			break
		}
	}
	out += "</table>"
	return out, topScore, nil
}

func (ntop *nTopRecords) getJson(msg string, recordsToShow int) ([]*nTopOutRec,
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
		outRec.lastDate = logr.lastDate
		outRec.record = te
		outRecs = append(outRecs, outRec)

		if i+1 >= recordsToShow {
			break
		}
	}
	return outRecs, topScore, nil
}
