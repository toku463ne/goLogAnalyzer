package logan

import (
	"bufio"
	"fmt"
	"goLogAnalyzer/pkg/csvdb"
	"goLogAnalyzer/pkg/utils"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type patternkey struct {
	epoch   int64
	matched bool  // true if it is the regex matched line
	groupId int64 // logGroupId
}

type pattern struct {
	startEpoch int64 // start epoch of the pattern
	count      int
}

type patternkeys struct {
	*csvdb.CircuitDB
	ac                   *utils.AC
	regexRes             []*regexp.Regexp
	regexPatternKeyPoses map[*regexp.Regexp]int
	regexTagPoses        map[*regexp.Regexp](map[string]int)
	regexes              []string
	records              map[string][]patternkey
	testMode             bool
	idFilePath           string       // path to the keygroup IDs file
	searchKeyIds         []string     // keys to search for in the patternkeys
	pt                   *patternTags // pt for patternKeyId -> relationKey
}

// Newpatternkeys creates a new patternkeys instance
func newPatternKeys(dataDir string, regexes []string, useGzip bool, testMode bool) (*patternkeys, error) {

	pk := &patternkeys{
		ac:                   utils.NewAC(),
		regexRes:             make([]*regexp.Regexp, 0),
		regexPatternKeyPoses: make(map[*regexp.Regexp]int),
		regexTagPoses:        make(map[*regexp.Regexp](map[string]int)),
		regexes:              regexes,
		records:              make(map[string][]patternkey, 10000),
		idFilePath:           dataDir + "/patternKeyIds.txt",
	}
	pk.regexRes = make([]*regexp.Regexp, 0)
	pk.regexPatternKeyPoses = make(map[*regexp.Regexp]int)
	pk.regexTagPoses = make(map[*regexp.Regexp]map[string]int)
	pk.testMode = testMode

	// Compile regexes and store them in the patternkeys
	for _, classRegex := range pk.regexes {
		re := regexp.MustCompile(`` + classRegex)
		pk.regexRes = append(pk.regexRes, re)
		names := re.SubexpNames()
		for i, name := range names {
			if name == cPatternKey {
				pk.regexPatternKeyPoses[re] = i
			} else if name != "" {
				if _, ok := pk.regexTagPoses[re]; !ok {
					pk.regexTagPoses[re] = make(map[string]int)
				}
				pk.regexTagPoses[re][name] = i
			}
		}
	}

	if pk.testMode {
		return pk, nil
	}

	kgdb, err := csvdb.NewCircuitDB(dataDir, "patternkeys",
		tableDefs["patternKeys"], 0, 0, 0, 0, useGzip)
	if err != nil {
		return nil, err
	}
	pk.CircuitDB = kgdb

	pt, err := newPatternTags(dataDir, useGzip, testMode)
	if err != nil {
		return nil, fmt.Errorf("error creating pattern pt: %v", err)
	}
	pk.pt = pt

	return pk, nil
}

func (pk *patternkeys) register(patternKeyId string) {
	// Register the patternKeyId in the Aho-Corasick automaton
	pk.ac.Register(patternKeyId)
	if _, exists := pk.records[patternKeyId]; !exists {
		pk.records[patternKeyId] = make([]patternkey, 0, 10)
	}
}

// return: patternKeyId, tags, matched, error
func (pk *patternkeys) findAndRegister(line string) (string, map[string]string, bool, error) {
	patternKeyId := ""
	tags := make(map[string]string)
	matched := false
	// Find the first matching regex and register the patternKeyId
	for _, re := range pk.regexRes {
		ma := re.FindStringSubmatch(line)
		if len(ma) > 0 {
			if pk.regexPatternKeyPoses[re] >= 0 && pk.regexPatternKeyPoses[re] < len(ma) {
				patternKeyId = ma[pk.regexPatternKeyPoses[re]]
				if patternKeyId != "" {
					matched = true
					pk.register(patternKeyId)
				}
				for tagName, pos := range pk.regexTagPoses[re] {
					if pos > 0 && pos < len(ma) {
						pk.pt.set(patternKeyId, tagName, ma[pos])
						tags[tagName] = ma[pos]
					}
				}
			}
		}
	}
	return patternKeyId, tags, matched, nil
}

func (pk *patternkeys) hasMatch(term []byte) bool {
	// Check if the term matches any of the registered patternKeyIds
	return len(pk.ac.MatchExact(term)) > 0
}

func (pk *patternkeys) appendLogGroup(patternKeyId string, epoch int64, matched bool, logGroupId int64) {
	// append a log group ID to the list for the given patternKeyId
	// Do not check the existance of patternKeyId in the records map. It indicates a bug if it is not registered
	pk.records[patternKeyId] = append(pk.records[patternKeyId], patternkey{epoch: epoch, matched: matched, groupId: logGroupId})
}

func (pk *patternkeys) flush() error {
	if pk.DataDir == "" || pk.testMode {
		return nil
	}

	// insert the patternkeys into the CSV DB
	for patternKeyId, records := range pk.records {
		for _, rec := range records {
			// inserts to buffer only
			if err := pk.InsertRow(tableDefs["patternKeys"],
				patternKeyId, rec.epoch, rec.matched, rec.groupId); err != nil {
				return err
			}
		}
	}

	// flush buffer to the block table file with WRITE mode (not APPEND)
	if err := pk.FlushOverwriteCurrentTable(); err != nil {
		return err
	}

	// write the keygroup IDs to the file
	if err := utils.WriteStringToFile(pk.idFilePath, pk.ac.GetPatterns()); err != nil {
		return err
	}

	return nil
}

func (pk *patternkeys) commit(completed bool) error {
	if pk.DataDir == "" {
		return nil
	}
	if err := pk.flush(); err != nil {
		return err
	}

	if err := pk.UpdateBlockStatus(completed); err != nil {
		return err
	}

	if err := pk.pt.commit(completed); err != nil {
		return fmt.Errorf("error committing pattern pt: %v", err)
	}
	return nil
}

func (pk *patternkeys) next(updated int64) error {
	// save logGroup IDs to CSV DB
	// write current block to the block table
	if err := pk.flush(); err != nil {
		return err
	}
	if err := pk.NextBlock(updated); err != nil {
		return err
	}
	if err := pk.pt.next(updated); err != nil {
		return err
	}

	// clear the logGroupIds map for the next block
	pk.records = make(map[string][]patternkey, 10000)

	return nil
}

func (pk *patternkeys) loadKeyIdsFromFile() error {
	// reset the Aho-Corasick automaton
	pk.ac = utils.NewAC()

	// read the keygroup IDs from the file
	file, err := os.Open(pk.idFilePath)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		pk.register(scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return err
	}
	return nil
}

func (pk *patternkeys) load() error {
	// load keygroup IDs from the file
	if err := pk.loadKeyIdsFromFile(); err != nil {
		return err
	}

	cnt := pk.CountFromStatusTable(nil)
	if cnt <= 0 {
		return nil
	}

	if err := pk.LoadCircuitDBStatus(); err != nil {
		return err
	}

	// load from the last block table
	trows, err := pk.SelectFromCurrentTable(nil, tableDefs["patternKeys"])
	if err != nil {
		return err
	}
	if trows == nil {
		return nil
	}
	for trows.Next() {
		var patternKeyId string
		var epoch int64
		var matched bool
		var logGroupId int64
		if err := trows.Scan(&patternKeyId, &epoch, &matched, &logGroupId); err != nil {
			return err
		}
		pk.appendLogGroup(patternKeyId, epoch, matched, logGroupId)
	}
	return nil
}

// values: ["patternKeyId"]
func (pk *patternkeys) filterKeys(values []string) bool {
	for _, keyId := range pk.searchKeyIds {
		if keyId == values[0] {
			return true
		}
	}
	return false
}

func (pk *patternkeys) loadAll(searchKeyIds []string) error {
	// init pk
	pk.records = make(map[string][]patternkey, 10000)

	// load keyIds from the file
	if err := pk.loadKeyIdsFromFile(); err != nil {
		return err
	}

	pk.searchKeyIds = nil
	var f func(values []string) bool
	if searchKeyIds != nil {
		pk.searchKeyIds = searchKeyIds
		f = pk.filterKeys
	} else {
		f = nil
	}
	rows, err := pk.SelectRows(f, nil, tableDefs["patternKeys"])
	if err != nil {
		return err
	}
	if rows == nil {
		return nil
	}

	// register all keygroup IDs
	for rows.Next() {
		var patternKeyId string
		var epoch int64
		var matched bool
		var logGroupId int64
		if err := rows.Scan(&patternKeyId, &epoch, &matched, &logGroupId); err != nil {
			return err
		}
		pk.appendLogGroup(patternKeyId, epoch, matched, logGroupId)
	}

	if err := pk.pt.loadAll(); err != nil {
		return fmt.Errorf("error loading pattern pt: %v", err)
	}

	return nil
}

/*
Detect patterns from appearing loggroupIds per keyId
example1)
keyId: 1234
groupIds: 5 6 7 5 6 7 7 5 6 7 5
matched:  1 0 0 1 0 0 0 1 0 0 1

the the patterns below will be detected:
5 6 7 => 2 times
5 6 7 7 => 1 time
5 => 1 time
*/
func (pk *patternkeys) detectPatternsByFirstMatch() map[string](map[string]*pattern) {
	// commit first
	if err := pk.commit(true); err != nil {
		return nil
	}

	if err := pk.loadAll(nil); err != nil {
		return nil
	}

	patterns := make(map[string](map[string]*pattern)) // patternStr -> patternKeyId -> pattern

	patternStr := ""
	startEpoch := int64(0)
	patternKeyId := ""
	matched_process := func() {
		subpat, ok := patterns[patternStr]
		if !ok {
			subpat = make(map[string]*pattern)
			patterns[patternStr] = subpat
		}
		pat, ok := subpat[patternKeyId]
		if !ok {
			pat = &pattern{startEpoch: startEpoch, count: 0}
			subpat[patternKeyId] = pat
		}
		pat.count++
		patternStr = ""
	}

	for _patternKeyId, records := range pk.records {
		patternKeyId = _patternKeyId
		if len(records) == 0 {
			continue
		}
		startEpoch = records[0].epoch
		patternStr = ""
		for _, rec := range records {
			if rec.matched && patternStr != "" {
				matched_process()
				startEpoch = int64(rec.epoch)
			}
			if patternStr == "" {
				patternStr = fmt.Sprintf("%d", rec.groupId)
			} else {
				patternStr += fmt.Sprintf(" %d", rec.groupId)
			}
		}
		if patternStr != "" {
			// process the last pattern
			matched_process()
			startEpoch = 0
		}
	}

	return patterns
}

/*
show patterns by patternStr ordered by count descending
example1)
5 6 7 => total 2

	1234: {startEpoch: 1, count: 2}
	5678: {startEpoch: 25, count: 1}

5 6 7 7 => total 1

	1234: {startEpoch: 11, count: 1}

5 => total 1

	1234: {startEpoch: 30, count: 1}

9 => total 1

	5678: {startEpoch: 35, count: 1}
*/
func (pk *patternkeys) ShowPatternsByFirstMatch(minCount int,
	lgs *logGroups) (map[string](map[string]*pattern), error) {
	patterns := pk.detectPatternsByFirstMatch()
	if patterns == nil {
		return nil, nil
	}

	type patSum struct {
		patternStr string
		total      int
	}

	sums := make([]patSum, 0, len(patterns))
	// calculate totals and filter by minCount
	for pstr, sub := range patterns {
		total := 0
		for _, pat := range sub {
			total += pat.count
		}
		if total < minCount {
			delete(patterns, pstr)
			continue
		}
		sums = append(sums, patSum{patternStr: pstr, total: total})
	}

	// selection sort sums by total desc
	for i := 0; i < len(sums)-1; i++ {
		maxIdx := i
		for j := i + 1; j < len(sums); j++ {
			if sums[j].total > sums[maxIdx].total {
				maxIdx = j
			}
		}
		if maxIdx != i {
			sums[i], sums[maxIdx] = sums[maxIdx], sums[i]
		}
	}

	for _, ps := range sums {
		fmt.Printf("%s => total %d\n", ps.patternStr, ps.total)

		// collect sub patterns for ordering
		type subInfo struct {
			keyId      string
			startEpoch int64
			count      int
		}
		subInfos := make([]subInfo, 0, len(patterns[ps.patternStr]))
		for keyId, pat := range patterns[ps.patternStr] {
			subInfos = append(subInfos, subInfo{keyId: keyId, startEpoch: pat.startEpoch, count: pat.count})
		}
		// sort subInfos by count desc
		for i := 0; i < len(subInfos)-1; i++ {
			maxIdx := i
			for j := i + 1; j < len(subInfos); j++ {
				if subInfos[j].count > subInfos[maxIdx].count {
					maxIdx = j
				}
			}
			if maxIdx != i {
				subInfos[i], subInfos[maxIdx] = subInfos[maxIdx], subInfos[i]
			}
		}
		for _, si := range subInfos {
			ts := time.Unix(si.startEpoch, 0).Local().Format("2006/01/02 15:04:05")
			fmt.Printf("%s: {startEpoch: %s, count: %d}\n", si.keyId, ts, si.count)
		}
		fmt.Println("---")

		// print display strings for the pattern
		for _, groupIdStr := range strings.Split(ps.patternStr, " ") {
			groupId, err := strconv.ParseInt(groupIdStr, 10, 64)
			if err != nil {
				return nil, fmt.Errorf("error parsing groupId %s: %v", groupIdStr, err)
			}
			displayString, ok := lgs.displayStrings[groupId]
			if ok {
				fmt.Printf("%s\n", displayString)
			} else {
				fmt.Printf("(not found for groupId %d)\n", groupId)
			}
		}
		fmt.Println()
		fmt.Println("--------------------------------------------------")
	}

	return patterns, nil
}

/*
Count patternKeys in the patterns which are group of ordered logGroupIds
example)
input
patternKey: p1
logGroupIds: 5 6 7
relationKey: 1234

patternKey: p2
logGroupIds: 5 6 7
relationKey: 1234

patternKey: p3
logGroupIds: 5 6 7
relationKey: 5678

patternKey: p4
logGroupIds: 5 6 7 7
relationKey: 5678

then the patterns will be:
5 6 7 => 3 times

	1234: {startEpoch: 0, count: 2}
	5678: {startEpoch: 0, count: 1}

5 6 7 7 => 1 time

	5678: {startEpoch: 0, count: 1}
*/
func (pk *patternkeys) detectPatternsByPatternKeys() map[string](map[string]*pattern) {
	// commit first
	if err := pk.commit(true); err != nil {
		return nil
	}

	if err := pk.loadAll(nil); err != nil {
		return nil
	}

	patterns := make(map[string](map[string]*pattern)) // patternStr -> tagName1:tagValue1,tagName2:tagValue2,.. -> pattern
	// => case no relationKey, relationKey="all" and save the summary in pattern struct
	// Build one ordered pattern (space-separated groupIds) per patternKeyId,
	// then aggregate counts by relationKey (or "all" if none).
	for patternKeyId, records := range pk.records {
		if len(records) == 0 {
			continue
		}

		var b strings.Builder
		for i, rec := range records {
			if i > 0 {
				b.WriteByte(' ')
			}
			b.WriteString(strconv.FormatInt(rec.groupId, 10))
		}
		patternStr := b.String()
		startEpoch := records[0].epoch

		// Collect unique relation keys; fallback to "all" when none exist.
		relSet := make(map[string]struct{})
		if pk.pt != nil {
			tagStr := ""
			for tagName, tagValue := range pk.pt.get(patternKeyId) {
				if tagValue != "" {
					tagStr += fmt.Sprintf("%s:%s, ", tagName, tagValue)
				}
			}
			if len(tagStr) > 0 {
				tagStr = tagStr[:len(tagStr)-2] // remove trailing comma and space
			}
			relSet[tagStr] = struct{}{}
		}
		if len(relSet) == 0 {
			relSet["all"] = struct{}{}
		}

		sub, ok := patterns[patternStr]
		if !ok {
			sub = make(map[string]*pattern)
			patterns[patternStr] = sub
		}
		for rk := range relSet {
			if pat, ok := sub[rk]; ok {
				pat.count++
				if pat.startEpoch == 0 || startEpoch < pat.startEpoch {
					pat.startEpoch = startEpoch
				}
			} else {
				sub[rk] = &pattern{startEpoch: startEpoch, count: 1}
			}
		}
	}

	return patterns

}

/*
Show patterns by pt ordered by count descending
If there are no pt, it will show "all" relation.
example1)
5 6 7 => total 3

	1234: {startEpoch: 1, count: 2}
	5678: {startEpoch: 25, count: 1}

5 6 7 7 => total 1

	1234: {startEpoch: 11, count: 1}

7 8 9 => total 5

	all: {startEpoch: 15, count: 5}
*/
func (pk *patternkeys) ShowPatternsByPatternsKeys(minCount int, lgs *logGroups) (map[string](map[string]*pattern), error) {
	patterns := pk.detectPatternsByPatternKeys()
	if patterns == nil {
		return nil, nil
	}

	// summarize totals and filter by minCount
	type patSum struct {
		patternStr string
		total      int
	}
	sums := make([]patSum, 0, len(patterns))
	for pstr, sub := range patterns {
		total := 0
		for _, pat := range sub {
			total += pat.count
		}
		if total < minCount {
			delete(patterns, pstr)
			continue
		}
		sums = append(sums, patSum{patternStr: pstr, total: total})
	}

	// selection sort by total desc
	for i := 0; i < len(sums)-1; i++ {
		maxIdx := i
		for j := i + 1; j < len(sums); j++ {
			if sums[j].total > sums[maxIdx].total {
				maxIdx = j
			}
		}
		if maxIdx != i {
			sums[i], sums[maxIdx] = sums[maxIdx], sums[i]
		}
	}

	for _, ps := range sums {
		fmt.Printf("%s => total %d\n", ps.patternStr, ps.total)

		// collect pt for ordering
		type relInfo struct {
			relationKey string
			startEpoch  int64
			count       int
		}
		relInfos := make([]relInfo, 0, len(patterns[ps.patternStr]))
		for rk, pat := range patterns[ps.patternStr] {
			relInfos = append(relInfos, relInfo{relationKey: rk, startEpoch: pat.startEpoch, count: pat.count})
		}
		// sort by count desc
		for i := 0; i < len(relInfos)-1; i++ {
			maxIdx := i
			for j := i + 1; j < len(relInfos); j++ {
				if relInfos[j].count > relInfos[maxIdx].count {
					maxIdx = j
				}
			}
			if maxIdx != i {
				relInfos[i], relInfos[maxIdx] = relInfos[maxIdx], relInfos[i]
			}
		}
		for _, ri := range relInfos {
			ts := time.Unix(ri.startEpoch, 0).Local().Format("2006/01/02 15:04:05")
			fmt.Printf("%s: {startEpoch: %s, count: %d}\n", ri.relationKey, ts, ri.count)
		}
		fmt.Println("---")

		// print display strings for the pattern
		for _, groupIdStr := range strings.Split(ps.patternStr, " ") {
			groupId, err := strconv.ParseInt(groupIdStr, 10, 64)
			if err != nil {
				return nil, fmt.Errorf("error parsing groupId %s: %v", groupIdStr, err)
			}
			if displayString, ok := lgs.displayStrings[groupId]; ok {
				fmt.Printf("%s\n", displayString)
			} else {
				fmt.Printf("(not found for groupId %d)\n", groupId)
			}
		}
		fmt.Println()
		fmt.Println("--------------------------------------------------")
	}

	return patterns, nil
}
