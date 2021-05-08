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
	errNotInitialized = errors.New("not initialized")
	errEndOfCursor    = errors.New("end Of Cursor")
	errNoRecords      = errors.New("no records have found")
	errNoFileMatched  = errors.New("no files matched")
)
