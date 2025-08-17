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

	"github.com/sirupsen/logrus"
)

type trans struct {
	te                  *terms
	lt                  *logTree
	lgs                 *logGroups
	pk                  *patternkeys
	customLogGroups     []string
	replacer            *strings.Replacer
	logFormatRe         *regexp.Regexp
	msgFormatRes        []*regexp.Regexp
	msgPoses            map[*regexp.Regexp]int
	timestampLayout     string
	useUtcTime          bool
	timestampPos        int
	messagePos          int
	readOnly            bool
	totalLines          int
	filterRe            []*regexp.Regexp
	xFilterRe           []*regexp.Regexp
	termCountBorderRate float64
	termCountBorder     int
	minMatchRate        float64
	keywords            map[string]bool
	ignorewords         map[string]bool
	keyRes              []*regexp.Regexp
	ignoreRes           []*regexp.Regexp
	unitSecs            int64
	countByBlock        int
	maxCountByBlock     int
	currRetentionPos    int64
	separators          string
	timestampRe         *regexp.Regexp
	testMode            bool
	ignoreNumbers       bool
}

func newTrans(dataDir, logFormat, timestampLayout string,
	useUtcTime bool,
	maxBlocks, blockSize int,
	unitSecs int64, keepPeriod int64,
	termCountBorderRate float64,
	termCountBorder int,
	minMatchRate float64,
	searchRegex, exludeRegex,
	_keywords, _ignorewords,
	_keyRegexes, _ignoreRegexes,
	_msgFormats []string,
	_kgRegexes []string,
	_customLogGroups []string,
	separators string,
	useGzip, readOnly, testMode, ignoreNumbers bool) (*trans, error) {
	tr := new(trans)
	tr.replacer = getDelimReplacer(separators)
	tr.timestampLayout = timestampLayout
	tr._parseLogFormat(logFormat)
	tr.termCountBorder = termCountBorder
	tr.termCountBorderRate = termCountBorderRate
	tr.minMatchRate = minMatchRate
	tr.useUtcTime = useUtcTime
	tr.separators = separators
	tr.readOnly = readOnly
	tr.testMode = testMode
	tr.ignoreNumbers = ignoreNumbers
	tr._setFilters(searchRegex, exludeRegex)
	tr._setKeyRegexes(_keyRegexes, _ignoreRegexes)

	tr.keywords = make(map[string]bool)
	tr.ignorewords = make(map[string]bool)
	for _, word := range _keywords {
		tr.keywords[word] = true
	}
	for _, word := range _ignorewords {
		tr.ignorewords[word] = true
	}
	tr.unitSecs = unitSecs
	tr.maxCountByBlock = blockSize

	// don't need blockSize for terms because the rotation follows trans.next()
	te, err := newTerms(dataDir, maxBlocks, unitSecs, keepPeriod, useGzip, tr.testMode)
	if err != nil {
		return nil, err
	}
	tr.te = te

	if len(_msgFormats) > 0 {
		tr._parseMsgFormat(_msgFormats)
	}

	if len(_kgRegexes) > 0 {
		tr.pk, err = newPatternKeys(dataDir, _kgRegexes, useGzip, tr.testMode)
		if err != nil {
			return nil, err
		}
	}

	tr.lt = newLogTree(0)
	// don't need blockSize for terms because the rotation follows trans.next()
	lgs, err := newLogGroups(dataDir, maxBlocks, unitSecs, keepPeriod, useGzip, tr.testMode)
	if err != nil {
		return nil, err
	}
	tr.lgs = lgs

	tr.initCounters()
	return tr, nil
}

func (tr *trans) initCounters() {
	tr.currRetentionPos = 0
	tr.totalLines = 0
	tr.countByBlock = 0
}

func (tr *trans) _parseLogFormat(logFormat string) {
	suffixPattern := regexp.MustCompile(`\b(\d{1,2})(st|nd|rd|th)\b`)

	// If suffixes are found, replace them
	if suffixPattern.MatchString(tr.timestampLayout) {
		tr.timestampLayout = suffixPattern.ReplaceAllString(tr.timestampLayout, `$1`)
		tr.timestampRe = regexp.MustCompile(`(\d{1,2})(st|nd|rd|th)`)
		needDateFormatCleaning = true
	}

	//re := regexp.MustCompile(logFormat)
	re := regexp.MustCompile(`` + logFormat) // Use (?i) for case-insensitive matching if needed

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

func (tr *trans) _parseMsgFormat(msgFormats []string) {
	tr.msgFormatRes = make([]*regexp.Regexp, 0)
	tr.msgPoses = make(map[*regexp.Regexp]int)
	for _, msgFormat := range msgFormats {
		re := regexp.MustCompile(`` + msgFormat)
		tr.msgFormatRes = append(tr.msgFormatRes, re)
		names := re.SubexpNames()
		for i, name := range names {
			if name == "message" {
				tr.msgPoses[re] = i
			}
		}
	}
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

func (tr *trans) _setKeyRegexes(keyRegexes, ignoreRegexes []string) {
	tr.keyRes = make([]*regexp.Regexp, 0)
	for _, s := range keyRegexes {
		tr.keyRes = append(tr.keyRes, utils.GetRegex(s))
	}

	tr.ignoreRes = make([]*regexp.Regexp, 0)
	for _, s := range ignoreRegexes {
		tr.ignoreRes = append(tr.ignoreRes, utils.GetRegex(s))
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

func (tr *trans) _matchKey(res []*regexp.Regexp, text string) bool {
	b := []byte(text)
	matched := false
	for _, re := range res {
		if re.Match(b) {
			matched = true
		}
	}
	return matched
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

func (tr *trans) parseLine(line string, updated int64) (string, int64, int64, error) {
	var lastdt time.Time
	var err error
	lastUpdate := int64(0)
	retentionPos := int64(-1)
	line = strings.TrimSpace(reMultiSpace.ReplaceAllString(line, " "))
	if line == "" {
		return "", 0, 0, nil
	}

	if tr.timestampPos >= 0 || tr.messagePos >= 0 {
		// Optimized: avoid allocating every submatch string when we only need timestamp/message.
		// Original: ma := tr.logFormatRe.FindStringSubmatch(line)
		var ma []string
		if tr.logFormatRe != nil {
			idxs := tr.logFormatRe.FindStringSubmatchIndex(line)
			if len(idxs) > 0 {
				totalGroups := len(idxs) / 2 // includes full match
				needPos := tr.timestampPos
				if tr.messagePos > needPos {
					needPos = tr.messagePos
				}
				if needPos < 0 {
					needPos = 0
				}
				if needPos+1 > totalGroups {
					ma = nil // treat as no match
				} else {
					ma = make([]string, needPos+1) // only up to highest needed group
					if tr.timestampPos >= 0 {
						s, e := idxs[2*tr.timestampPos], idxs[2*tr.timestampPos+1]
						if s >= 0 && e >= 0 {
							ma[tr.timestampPos] = line[s:e]
						}
					}
					if tr.messagePos >= 0 {
						s, e := idxs[2*tr.messagePos], idxs[2*tr.messagePos+1]
						if s >= 0 && e >= 0 {
							if tr.messagePos >= len(ma) {
								tmp := make([]string, tr.messagePos+1)
								copy(tmp, ma)
								ma = tmp
							}
							ma[tr.messagePos] = line[s:e]
						}
					}
				}
			}
		}
		if tr.timestampPos >= 0 && len(ma) == 0 {
			//return "", 0, 0, fmt.Errorf("line does not match format:\n%s", line)
			return "", 0, 0, nil // treat as no match
		}
		if len(ma) > 0 {
			if tr.timestampPos >= 0 && tr.timestampLayout != "" && len(ma) > tr.timestampPos {
				if tr.useUtcTime {
					lastdt, err = utils.Str2Timestamp(tr.timestampLayout, ma[tr.timestampPos])
				} else {
					// system time zone
					dtstr := ma[tr.timestampPos]
					if needDateFormatCleaning && tr.timestampRe != nil {
						dtstr = tr.timestampRe.ReplaceAllString(dtstr, "$1")
					}

					lastdt, err = utils.Str2date(tr.timestampLayout, dtstr)
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
	return line, lastUpdate, retentionPos, nil
}

func (tr *trans) parseMessage(line string) string {
	for _, re := range tr.msgFormatRes {
		ma := re.FindStringSubmatch(line)
		if len(ma) > 0 {
			line = ma[tr.msgPoses[re]]
		}
	}
	return line
}

// convert line to list of tokens and register to tr.te.
// returns tokens, displayString and error
func (tr *trans) toTokens(line string, addCnt int,
	useTermBorder, needDisplayString, onlyCurrTerms, doPatternKeyMatching bool,
) ([]int, string, string, error) {
	displayString := line
	line = tr.replacer.Replace(line)
	//line = strings.TrimSpace(reMultiSpace.ReplaceAllString(line, " "))
	words := strings.Split(line, " ")
	tokens := make([]int, 0)
	uniqTokens := make(map[int]bool, 0)
	excludesMap := make(map[string]bool)
	excludedNumbers := make(map[string]bool)

	termId := -1

	patternKey := ""
	for _, w := range words {
		if w == "" {
			continue
		}

		word := strings.ToLower(w)
		if doPatternKeyMatching && tr.pk != nil {
			if ok := tr.pk.hasMatch([]byte(word)); ok {
				patternKey = word
			}
		}

		if tr.ignorewords[word] || tr._matchKey(tr.ignoreRes, word) {
			excludesMap[word] = true
			word = "*"
		}

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
		if !tr.keywords[word] && tr.ignoreNumbers && utils.IsRealNumber(word) && !tr._matchKey(tr.keyRes, word) {
			termId = cAsteriskItemID
			excludedNumbers[word] = true
		} else if word == "*" {
			termId = cAsteriskItemID
		} else {
			termId = tr.te.register(word)
		}
		if termId != cAsteriskItemID && !uniqTokens[termId] {
			if onlyCurrTerms {
				tr.te.addCurrCount(termId, addCnt)
			} else {
				tr.te.addCount(termId, addCnt, false)
			}
		}
		uniqTokens[termId] = true

		//if useTermBorder {
		//	cnt := tr.te.counts[termId]
		//	if !keyOK && tr.termCountBorder > cnt {
		//		//termId = cAsteriskItemID
		//		excludesMap[word] = true
		//		termId = cAsteriskItemID
		//	}
		//	counts = append(counts, cnt)
		//}
		tokens = append(tokens, termId)
	}

	if useTermBorder {
		//if len(tokens) != len(counts) {
		//	return nil, "", fmt.Errorf("length of tokens and counts does not match: tokens:%d counts:%d", len(tokens), len(counts))
		//}
		border := tr.termCountBorder

		if (tr.minMatchRate > 0.0 && tr.minMatchRate < 1.0) || tr.termCountBorderRate > 0.0 {
			minMatchLen := int(float64(len(tokens)) * tr.minMatchRate)
			counts := make([]int, 0)
			for _, termId := range tokens {
				cnt := tr.te.counts[termId]
				counts = append(counts, cnt)
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

			if matchBorderCount < border {
				border = matchBorderCount
			}
		}

		for i, termId := range tokens {
			if termId == cAsteriskItemID {
				continue
			}
			cnt := tr.te.counts[termId]
			w := tr.te.id2term[termId]
			if !tr.keywords[w] && !tr._matchKey(tr.keyRes, w) && cnt < border {
				tokens[i] = cAsteriskItemID
				excludesMap[w] = true
			}
		}
	}

	if needDisplayString {
		for target := range excludesMap {
			displayString = utils.Replace(displayString, target, "*", tr.separators)
		}
		for target := range excludedNumbers {
			displayString = utils.Replace(displayString, target, "*", tr.separators)
		}
		// Combine multiple consecutive "*" into a single "*"
	} else {
		displayString = ""
	}

	return tokens, displayString, patternKey, nil
}

func (tr *trans) lineToTerms(line string, addCnt int) error {
	if !tr._match(line) {
		return nil
	}
	if line == "" {
		return nil
	}
	line, _, retentionPos, err := tr.parseLine(line, 0)
	if err != nil {
		return err
	}

	tr.toTokens(line, addCnt, false, false, false, false)
	if tr.currRetentionPos > 0 && retentionPos > tr.currRetentionPos {
		if tr.countByBlock > tr.maxCountByBlock {
			tr.maxCountByBlock = tr.countByBlock
		}
		tr.countByBlock = 0
	}
	tr.countByBlock++
	tr.totalLines++

	tr.currRetentionPos = retentionPos
	return nil
}

// analyze the line and
func (tr *trans) lineToLogGroup(orgLine string, addCnt int, updated int64) (int64, error) {
	//if orgLine == "09th, 20:35:34.107+0900 TBLV3 CALL:  [0x00610F9A: 0x8823B0B2-0x00000000] CTBCAFBridge:     m=audio 27868 RTP/AVP 0 101 " {
	//	print("")
	//}

	if !tr._match(orgLine) {
		return -1, nil
	}
	if orgLine == "" {
		return -1, nil
	}
	line, updated, retentionPos, err := tr.parseLine(orgLine, updated)
	if err != nil {
		return -1, err
	}
	// pick up classid from the line
	matched := false
	if tr.pk != nil {
		_, _, matched, err = tr.pk.findAndRegister(line)
		if err != nil {
			return -1, err
		}
	}

	//if matched {
	//	println("matched pattern key:", line)
	//}

	line = tr.parseMessage(line)
	if (tr.currRetentionPos > 0 && retentionPos > tr.currRetentionPos) || tr.countByBlock > tr.maxCountByBlock {
		if err := tr.next(updated); err != nil {
			return -1, err
		}
	}

	tokens, displayString, patternKey, err := tr.toTokens(line, addCnt, true, true, true, true)
	if err != nil {
		return -1, err
	}

	groupId := tr.lgs.registerLogTree(tokens, addCnt, displayString, updated, updated, true, retentionPos, -1)
	cnt := len(tr.lgs.alllg)
	if cnt > cMaxLogGroups {
		logrus.Error(displayString)
		return -1, fmt.Errorf("logTree size went over %d", cMaxLogGroups)
	}

	// register the logGroupId to the patternkeys
	if tr.pk != nil {
		if patternKey != "" && groupId >= 0 {
			tr.pk.appendLogGroup(patternKey, updated, matched, groupId)
		}
	}

	lg := tr.lgs.alllg[groupId]
	lg.calcScore(tokens, tr.te)

	tr.lgs.lastMessages[groupId] = orgLine

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
	if tr.pk != nil {
		if err := tr.pk.commit(completed); err != nil {
			return err
		}
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

func (tr *trans) load() error {
	lgs := tr.lgs
	if lgs.CircuitDB == nil || lgs.DataDir == "" {
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

	if err := lgs.readLastMessages(); err != nil {
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

		tokens, displayString, _, err := tr.toTokens(ds[groupId], 0, true, true, false, false)
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
		tokens, displayString, _, err := tr.toTokens(line, 0, true, true, false, false)
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
	// write the current patternkeys
	if tr.pk != nil {
		if err := tr.pk.next(updated); err != nil {
			return err
		}
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
			tokens, _, _, err := tr.toTokens(line, count, true, false, false, false)
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
	tokens, _, _, _ := tr.toTokens(displayString, 0, true, false, false, false)
	groupId := tr.lgs.lt.search(tokens)
	return tr.lgs.alllg[groupId]
}

func (tr *trans) setCountBorder() {
	if tr.termCountBorder == 0 {
		tr.termCountBorder = tr.te.getCountBorder(tr.termCountBorderRate)
	}
}

// get top N logGroups
func (tr *trans) getTopNGroupIds(N int, minLastUpdate int64,
	searchString, excludeString string,
	minCnt, maxCnt int, asc bool) []int64 {
	lgs := tr.lgs

	searchStrings := make([]string, 0)
	exludeStrings := make([]string, 0)
	if searchString != "" {
		searchStrings = append(searchStrings, searchString)
	}
	if excludeString != "" {
		exludeStrings = append(exludeStrings, excludeString)
	}
	tr._setFilters(searchStrings, exludeStrings)

	// Create a slice of key-value pairs
	groupIds := make([]int64, 0, len(lgs.alllg))
	for groupId, lg := range lgs.alllg {
		if lg.updated >= minLastUpdate && lg.count >= minCnt && (maxCnt == 0 || lg.count <= maxCnt) {
			if !tr._match(lg.displayString) {
				continue
			}
			groupIds = append(groupIds, groupId)
		}
	}

	// Sort the slice by Count in ascending order if asc is true else descending order
	sort.Slice(groupIds, func(i, j int) bool {
		cnti := lgs.alllg[groupIds[i]].count
		cntj := lgs.alllg[groupIds[j]].count
		if cnti == cntj {
			scorei := lgs.alllg[groupIds[i]].rareScore
			scorej := lgs.alllg[groupIds[j]].rareScore
			if scorei == scorej {
				return groupIds[i] < groupIds[j]
			} else {
				if asc {
					return scorei > scorej
				} else {
					return scorei < scorej
				}
			}
		}
		if asc {
			return cnti < cntj
		} else {
			return cnti > cntj
		}
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
		if lgs.minRetentionPos == 0 || (retentionPos < lgs.minRetentionPos && retentionPos > 0) {
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

func (tr *trans) detectPaterns(minCnt int, mode string) error {
	if tr.pk == nil {
		return nil
	}

	switch mode {
	case "firstMatch":
		tr.pk.ShowPatternsByFirstMatch(minCnt, tr.lgs)
	case "relations":
		tr.pk.ShowPatternsByPatternsKeys(minCnt, tr.lgs)
	default:
		return fmt.Errorf("unknown mode %s for detectPaterns", mode)
	}

	return nil
}
