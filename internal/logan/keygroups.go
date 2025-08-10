package logan

import (
	"bufio"
	"goLogAnalyzer/pkg/csvdb"
	"goLogAnalyzer/pkg/utils"
	"os"
	"regexp"
)

type keygroup struct {
	epoch   int64
	groupId int64 // logGroupId
}

type keygroups struct {
	*csvdb.CircuitDB
	ac         *utils.AC
	regexRes   []*regexp.Regexp
	regexPoses map[*regexp.Regexp]int
	regexes    []string
	records    map[string][]keygroup
	testMode   bool
	idFilePath string // path to the keygroup IDs file
}

// NewKeyGroups creates a new keygroups instance
func newKeyGroups(dataDir string, regexes []string, useGzip bool, testMode bool) (*keygroups, error) {

	kg := &keygroups{
		ac:         utils.NewAC(),
		regexRes:   make([]*regexp.Regexp, 0),
		regexPoses: make(map[*regexp.Regexp]int),
		regexes:    regexes,
		records:    make(map[string][]keygroup, 10000),
		idFilePath: dataDir + "/keygroupIds.txt",
	}
	kg.regexRes = make([]*regexp.Regexp, 0)
	kg.regexPoses = make(map[*regexp.Regexp]int)
	kg.testMode = testMode

	// Compile regexes and store them in the keygroups
	for _, classRegex := range kg.regexes {
		re := regexp.MustCompile(`` + classRegex)
		kg.regexRes = append(kg.regexRes, re)
		names := re.SubexpNames()
		for i, name := range names {
			if name == cKeyGroupId {
				kg.regexPoses[re] = i
			}
		}
	}

	if kg.testMode {
		return kg, nil
	}

	kgdb, err := csvdb.NewCircuitDB(dataDir, "keygroups",
		tableDefs["keygroups"], 0, 0, 0, 0, useGzip)
	if err != nil {
		return nil, err
	}
	kg.CircuitDB = kgdb

	return kg, nil
}

func (kg *keygroups) register(kgId string) {
	// Register the kgId in the Aho-Corasick automaton
	kg.ac.Register(kgId)
	if _, exists := kg.records[kgId]; !exists {
		kg.records[kgId] = make([]keygroup, 0, 10)
	}
}

func (kg *keygroups) findAndRegister(line string) (string, error) {
	keygroupId := ""
	// Find the first matching regex and register the kgId
	for _, re := range kg.regexRes {
		ma := re.FindStringSubmatch(line)
		if len(ma) > 0 {
			if kg.regexPoses[re] >= 0 && kg.regexPoses[re] < len(ma) {
				keygroupId = ma[kg.regexPoses[re]]
				if keygroupId != "" {
					kg.register(keygroupId)
				}
			}
		}
	}
	return keygroupId, nil
}

func (kg *keygroups) hasMatch(term []byte) bool {
	// Check if the term matches any of the registered kgIds
	return len(kg.ac.MatchExact(term)) > 0
}

func (kg *keygroups) appendLogGroup(kgId string, epoch, logGroupId int64) {
	// append a log group ID to the list for the given kgId
	// Do not check the existance of kgId in the records map. It indicates a bug if it is not registered
	kg.records[kgId] = append(kg.records[kgId], keygroup{epoch: epoch, groupId: logGroupId})
}

func (kg *keygroups) flush() error {
	if kg.DataDir == "" || kg.testMode {
		return nil
	}

	// insert the keygroups into the CSV DB
	for kgId, records := range kg.records {
		for _, rec := range records {
			// inserts to buffer only
			if err := kg.InsertRow(tableDefs["keygroups"],
				kgId, rec.epoch, rec.groupId); err != nil {
				return err
			}
		}
	}

	// flush buffer to the block table file with WRITE mode (not APPEND)
	if err := kg.FlushOverwriteCurrentTable(); err != nil {
		return err
	}

	// write the keygroup IDs to the file
	if err := utils.WriteStringToFile(kg.idFilePath, kg.ac.GetPatterns()); err != nil {
		return err
	}

	return nil
}

func (kg *keygroups) next(updated int64) error {
	// save logGroup IDs to CSV DB
	// write current block to the block table
	if err := kg.flush(); err != nil {
		return err
	}
	if err := kg.NextBlock(updated); err != nil {
		return err
	}

	// clear the logGroupIds map for the next block
	kg.records = make(map[string][]keygroup, 10000)

	return nil
}

func (kg *keygroups) load() error {
	cnt := kg.CountFromStatusTable(nil)
	if cnt <= 0 {
		return nil
	}

	if err := kg.LoadCircuitDBStatus(); err != nil {
		return err
	}

	// read the keygroup IDs from the file
	file, err := os.Open(kg.idFilePath)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		kg.register(scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	// load from the last block table
	trows, err := kg.SelectFromCurrentTable(nil, tableDefs["keygroups"])
	if err != nil {
		return err
	}
	if trows == nil {
		return nil
	}
	for trows.Next() {
		var kgId string
		var epoch int64
		var logGroupId int64
		if err := trows.Scan(&kgId, &epoch, &logGroupId); err != nil {
			return err
		}
		kg.appendLogGroup(kgId, epoch, logGroupId)
	}

	return nil
}
