package analyzer

import (
	"bufio"
	"compress/gzip"
	"os"
	"regexp"
	"strings"

	csvdb "goLogAnalyzer/csvdb"
)

type AnalConf struct {
	RootDir          string
	LogPathRegex     string
	BlockSize        int
	MaxBlocks        int
	MaxItemBlocks    int
	DatetimeStartPos int
	DatetimeLayout   string
	ScoreStyle       int
	ScoreNSize       int
	MinGapToRecord   float64
	NTopRecordsCount int
	ModeblockPerFile bool // if create block per file
	NItemTop         int
}

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
	blockSize       int
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
	dataDir     string
	name        string
	maxBlocks   int
	blockSize   int
	blockNo     int
	rowNo       int
	lastIndex   int64
	lastEpoch   int64
	currTable   *csvdb.CsvTable
	statusTable *csvdb.CsvTable
	writeMode   string
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
	rowid    int64
	score    float64
	maxScore float64
	record   string
	tran     []int
	count    int
	lastDate string
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
	lastRowId  int64
}

type nTopOutRec struct {
	rowid    int64
	score    float64
	record   string
	count    int
	lastDate string
}

type logRecords struct {
	*circuitDB
}

type trans struct {
	items            *items
	blockSize        int
	replacer         *strings.Replacer
	datetimeStartPos int
	datetimeEndPos   int
	datetimeLayout   string
	scoreStyle       int
	scoreNSize       int
	lastTimeResult   []int
}

type topNItems struct {
	n              int
	minScoreInTopN float64
	itemIDs        []int
	scores         []float64
}

type items struct {
	*circuitDB
	maxItemID          int
	maxRowsInItemBlock int
	terms              map[string]int
	termMap            map[int]string
	counts             map[int]int
	tranScoreAvg       map[int]float64
	currCounts         map[int]int
	totalCount         int
}

type rarityAnalyzer struct {
	*csvdb.CsvDB
	*AnalConf
	configTable     *csvdb.CsvTable
	lastStatusTable *csvdb.CsvTable
	trans           *trans
	stats           *stats
	logRecs         *logRecords
	fp              *filePointer
	filterRe        *regexp.Regexp
	xFilterRe       *regexp.Regexp
	lastFileEpoch   int64
	lastFileRow     int
	rowID           int64
	nTopRareLogs    *nTopRecords
	linesProcessed  int
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

type report struct {
	st         *stats
	conf       *LogConfRoot
	confGroups *logConfGroups
}

type LogConf struct {
	LogPath          string              `json:"path"`
	TopN             int                 `json:"topN"`
	ScoreStyle       int                 `json:"scoreStyle"`
	Search           string              `json:"search"`
	Exclude          string              `json:"exclude"`
	BlockSize        int                 `json:"blockSize"`
	MaxBlocks        int                 `json:"maxBlocks"`
	MaxItemBlocks    int                 `json:"maxItemBlocks"`
	MinGapToRecord   float64             `json:"minGapToRecord"`
	DatetimeStartPos int                 `json:"dateStartPos"`
	DatetimeLayout   string              `json:"dateLayout"`
	FromDate         string              `json:"fromDate"`
	ToDate           string              `json:"toDate"`
	KeyEmphasize     map[string][]string `json:"keyEmphasize"`
	ModeblockPerFile int                 `json:"modeblockPerFile"`
	MinScore         float64             `json:"minScore"`
	MaxScore         float64             `json:"maxScore"`
}

type LogNode struct {
	*LogConf
	Name         string     `json:"name"`
	TemplateName string     `json:"templateName"`
	GroupNames   []string   `json:"groupNames"`
	Children     []*LogNode `json:"children"`
	Categories   []*LogNode `json:"categories"`
	isCategory   bool
	isEnd        bool
	dataDir      string
	reportDir    string
}

type LogConfRoot struct {
	*LogConf
	RootDir   string              `json:"rootDir"`
	ReportDir string              `json:"reportDir"`
	Templates map[string]*LogConf `json:"templates"`
	Children  []*LogNode          `json:"children"`
}

type logConfGroups struct {
	g         map[string][]*LogNode
	reportDir string
}

type tranMatchRate struct {
	matchLen  int
	matchRate float64
}
