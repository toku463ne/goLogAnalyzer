package analyzer

import (
	"bufio"
	"compress/gzip"
	"os"
	"regexp"
	"strings"

	csvdb "github.com/toku463ne/goLogAnalyzer/csvdb"
)

type colStats struct {
	scoreSum    float64
	scoreSqrSum float64
	scoreCount  int64
}

type colScoresHist struct {
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

type colLogRecord struct {
	rowid  int64
	score  float64
	record string
	tran   []int
	count  int
	dates  []string
}

type nTopRecords struct {
	*csvdb.CsvDB
	name       string
	rootDir    string
	ntopTable  *csvdb.CsvTable
	n          int
	subN       int
	minScore   float64
	isUniqMode bool
	records    []*colLogRecord
	t          *trans
	memberCnt  int
	withDiff   bool
	diff       *nTopRecords
	lastRowId  int64
}

type nTopOutRec struct {
	rowid  int64
	score  float64
	record string
	count  int
	dates  []string
}

type logRecords struct {
	*circuitDB
}

type trans struct {
	items            *items
	maxRowsInBlock   int
	replacer         *strings.Replacer
	datetimeStartPos int
	datetimeEndPos   int
	datetimeLayout   string
	scoreStyle       int
	lastTimeResult   []int
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
	configTable      *csvdb.CsvTable
	lastStatusTable  *csvdb.CsvTable
	rootDir          string
	trans            *trans
	stats            *stats
	logRecs          *logRecords
	fp               *filePointer
	logPathRegex     string
	filterRe         *regexp.Regexp
	xFilterRe        *regexp.Regexp
	minGapToRecord   float64
	lastFileEpoch    int64
	lastFileRow      int
	rowID            int64
	linesInBlock     int
	maxBlocks        int
	maxItemBlocks    int
	nTopRecordsCount int
	nTopRareLogs     *nTopRecords
	datetimeStartPos int
	datetimeLayout   string
	scoreStyle       int
}

type filePointer struct {
	files    []string
	epochs   []int64
	r        *reader
	lastRow  int
	pos      int
	e        error
	currErr  error
	currText string
	currRow  int
	currPos  int
	isEOF    bool
}

type reader struct {
	fd       *os.File
	zr       *gzip.Reader
	reader   *bufio.Reader
	rowNum   int
	mode     string
	filename string
	e        error
	currText string
}

type countPerScore struct {
	score float64
	count int
}

type reports struct {
	dataDir string
	rep     map[string]*report
}

type report struct {
	name          string
	nTopNorm      *nTopRecords
	nTopErr       *nTopRecords
	includePhrase string
	st            *stats
	info          LogInfo
}

type LogInfo struct {
	DataDir          string  `json:"dataDir"`
	TopN             int     `json:"topN"`
	HistSize         int     `json:"histSize"`
	ScoreStyle       int     `json:"scoreStyle"`
	LogPath          string  `json:"path"`
	Search           string  `json:"search"`
	Exclude          string  `json:"exclude"`
	LinesInBlock     int     `json:"linesInBlock"`
	MaxBlocks        int     `json:"maxBlocks"`
	MaxItemBlocks    int     `json:"maxItemBlocks"`
	MinGapToRecord   float64 `json:"minGapToRecord"`
	DatetimeStartPos int     `json:"dateStart"`
	DatetimeLayout   string  `json:"dateLayout"`
}

type logInfoMap struct {
	*LogInfo
	Logs map[string]LogInfo `json:"logs"`
}
