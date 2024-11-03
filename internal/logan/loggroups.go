package logan

import (
	"bufio"
	"fmt"
	"goLogAnalyzer/pkg/csvdb"
	"goLogAnalyzer/pkg/utils"
	"io"
	"os"
	"strconv"
	"strings"
)

type logGroup struct {
	displayString string // lastValue with rare terms replaced with "*"
	count         int    // total count of this log group
	retentionPos  int64
	created       int64 // first epoch in the current block
	updated       int64 // last epoch in the current block
	countHistory  map[int64]int
}

type logGroups struct {
	*csvdb.CircuitDB
	maxLgId         int64
	totalCount      int // total count of entire log groups
	alllg           map[int64]*logGroup
	curlg           map[int64]*logGroup
	lt              *logTree
	retentionPos    int64
	maxRetentionPos int64
	minRetentionPos int64
	displayStrings  map[int64]string
}

func newLogGroups(dataDir string,
	maxBlocks int,
	unitSecs, keepPeriod int64,
	useGzip bool) (*logGroups, error) {

	lgs := new(logGroups)
	lgdb, err := csvdb.NewCircuitDB(dataDir, "logGroups",
		tableDefs["logGroups"], maxBlocks, 0, keepPeriod, unitSecs, useGzip)
	if err != nil {
		return nil, err
	}
	lgs.CircuitDB = lgdb

	lgs.maxLgId = 0
	lgs.totalCount = 0
	lgs.retentionPos = 0
	lgs.alllg = make(map[int64]*logGroup)
	lgs.curlg = make(map[int64]*logGroup)
	lgs.displayStrings = make(map[int64]string)
	lgs.lt = newLogTree(0)
	return lgs, nil
}

// generate group id in {epoch}-{lgid}-{randomNumber} format
// lgid is the lgid of the first time this logGroup registerd
func (lgs *logGroups) _genGroupId(created int64) int64 {
	lgs.maxLgId++
	lgid := lgs.maxLgId
	lgid = lgid % 1e9
	return created*1e9 + lgid
}

// Register logGroup info
func (lgs *logGroups) _registerLg(lgmap map[int64]*logGroup,
	groupId int64, retentionPos int64,
	addCnt int, displayString string,
	created, updated int64) int64 {
	var lg *logGroup
	ok := true

	// get logGroup if exists or create a new
	if groupId <= 0 {
		groupId = lgs._genGroupId(created)
		lg = new(logGroup)
		lgmap[groupId] = lg
	} else {
		lg, ok = lgmap[groupId]
		if !ok {
			lg = new(logGroup)
			lgmap[groupId] = lg
		}
	}
	lg.retentionPos = retentionPos
	lg.displayString = displayString
	lgs.displayStrings[groupId] = displayString

	if lg.created == 0 || created < lg.created {
		lg.created = created
	}
	if updated > lg.updated {
		lg.updated = updated
	}

	lg.count += addCnt

	return groupId
}

// register the item and return groupId
func (lgs *logGroups) registerLogTree(tokens []int,
	addCnt int, displayString string,
	created, updated int64, isNew bool, retentionPos int64, groupId int64) int64 {
	// register the tokens to logTree
	lt := lgs.lt.registerTokens(tokens)
	if groupId <= 0 {
		groupId = lt.groupId
	}

	groupId = lgs._registerLg(lgs.alllg, groupId, retentionPos,
		addCnt, displayString, created, updated)

	if lt.groupId <= 0 {
		lt.groupId = groupId
	}

	if isNew {
		lgs._registerLg(lgs.curlg, lt.groupId, retentionPos,
			addCnt, displayString, created, updated)
	}
	lgs.totalCount += addCnt

	return lt.groupId
}

func (lgs *logGroups) flush() error {
	if lgs.DataDir == "" {
		return nil
	}
	for groupId, lg := range lgs.curlg {
		if lg.count <= 0 {
			continue
		}
		//{"groupId", "retentionPos", "count", "created", "updated"}
		if err := lgs.InsertRow(tableDefs["logGroups"],
			//utils.Int64Tobase36(groupId), lg.retentionPos, lg.count, lg.created, lg.updated); err != nil {
			groupId, lg.retentionPos, lg.count, lg.created, lg.updated); err != nil {
			return err
		}
	}

	if err := lgs.FlushOverwriteCurrentTable(); err != nil {
		return err
	}

	lgs.curlg = make(map[int64]*logGroup)
	return nil
}

func (lgs *logGroups) next(updated int64) error {
	if err := lgs.flush(); err != nil {
		return err
	}
	if err := lgs.NextBlock(updated); err != nil {
		return err
	}
	lgs.maxLgId = 0
	return nil
}

func (lgs *logGroups) _getDisplayStringPath() string {
	return fmt.Sprintf("%s/displaystrings.txt", lgs.DataDir)
}

func (lgs *logGroups) writeDisplayStrings() error {
	file, err := os.Create(lgs._getDisplayStringPath())
	if err != nil {
		return fmt.Errorf("error creating file: +%v", err)
	}
	defer file.Close()
	writer := bufio.NewWriter(file)
	for groupId, lg := range lgs.alllg {
		if lg.count <= 0 {
			continue
		}
		//_, err := writer.WriteString(fmt.Sprintf("%s %s\n", utils.Int64Tobase36(groupId), lg.displayString))
		_, err := writer.WriteString(fmt.Sprintf("%d %s\n", groupId, lg.displayString))
		if err != nil {
			return fmt.Errorf("error creating file: +%v", err)
		}
	}
	writer.Flush()

	return nil
}

func (lgs *logGroups) readDisplayStrings() error {
	displayStrings := make(map[int64]string)

	file, err := os.Open(lgs._getDisplayStringPath())
	if err != nil {
		return err

	}
	defer file.Close()

	reader := bufio.NewReader(file)

	lineno := 0
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break // End of file
			}
			return err
		}

		lineno++
		line = strings.TrimSpace(line)
		parts := strings.SplitN(line, " ", 2)
		if len(parts) < 2 {
			return utils.ErrorStack("error at line %d: missing data", lineno)
		}

		groupId, err := strconv.ParseInt(parts[0], 10, 64)
		if err != nil {
			return utils.ErrorStack("error at line %d: %+v", lineno, err)
		}
		displayStrings[groupId] = parts[1]
	}

	lgs.displayStrings = displayStrings
	return nil
}

func (lgs *logGroups) commit(completed bool) error {
	if lgs.DataDir == "" {
		return nil
	}
	if err := lgs.flush(); err != nil {
		return err
	}

	if err := lgs.writeDisplayStrings(); err != nil {
		return err
	}

	if err := lgs.UpdateBlockStatus(completed); err != nil {
		return err
	}
	return nil
}

// read from specified block into a map[string]logGroup
// displayString will not be loaded
// mainly for testing
func (lgs *logGroups) getBlockData(blockNo int) (map[string]logGroup, error) {
	table, err := lgs.GetBlockTable(blockNo)
	if err != nil {
		return nil, err
	}

	if err := lgs.readDisplayStrings(); err != nil {
		return nil, err
	}

	rows, err := table.SelectRows(nil, nil)
	if err != nil {
		return nil, err
	}
	if rows == nil {
		return nil, nil
	}

	blockLgs := make(map[string]logGroup)
	for rows.Next() {
		var groupIdstr string
		var retentionPos int64
		var count int
		var created int64
		var updated int64
		err = rows.Scan(&groupIdstr, &retentionPos, &count, &created, &updated)
		if err != nil {
			return nil, err
		}
		lg := logGroup{
			retentionPos: retentionPos,
			count:        count,
			created:      created,
			updated:      updated,
		}
		//groupId, err := utils.Base36ToInt64(groupIdstr)
		groupId, err := strconv.ParseInt(groupIdstr, 10, 64)
		if err != nil {
			return nil, err
		}
		lg.displayString = lgs.displayStrings[groupId]
		blockLgs[groupIdstr] = lg
	}

	return blockLgs, nil
}

// load countHistory.
// call this function only when needed as it eats memory
func (lgs *logGroups) loadLogGroupHistory() error {
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

		displayString := ds[groupId]
		lg, ok := lgs.alllg[groupId]
		if !ok {
			lg = new(logGroup)
			lg.countHistory = make(map[int64]int)
		} else if lg.countHistory == nil {
			lg.countHistory = make(map[int64]int)
		}
		lg.countHistory[retentionPos] = count
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
