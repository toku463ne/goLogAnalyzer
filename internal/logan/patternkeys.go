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
	ac           *utils.AC
	regexRes     []*regexp.Regexp
	regexPoses   map[*regexp.Regexp]int
	regexes      []string
	records      map[string][]patternkey
	testMode     bool
	idFilePath   string   // path to the keygroup IDs file
	searchKeyIds []string // keys to search for in the patternkeys
}

// Newpatternkeys creates a new patternkeys instance
func newpatternkeys(dataDir string, regexes []string, useGzip bool, testMode bool) (*patternkeys, error) {

	pk := &patternkeys{
		ac:         utils.NewAC(),
		regexRes:   make([]*regexp.Regexp, 0),
		regexPoses: make(map[*regexp.Regexp]int),
		regexes:    regexes,
		records:    make(map[string][]patternkey, 10000),
		idFilePath: dataDir + "/patternKeyIds.txt",
	}
	pk.regexRes = make([]*regexp.Regexp, 0)
	pk.regexPoses = make(map[*regexp.Regexp]int)
	pk.testMode = testMode

	// Compile regexes and store them in the patternkeys
	for _, classRegex := range pk.regexes {
		re := regexp.MustCompile(`` + classRegex)
		pk.regexRes = append(pk.regexRes, re)
		names := re.SubexpNames()
		for i, name := range names {
			if name == cPatternKey {
				pk.regexPoses[re] = i
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

	return pk, nil
}

func (pk *patternkeys) register(patternKeyId string) {
	// Register the patternKeyId in the Aho-Corasick automaton
	pk.ac.Register(patternKeyId)
	if _, exists := pk.records[patternKeyId]; !exists {
		pk.records[patternKeyId] = make([]patternkey, 0, 10)
	}
}

func (pk *patternkeys) findAndRegister(line string) (string, bool, error) {
	patternKeyId := ""
	matched := false
	// Find the first matching regex and register the patternKeyId
	for _, re := range pk.regexRes {
		ma := re.FindStringSubmatch(line)
		if len(ma) > 0 {
			if pk.regexPoses[re] >= 0 && pk.regexPoses[re] < len(ma) {
				patternKeyId = ma[pk.regexPoses[re]]
				if patternKeyId != "" {
					matched = true
					pk.register(patternKeyId)
				}
			}
		}
	}
	return patternKeyId, matched, nil
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

func (pk *patternkeys) loadpatternkeysFromDb(searchKeyIds []string) error {
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
			fmt.Printf("%s: {startEpoch: %d, count: %d}\n", si.keyId, si.startEpoch, si.count)
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
