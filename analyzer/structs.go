package analyzer

import (
	"database/sql"
	"regexp"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

type db struct {
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
	*db
	*colStats
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
	*db
	maxBlocks      int
	maxRowsInBlock int
	blockNo        int
	rowNo          int
	lastIndex      int64
	lastEpoch      int64
}

type circuitRows struct {
	tableNames     []string
	rows           *sql.Rows
	pos            int
	err            error
	db             *db
	fields         string
	conds          string
	blockCompleted bool
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
	rows       []colLogRecords
	startRowNo int
}

type trans struct {
	items          *items
	maxRowsInBlock int
	maxTranID      int
	replacer       *strings.Replacer
}

type items struct {
	*circuitDB
	maxItemID  int
	terms      map[string]int
	termMap    map[int]string
	counts     map[int]int
	scores     map[int]float64
	currCounts map[int]int
	totalCount int
}

type rarityAnalyzer struct {
	db             *db
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
