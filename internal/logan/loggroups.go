package logan

import (
	"bufio"
	"fmt"
	"goLogAnalyzer/pkg/csvdb"
	"goLogAnalyzer/pkg/utils"
	"os"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

type logGroup struct {
	displayString string // lastValue with rare terms replaced with "*"
	count         int    // total count of this log group
	retentionPos  int
	created       int64 // first epoch in the current block
	updated       int64 // last epoch in the current block
}

type logGroups struct {
	*csvdb.CircuitDB
	maxLgId        int64
	totalCount     int // total count of entire log groups
	alllg          map[int64]*logGroup
	curlg          map[int64]*logGroup
	lt             *logTree
	retentionPos   int
	displayStrings map[int64]string
}

func newLogGroups(dataDir string,
	maxBlocks, blockSize,
	keepUnit int, keepPeriod int64,
	useGzip bool) (*logGroups, error) {

	lgs := new(logGroups)
	lgdb, err := csvdb.NewCircuitDB(dataDir, "logGroups",
		tableDefs["logGroups"], maxBlocks, blockSize, keepPeriod, keepUnit, useGzip)
	if err != nil {
		return nil, err
	}
	lgs.CircuitDB = lgdb

	lgs.maxLgId = 0
	lgs.totalCount = 0
	lgs.retentionPos = 0
	lgs.alllg = make(map[int64]*logGroup)
	lgs.curlg = make(map[int64]*logGroup)
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
	groupId int64, retentionPos int,
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
		}
	}
	lg.retentionPos = retentionPos
	lg.displayString = displayString

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
	created, updated int64, isNew bool, retentionPos int) int64 {
	// register the tokens to logTree
	lt := lgs.lt.registerTokens(tokens)

	groupId := lgs._registerLg(lgs.alllg, lt.groupId, retentionPos,
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
		if err := lgs.InsertRow(tableDefs["logGroups"],
			utils.Int64Tobase36(groupId), lg.retentionPos, lg.count, lg.created, lg.updated); err != nil {
			return err
		}
	}
	if err := lgs.FlushOverwriteCurrentTable(); err != nil {
		return errors.WithStack(err)
	}
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
		return errors.Errorf("error opening file: %+v", err)

	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	lineno := 0
	for scanner.Scan() {
		lineno++
		line := scanner.Text()
		parts := strings.SplitN(line, " ", 2)
		groupId, err := strconv.ParseInt(parts[0], 10, 64)
		if err != nil {
			return errors.Errorf("error at line %d: %+v", lineno, err)
		}
		displayStrings[groupId] = parts[1]
	}

	if err := scanner.Err(); err != nil {
		return errors.Errorf("error opening file: %+v", err)
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
	if err := lgs.UpdateBlockStatus(completed); err != nil {
		return err
	}
	return nil
}
