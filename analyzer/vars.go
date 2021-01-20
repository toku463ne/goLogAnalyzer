package analyzer

import (
	"errors"
	"regexp"
)

var (
	verbose           = false
	curLogLevel       = cLogLevelInfo
	remTags           = regexp.MustCompile(`<[^>]*>`)
	oneSpace          = regexp.MustCompile(`\s{2,}`)
	numberRe          = regexp.MustCompile(`^[0-9]+$`)
	reNewline         = regexp.MustCompile(`\r\n|\r|\n`)
	errNotInitialized = errors.New("Not initialized")
	errEndOfCursor    = errors.New("End Of Cursor")
	errNoRecords      = errors.New("No records have found")
	errNoFileMatched  = errors.New("No files matched")
)
