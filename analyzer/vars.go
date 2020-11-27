package analyzer

import (
	"errors"
	"regexp"
)

var (
	//cfg *ini.File
	//rootDir = "."
	//logName string
	//logHost               string
	//logPathRegex          string
	//filterRe              string
	//xFilterRe             string
	//rarityThreshold float64
	//linesInBlock          int
	//maxBlocks             int
	maxBlockDitigs = 10
	//minSupportPerBlock = 0.1
	verbose = false
	//isIniLoaded = false
	//frequencyCheck        = true
	remTags               = regexp.MustCompile(`<[^>]*>`)
	oneSpace              = regexp.MustCompile(`\s{2,}`)
	numberRe              = regexp.MustCompile(`^[0-9]+$`)
	reNewline             = regexp.MustCompile(`\r\n|\r|\n`)
	errNotInitialized     = errors.New("Not initialized")
	errEndOfCursor        = errors.New("End Of Cursor")
	errNoRecords          = errors.New("No records have found")
	errNoFileMatched      = errors.New("No files matched")
	curLogLevel           = cLogLevelInfo
	autoIncreaseSize      = 1000
	autoIncreaseSizeSmall = 1000
	bolShowProgress       = true
	printClosedSetNum     = 20
	maxBitMatrixXLen      = 10000
	itemsInsertCntInOneQ  = 100
	sqliteDBLockWaitCnt   = 60
)
