package csvdb

import (
	"compress/gzip"
	"encoding/csv"
	"os"
)

type CsvDB struct {
	Groups  map[string]*CsvTableGroup
	baseDir string
}

type CsvTableGroup struct {
	groupName      string
	rootDir        string
	dataDir        string
	iniFile        string
	tableDefs      map[string]*CsvTableDef
	columns        []string
	useGzip        bool
	bufferSize     int
	readBufferSize int
}

type CsvTableDef struct {
	groupName string
	tableName string
	path      string
}

type CsvTable struct {
	*CsvTableDef
	columns        []string
	colMap         map[string]int
	useGzip        bool
	bufferSize     int
	readBufferSize int
	path           string
	iBuff          *insertBuff
	reader         *CsvReader
}

type CsvRows struct {
	reader             *CsvReader
	selectedColIndexes []int
	tableCols          []string
	conditionCheckFunc func([]string) bool
	orderbyBuff        orderBuffRows
	orderbyBuffPos     int
	orderbyExecuted    bool
	orderbyErr         error
}

type insertBuff struct {
	rows   [][]string
	pos    int
	isFull bool
	size   int
}

type readBuff struct {
	path     string
	rows     [][]string
	pos      int
	readPos  int
	pageSize int
	values   []string
}

type CsvReader struct {
	fr       *os.File
	zr       *gzip.Reader
	reader   *csv.Reader
	values   []string
	err      error
	filename string
	mode     string
	readBuff *readBuff
}

type CsvWriter struct {
	fw     *os.File
	zw     *gzip.Writer
	writer *csv.Writer
	path   string
	mode   string
}

type orderBuffRow struct {
	v               []string
	orderFieldTypes []string
	orderFieldIdxs  []int
	direction       int
}

type orderBuffRows []orderBuffRow
