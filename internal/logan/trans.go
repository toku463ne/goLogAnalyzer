package logan

import (
	"fmt"
	"goLogAnalyzer/pkg/utils"
	"math"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

type trans struct {
	te              *terms
	lt              *logTree
	lgs             *logGroups
	customLogGroups []string
	replacer        *strings.Replacer
	logFormatRe     *regexp.Regexp
	timestampLayout string
	useUtcTime      bool

	timestampPos        int
	messagePos          int
	readOnly            bool
	totalLines          int
	filterRe            []*regexp.Regexp
	xFilterRe           []*regexp.Regexp
	termCountBorderRate float64
	termCountBorder     int
	minMatchRate        float64
	keywords            map[string]string
	keyTermIds          map[int]string
	ignorewords         map[string]string
	unitSecs            int64
	countByBlock        int
	maxCountByBlock     int
	currRetentionPos    int64
	separators          string
}

func newTrans(dataDir, logFormat, timestampLayout string,
	useUtcTime bool,
	maxBlocks, blockSize int,
	unitSecs int64, keepPeriod int64,
	termCountBorderRate float64,
	termCountBorder int,
	minMatchRate float64,
	searchRegex, exludeRegex []string,
	_keywords []string, _ignorewords []string,
	_customLogGroups []string,
	separators string,
	useGzip, readOnly bool) (*trans, error) {
	tr := new(trans)

	// don't need blockSize for terms because the rotation follows trans.next()
	te, err := newTerms(dataDir, maxBlocks, unitSecs, keepPeriod, useGzip)
	if err != nil {
		return nil, err
	}
	tr.te = te

	tr.lt = newLogTree(0)
	// don't need blockSize for terms because the rotation follows trans.next()
	lgs, err := newLogGroups(dataDir, maxBlocks, unitSecs, keepPeriod, useGzip)
	if err != nil {
		return nil, err
	}
	tr.lgs = lgs
	tr.replacer = getDelimReplacer(separators)
	tr._parseLogFormat(logFormat)
	tr.timestampLayout = timestampLayout
	tr.termCountBorder = termCountBorder
	tr.termCountBorderRate = termCountBorderRate
	tr.minMatchRate = minMatchRate
	tr.useUtcTime = useUtcTime
	tr.separators = separators
	tr.readOnly = readOnly
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
	tr.unitSecs = unitSecs
	tr.maxCountByBlock = blockSize
	tr.initCounters()
	return tr, nil
}

func (tr *trans) initCounters() {
	tr.currRetentionPos = 0
	tr.totalLines = 0
	tr.countByBlock = 0
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
	tr.maxCountByBlock = blockSize
}

func (tr *trans) parseLine(line string, updated int64) (string, int64, int64) {
	var lastdt time.Time
	var err error
	lastUpdate := int64(0)
	retentionPos := int64(-1)

	if tr.timestampPos >= 0 || tr.messagePos >= 0 {
		ma := tr.logFormatRe.FindStringSubmatch(line)
		if len(ma) > 0 {
			if tr.timestampPos >= 0 && tr.timestampLayout != "" && len(ma) > tr.timestampPos {
				if tr.useUtcTime {
					lastdt, err = utils.Str2Timestamp(tr.timestampLayout, ma[tr.timestampPos])
				} else {
					// system time zone
					lastdt, err = utils.Str2date(tr.timestampLayout, ma[tr.timestampPos])
				}
				if err == nil {
					lastUpdate = lastdt.Unix()
				}

				retentionPos = int64(math.Floor(float64(lastUpdate)/float64(tr.unitSecs))) * tr.unitSecs
			}

			if lastUpdate > 0 {
				if tr.messagePos >= 0 && len(ma) > tr.messagePos {
					line = ma[tr.messagePos]
				}
			}
		}
	}
	if lastUpdate == 0 {
		lastUpdate = updated
	}
	return line, lastUpdate, retentionPos
}

// convert line to list of tokens and register to tr.te.
// returns tokens, displayString and error
func (tr *trans) toTokens(line string, addCnt int,
	useTermBorder, needDisplayString, onlyCurrTerms bool,
) ([]int, string, error) {
	displayString := line
	line = tr.replacer.Replace(line)
	line = strings.TrimSpace(reMultiSpace.ReplaceAllString(line, " "))
	words := strings.Split(line, " ")
	tokens := make([]int, 0)
	counts := make([]int, 0)
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

		// ignore numbers??
		//if !keyOK && utils.IsInt(word) && len(word) > cMaxNumDigits {
		//	excludesMap[word] = true
		//	tokens = append(tokens, cAsteriskItemID)
		//	continue
		//}
		termId = tr.te.register(word)
		if !uniqTokens[termId] {
			if onlyCurrTerms {
				tr.te.addCurrCount(termId, addCnt)
			} else {
				tr.te.addCount(termId, addCnt, false)
			}
		}
		uniqTokens[termId] = true

		if useTermBorder {
			cnt := tr.te.counts[termId]
			if tr.termCountBorder > cnt {
				//termId = cAsteriskItemID
				excludesMap[word] = true
			}
			counts = append(counts, cnt)
		}
		tokens = append(tokens, termId)
		if keyOK {
			tr.keyTermIds[termId] = ""
		}
	}

	if useTermBorder {
		minMatchLen := int(float64(len(words)) * tr.minMatchRate)
		if len(tokens) != len(counts) {
			return nil, "", fmt.Errorf("length of tokens and counts does not match: tokens:%d counts:%d", len(tokens), len(counts))
		}
		// sort in descending order
		sort.Slice(counts, func(i, j int) bool {
			return counts[i] > counts[j]
		})

		// replace words with "*" if they are not frequent words
		// but put priority on match rate to avoid having groups with many "*"s
		matchedCount := 0
		matchBorderCount := tr.termCountBorder
		for _, cnt := range counts {
			if cnt == 0 {
				break
			}
			matchedCount++
			if matchedCount >= minMatchLen {
				matchBorderCount = cnt
				break
			}
		}
		border := tr.termCountBorder
		if matchBorderCount < border {
			border = matchBorderCount
		}
		for i, termId := range tokens {
			if termId == cAsteriskItemID {
				continue
			}
			cnt := tr.te.counts[termId]
			if cnt < border {
				tokens[i] = cAsteriskItemID
				excludesMap[tr.te.id2term[termId]] = true
			}
		}

	}

	if needDisplayString {
		for target := range excludesMap {
			displayString = utils.Replace(displayString, target, "*", tr.separators)
		}
		// Combine multiple consecutive "*" into a single "*"
	} else {
		displayString = ""
	}

	return tokens, displayString, nil
}

func (tr *trans) lineToTerms(line string, addCnt int) {
	if !tr._match(line) {
		return
	}
	line, _, retentionPos := tr.parseLine(line, 0)

	tr.toTokens(line, addCnt, false, false, false)
	if tr.currRetentionPos > 0 && retentionPos > tr.currRetentionPos {
		if tr.countByBlock > tr.maxCountByBlock {
			tr.maxCountByBlock = tr.countByBlock
		}
		tr.countByBlock = 0
	}
	tr.countByBlock++
	tr.totalLines++

	tr.currRetentionPos = retentionPos
}

// analyze the line and
func (tr *trans) lineToLogGroup(line string, addCnt int, updated int64) (int64, error) {
	if !tr._match(line) {
		return -1, nil
	}
	line, updated, retentionPos := tr.parseLine(line, updated)
	if (tr.currRetentionPos > 0 && retentionPos > tr.currRetentionPos) || tr.countByBlock > tr.maxCountByBlock {
		if err := tr.next(updated); err != nil {
			return -1, err
		}
	}

	tokens, displayString, err := tr.toTokens(line, addCnt, true, true, true)
	if err != nil {
		return -1, err
	}
	groupId := tr.lgs.registerLogTree(tokens, addCnt, displayString, updated, updated, true, retentionPos, -1)

	tr.currRetentionPos = retentionPos
	return groupId, nil
}

func (tr *trans) commit(completed bool) error {
	if tr.readOnly {
		return nil
	}
	if err := tr.te.commit(completed); err != nil {
		return err
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
	if tr.te.CircuitDB != nil {
		tr.te = nil
	}
}

func (tr *trans) rebuildLogTree(termCountBorder int) *logTree {
	lt := tr.lgs.lt
	newTree := newLogTree(lt.depth)
	lt.rebuildHelper(newTree, tr.te, termCountBorder)
	return newTree
}

func (tr *trans) load() error {
	lgs := tr.lgs
	if lgs.DataDir == "" {
		return nil
	}

	if err := tr.te.load(); err != nil {
		return err
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

	rows, err := lgs.SelectCompletedRows(nil, nil, tableDefs["logGroups"])
	if err != nil {
		return err
	}
	if rows == nil {
		return nil
	}

	ds := lgs.displayStrings
	for rows.Next() {
		var groupIdstr string
		var retentionPos int64
		var count int
		var created int64
		var updated int64
		err = rows.Scan(&groupIdstr, &retentionPos, &count, &created, &updated)
		if err != nil {
			return err
		}
		//groupId, err := utils.Base36ToInt64(groupIdstr)
		groupId, err := strconv.ParseInt(groupIdstr, 10, 64)
		if err != nil {
			return fmt.Errorf("error parsing %s to int64", groupIdstr)
		}

		tokens, displayString, err := tr.toTokens(ds[groupId], 0, true, true, false)
		if err != nil {
			return err
		}
		ds[groupId] = displayString

		// for debugging
		//if displayString != line {
		//	_, _, _ = tr.toTokens(line, 0, true, true, false)
		//	return utils.ErrorStack("loaded displayString does not match parsed displayString\nparsed:\n%s \n\nloaded:\n%s\n\n",
		//		displayString, line)
		//}
		tr.lgs.registerLogTree(tokens, count, displayString, updated, updated, false,
			retentionPos, groupId)

		if retentionPos > tr.currRetentionPos {
			tr.currRetentionPos = retentionPos
		}
	}

	trows, err := tr.lgs.SelectFromCurrentTable(nil, tableDefs["logGroups"])
	if err != nil {
		return err
	}
	if trows == nil {
		return nil
	}
	for trows.Next() {
		var groupIdstr string
		var retentionPos int64
		var count int
		var created int64
		var updated int64
		err = trows.Scan(&groupIdstr, &retentionPos, &count, &created, &updated)
		if err != nil {
			return err
		}
		//groupId, err := utils.Base36ToInt64(groupIdstr)
		groupId, err := strconv.ParseInt(groupIdstr, 10, 64)
		if err != nil {
			return fmt.Errorf("error parsing %s to int64", groupIdstr)
		}

		line := ds[groupId]
		tokens, displayString, err := tr.toTokens(line, 0, true, true, false)
		if err != nil {
			return err
		}
		tr.lgs.registerLogTree(tokens, count, displayString, updated, updated, true, retentionPos, groupId)

		if retentionPos > tr.currRetentionPos {
			tr.currRetentionPos = retentionPos
		}
	}
	return nil
}

func (tr *trans) next(updated int64) error {
	lgs := tr.lgs
	if tr.readOnly {
		return nil
	}

	// write the current block
	if err := tr.te.next(updated); err != nil {
		return err
	}
	// write the current block
	if err := lgs.next(updated); err != nil {
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
		//groupId, err := utils.Base36ToInt64(groupIdstr)
		groupId, err := strconv.ParseInt(groupIdstr, 10, 64)
		if err != nil {
			return fmt.Errorf("error parsing %s to int64", groupIdstr)
		}
		// if the groupId exist in the new Block
		if lg, ok := lgs.alllg[groupId]; ok {
			lg.count -= count
		} else {
			line := ds[groupId]
			// will reach an existing logGrouup using termCountBorder
			tokens, _, err := tr.toTokens(line, count, true, false, false)
			if err != nil {
				return err
			}
			_groupId := tr.lgs.lt.search(tokens)
			if groupId == _groupId {
				lg.count -= count
			}
		}
	}

	tr.countByBlock = 0

	return nil
}

/*
returns table of LogGroup history, list of groupIds and corresponding displayStrings

		x axis is epochs of the block
		y axis is logGroups
	         epoch0 epoch1 epoch2
		grp0     10     11      5
		grp1     20     21     25
		grp2     30     31      5
*/
func (tr *trans) getLogGroupsHistory(groupIds []int64) (*logGroupsHistory, error) {
	lgs := tr.lgs
	if err := tr.loadLogGroupHistory(); err != nil {
		return nil, err
	}

	lgsh := newLogGroupsHistory(lgs, tr.lgs.minRetentionPos,
		tr.lgs.maxRetentionPos, tr.unitSecs, groupIds)
	return lgsh, nil
}

// useful for testing
func (tr *trans) searchLogGroup(displayString string) *logGroup {
	tokens, _, _ := tr.toTokens(displayString, 0, true, false, false)
	groupId := tr.lgs.lt.search(tokens)
	return tr.lgs.alllg[groupId]
}

func (tr *trans) setCountBorder() {
	if tr.termCountBorder == 0 {
		tr.termCountBorder = tr.te.getCountBorder(tr.termCountBorderRate)
	}
}

// get list of biggest N logGroups
func (tr *trans) getBiggestGroupIds(N int) []int64 {
	lgs := tr.lgs

	// Create a slice of key-value pairs
	groupIds := make([]int64, 0, len(lgs.alllg))
	for groupId := range lgs.alllg {
		groupIds = append(groupIds, groupId)
	}

	// Sort the slice by Count in descending order
	sort.Slice(groupIds, func(i, j int) bool {
		cnti := lgs.alllg[groupIds[i]].count
		cntj := lgs.alllg[groupIds[j]].count
		if cnti == cntj {
			return groupIds[i] < groupIds[j]
		}
		return cnti > cntj
	})

	// Extract the top N itemIDs
	if N > 0 && len(groupIds) > N {
		return groupIds[:N]
	}

	return groupIds
}

// load countHistory.
// call this function only when needed as it eats memory
func (tr *trans) loadLogGroupHistory() error {
	lgs := tr.lgs
	if lgs.DataDir == "" {
		return nil
	}

	rows, err := lgs.SelectRows(nil, nil, tableDefs["logGroups"])
	if err != nil {
		return err
	}
	if rows == nil {
		return nil
	}

	ds := lgs.displayStrings

	// for when after analyzer.rebuildTrans() has called
	orgDs := lgs.orgDisplayStrings

	for rows.Next() {
		var groupIdstr string
		var retentionPos int64
		var count int
		var created int64
		var updated int64
		err = rows.Scan(&groupIdstr, &retentionPos, &count, &created, &updated)
		if err != nil {
			return err
		}
		//groupId, err := utils.Base36ToInt64(groupIdstr)
		groupId, err := strconv.ParseInt(groupIdstr, 10, 64)
		if err != nil {
			return fmt.Errorf("error parsing %s to int64", groupIdstr)
		}

		if retentionPos > lgs.maxRetentionPos {
			lgs.maxRetentionPos = retentionPos
		}
		if lgs.minRetentionPos == 0 || retentionPos < lgs.minRetentionPos {
			lgs.minRetentionPos = retentionPos
		}

		displayString := ""
		if orgDs == nil {
			// rebuildTrans() is not called so no need to consider union logGroups
			displayString = ds[groupId]
		} else {
			if line, ok := orgDs[groupId]; ok {
				groupId, err = tr.lineToLogGroup(line, 0, 0)
				if err != nil {
					return err
				}
			} else {
				return utils.ErrorStack("displayString below did not match any logGrouop\n%s\n\n", line)
			}
			displayString = lgs.displayStrings[groupId]
		}

		lg, ok := lgs.alllg[groupId]
		if !ok {
			lg = new(logGroup)
			lg.countHistory = make(map[int64]int)
		} else if lg.countHistory == nil {
			lg.countHistory = make(map[int64]int)
		}
		lg.countHistory[retentionPos] += count
		lg.displayString = displayString
		lg.count += count
		if created > 0 && lg.created > created {
			lg.created = created
		}
		if updated > 0 && lg.updated < updated {
			lg.updated = updated
		}
	}

	return nil
}
