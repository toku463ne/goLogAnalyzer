package analyzer

import (
	"regexp"
	"strings"

	csvdb "github.com/toku463ne/goCsvDb"
)

type colStats struct {
	scoreSum    float64
	scoreSqrSum float64
	scoreCount  int64
}

type colScoreshist struct {
	lastFileEpoch int64
	avg           float64
	std           float64
	max           float64
}

type stats struct {
	*csvdb.CsvDB
	*colStats
	statsTable      *csvdb.CsvTable
	scoresTable     *csvdb.CsvTable
	scoresHistTable *csvdb.CsvTable
	rootDir         string
	currBlock       *colStats
	countPerScore   []int
	maxBlocks       int
	maxRowsInBlock  int
	blockNo         int
	rowNo           int
	seqNo           int64
	lastAverage     float64
	lastStd         float64
	lastGap         float64
	lastFileEpoch   int64
	scoreMax        float64
}

type circuitDB struct {
	*csvdb.CsvDB
	dataDir        string
	name           string
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
	groupName  string
	tableNames []string
	rows       *csvdb.CsvRows
	pos        int
	err        error
	*csvdb.CsvDB
	columns            []string
	conditionCheckFunc func([]string) bool
	blockCompleted     bool
	completedIdx       int
	blockIDIdx         int
	statusTable        *csvdb.CsvTable
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
	currCounts         map[int]int
	totalCount         int
}

type rarityAnalyzer struct {
	*csvdb.CsvDB
	configTable     *csvdb.CsvTable
	lastStatusTable *csvdb.CsvTable
	rootDir         string
	trans           *trans
	stats           *stats
	logRecs         *logRecords
	fp              *filePointer
	logPathRegex    string
	filterRe        *regexp.Regexp
	xFilterRe       *regexp.Regexp
	minGapToRecord  float64
	lastFileEpoch   int64
	lastFileRow     int
	rowID           int64
	linesInBlock    int
	maxBlocks       int
	maxItemBlocks   int
}
