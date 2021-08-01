package analyzer

import (
	"database/sql"
	"regexp"
	"strings"

	_ "github.com/mattn/go-sqlite3"
	csvdb "github.com/toku463ne/goCsvDb"
)

type csvdbObj struct {
	dataDir string
	dbName  string
	*csvdb.CsvDB
}

type sqliteObj struct {
	dataDir string
	dbName  string
	conn    *sql.DB
}

type colStats struct {
	scoreSum    float64
	scoreSqrSum float64
	scoreCount  int64
}

type stats struct {
	*sqliteObj
	*colStats
	rootDir        string
	currBlock      *colStats
	countPerScore  []int
	maxBlocks      int
	maxRowsInBlock int
	blockNo        int
	rowNo          int
	seqNo          int64
	lastAverage    float64
	lastStd        float64
	lastGap        float64
}

type circuitDB struct {
	*csvdbObj
	name           string
	dataDir        string
	maxBlocks      int
	maxRowsInBlock int
	blockNo        int
	rowNo          int
	lastIndex      int64
	lastEpoch      int64
	currTable      *csvdb.CsvTable
	statusTable    *csvdb.CsvTable
	writeMode      string
}

type circuitRows struct {
	tableNames []string
	rows       *csvdb.CsvRows
	pos        int
	err        error
	*csvdbObj
	columns            []string
	conditionCheckFunc func([]string) bool
	blockCompleted     bool
	completedIdx       int
	blockIDIdx         int
	statusTable        *csvdb.CsvTable
}

type colItems struct {
	item      string
	itemCount int
}

type colLogRecords struct {
	rowid  int64
	score  float64
	record string
}

type logRecords struct {
	*circuitDB
}

type trans struct {
	items          *items
	maxRowsInBlock int
	maxTranID      int
	replacer       *strings.Replacer
}

type items struct {
	*circuitDB
	maxItemID          int
	maxRowsInItemBlock int
	terms              map[string]int
	termMap            map[int]string
	counts             map[int]int
	scores             map[int]float64
	currCounts         map[int]int
	totalCount         int
}

type rarityAnalyzer struct {
	*sqliteObj
	rootDir        string
	trans          *trans
	stats          *stats
	logRecs        *logRecords
	fp             *filePointer
	logPathRegex   string
	filterRe       *regexp.Regexp
	xFilterRe      *regexp.Regexp
	minGapToRecord float64
	lastFileEpoch  int64
	lastFileRow    int
	rowID          int64
	linesInBlock   int
	maxBlocks      int
	maxItemBlocks  int
}
