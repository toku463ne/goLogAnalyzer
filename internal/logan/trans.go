package logan

import (
	"fmt"
	"goLogAnalyzer/pkg/utils"
	"regexp"
	"strings"
	"time"
)

type trans struct {
	te                  *terms
	lt                  *logTree
	lgs                 *logGroups
	customLogGroups     []string
	replacer            *strings.Replacer
	logFormatRe         *regexp.Regexp
	timestampLayout     string
	timestampPos        int
	messagePos          int
	readOnly            bool
	totalLines          int
	filterRe            []*regexp.Regexp
	xFilterRe           []*regexp.Regexp
	termCountBorderRate float64
	termCountBorder     int
	keywords            map[string]string
	keyTermIds          map[int]string
	ignorewords         map[string]string
	keepUnit            int
	countByBlock        int
	maxCountByBlock     int
	currRetentionPos    int
}

func newTrans(dataDir, logFormat, timestampLayout string,
	maxBlocks, blockSize,
	keepUnit int, keepPeriod int64,
	searchRegex, exludeRegex []string,
	_keywords []string, _ignorewords []string, _customLogGroups []string,
	useGzip, readOnly bool) (*trans, error) {
	tr := new(trans)
	tr.te = newTerms()
	tr.lt = newLogTree(0)
	lgs, err := newLogGroups(dataDir, maxBlocks, blockSize, keepUnit, keepPeriod, useGzip)
	if err != nil {
		return nil, err
	}
	tr.lgs = lgs
	tr.replacer = getDelimReplacer()
	tr._parseLogFormat(logFormat)
	tr.timestampLayout = timestampLayout
	tr.readOnly = readOnly
	tr.totalLines = 0
	tr._setFilters(searchRegex, exludeRegex)

	tr.keywords = make(map[string]string)
	tr.ignorewords = make(map[string]string)
	tr.keyTermIds = make(map[int]string)
	for _, word := range _keywords {
		tr.keywords[word] = ""
	}
	for _, word := range _ignorewords {
		tr.ignorewords[word] = ""
	}
	tr.keepUnit = keepUnit
	tr.currRetentionPos = 0

	return tr, nil
}

func (tr *trans) _parseLogFormat(logFormat string) {
	re := regexp.MustCompile(logFormat)
	names := re.SubexpNames()
	tr.timestampPos = -1
	tr.messagePos = -1
	for i, name := range names {
		switch {
		case name == "timestamp":
			tr.timestampPos = i
		case name == "message":
			tr.messagePos = i
		}
	}
	tr.logFormatRe = re
}

func (tr *trans) _setFilters(searchRegex, exludeRegex []string) {
	tr.filterRe = make([]*regexp.Regexp, 0)
	for _, s := range searchRegex {
		tr.filterRe = append(tr.filterRe, utils.GetRegex(s))
	}

	tr.xFilterRe = make([]*regexp.Regexp, 0)
	for _, s := range exludeRegex {
		tr.xFilterRe = append(tr.xFilterRe, utils.GetRegex(s))
	}
}

// filtering text
func (tr *trans) _match(text string) bool {
	if tr.filterRe == nil && tr.xFilterRe == nil {
		return true
	}

	b := []byte(text)
	matched := true
	for _, filterRe := range tr.filterRe {
		if !filterRe.Match(b) {
			matched = false
			break
		}
	}
	if !matched {
		return false
	}

	matched = false
	for _, xFilterRe := range tr.xFilterRe {
		if xFilterRe.Match(b) {
			matched = true
			break
		}
	}
	return !matched
}

func (tr *trans) setMaxBlocks(maxBlocks int) {
	if tr.lgs != nil {
		tr.lgs.SetMaxBlocks(maxBlocks)
	}
}
func (tr *trans) setBlockSize(blockSize int) {
	if tr.lgs != nil {
		tr.lgs.SetBlockSize(blockSize)
	}
}

func (tr *trans) parseLine(line string) (string, int64, int) {
	var lastdt time.Time
	var err error
	lastUpdate := int64(0)
	retentionPos := -1

	if tr.timestampPos >= 0 || tr.messagePos >= 0 {
		ma := tr.logFormatRe.FindStringSubmatch(line)
		if len(ma) > 0 {
			if tr.timestampPos >= 0 && tr.timestampLayout != "" && len(ma) > tr.timestampPos {
				lastdt, err = utils.Str2date(tr.timestampLayout, ma[tr.timestampPos])
				switch tr.keepUnit {
				case utils.CFreqHour:
					retentionPos = lastdt.Year()*100000 + lastdt.YearDay()*100 + lastdt.Hour()
				case utils.CFreqDay:
					retentionPos = lastdt.Year()*1000 + lastdt.YearDay()
				default:
					retentionPos = 0
				}
			}
			if err == nil {
				lastUpdate = lastdt.Unix()
			}
			if lastUpdate > 0 {
				if tr.messagePos >= 0 && len(ma) > tr.messagePos {
					line = ma[tr.messagePos]
				}
			}
		}
	}
	return line, lastUpdate, retentionPos
}

// convert line to token list and register to tr.te only once
// returns tokens, excludesMap(terms that replaced with *)
func (tr *trans) toTokens(line string, addCnt int,
	useTermBorder bool, needDisplayString bool,
) ([]int, string) {
	line = tr.replacer.Replace(line)
	words := strings.Split(line, " ")
	tokens := make([]int, 0)
	uniqTokens := make(map[int]bool, 0)
	excludesMap := make(map[string]bool)

	termId := -1

	for _, w := range words {
		if w == "" {
			continue
		}

		if _, ok := tr.ignorewords[w]; ok {
			excludesMap[w] = true
			w = "*"
		}
		_, keyOK := tr.keywords[w]
		if _, ok := enStopWords[w]; ok {
			if !keyOK {
				w = "*"
			}
		}

		word := strings.ToLower(w)
		lenw := len(word)
		if lenw > 1 && string(word[lenw-1]) == "." {
			word = word[:lenw-1]
		}

		if keyOK || len(word) > 2 {
			if !keyOK && utils.IsInt(word) && len(word) > cMaxNumDigits {
				excludesMap[word] = true
				continue
			}
			termId = tr.te.register(word, 0)
			if !uniqTokens[termId] {
				tr.te.counts[termId] += addCnt
			}
			uniqTokens[termId] = true

			if useTermBorder && tr.termCountBorder > tr.te.counts[termId] {
				termId = cAsteriskItemID
			}
			tokens = append(tokens, termId)
			if keyOK {
				tr.keyTermIds[termId] = ""
			}
		} else if word == "*" {
			tokens = append(tokens, cAsteriskItemID)
		} else {
			excludesMap[word] = true
		}
	}

	if needDisplayString {
		for word := range excludesMap {
			// Use capturing groups to capture delimiters and replace only the word
			pattern := `(?i)(^|` + cDelimiters + `)(` + regexp.QuoteMeta(word) + `)($|` + cDelimiters + `)`
			reg := regexp.MustCompile(pattern)
			line = reg.ReplaceAllString(line, `$1`+"*"+`$3`)
		}
		// Combine multiple consecutive "*" into a single "*"
	} else {
		line = ""
	}

	return tokens, line
}

// analyze the line and
func (tr *trans) tokenizeLine(line string, addCnt int, registerLogGroup bool) {
	if !tr._match(line) {
		return
	}
	line, updated, retentionPos := tr.parseLine(line)

	// we scan the logs 2 times
	if registerLogGroup {
		// register logGroups => 2nd round
		tokens, displayString := tr.toTokens(line, addCnt, true, true)
		tr.lgs.registerLogTree(tokens, addCnt, displayString, updated, updated, true, retentionPos)
	} else {
		// register terms => 1st round
		tr.toTokens(line, addCnt, false, false)
		if tr.currRetentionPos > 0 && retentionPos > tr.currRetentionPos {
			if tr.countByBlock > tr.maxCountByBlock {
				tr.maxCountByBlock = tr.countByBlock
			}
			tr.countByBlock = 0
		}
		tr.countByBlock++
		tr.totalLines++
	}
	tr.currRetentionPos = retentionPos

}

func (tr *trans) commit(completed bool) error {
	if tr.readOnly {
		return nil
	}
	if err := tr.lgs.commit(completed); err != nil {
		return err
	}
	return nil
}

func (tr *trans) close() {
	if tr.lgs.CircuitDB != nil {
		tr.lgs = nil
	}
}

func (tr *trans) rebuildLogTree(termCountBorder int) *logTree {
	lt := tr.lgs.lt
	newTree := newLogTree(lt.depth)
	lt.rebuildHelper(newTree, tr.te, termCountBorder)
	return newTree
}

func (tr *trans) loadLogGroups() error {
	lgs := tr.lgs
	if lgs.DataDir == "" {
		return nil
	}
	cnt := lgs.CountFromStatusTable(nil)
	if cnt <= 0 {
		return nil
	}

	if err := lgs.LoadCircuitDBStatus(); err != nil {
		return err
	}

	if err := lgs.readDisplayStrings(); err != nil {
		return err
	}

	rows, err := lgs.SelectRows(nil, nil, tableDefs["logGroups"])
	if err != nil {
		return err
	}
	if rows == nil {
		return nil
	}

	ds := lgs.displayStrings
	for rows.Next() {
		var groupIdstr string
		var retentionPos int
		var count int
		var created int64
		var updated int64
		err = rows.Scan(&groupIdstr, &retentionPos, &count, &created, &updated)
		if err != nil {
			return err
		}
		groupId, err := utils.Base36ToInt64(groupIdstr)
		if err != nil {
			return fmt.Errorf("error parsing %s to int64", groupIdstr)
		}

		line := ds[groupId]
		tokens, displayString := tr.toTokens(line, count, true, true)
		tr.lgs.registerLogTree(tokens, count, displayString, updated, updated, true, retentionPos)

		//lgs._registerLg(lgs.alllg, groupId, count, displayString, created, updated)
	}
	return nil
}

func (tr *trans) next() error {
	lgs := tr.lgs
	if tr.readOnly {
		return nil
	}
	// write the current block
	if err := lgs.flush(); err != nil {
		return err
	}

	// clear "current" logGroup
	lgs.curlg = make(map[int64]*logGroup)

	// in case the block table already exists and will be overrided
	// we subtract counts in the block table from total item counts
	rows, err := lgs.SelectFromCurrentTable(nil, tableDefs["logGroups"])
	if err != nil {
		return err
	}
	if rows == nil {
		return nil
	}

	ds := lgs.displayStrings
	for rows.Next() {
		var groupIdstr string
		var retentionPos int
		var count int
		var created int64
		var updated int64
		err = rows.Scan(&groupIdstr, &retentionPos, &count, &created, &updated)
		if err != nil {
			return err
		}
		groupId, err := utils.Base36ToInt64(groupIdstr)
		if err != nil {
			return fmt.Errorf("error parsing %s to int64", groupIdstr)
		}
		if lg, ok := lgs.alllg[groupId]; ok {
			lg.count -= count
		} else {
			line := ds[groupId]
			// will reach an existing logGrouup using termCountBorder
			tokens, _ := tr.toTokens(line, count, true, false)
			_groupId := tr.lgs.lt.search(tokens)
			if groupId == _groupId {
				lg.count -= count
			}
		}
	}

	return nil
}
