package analyzer

import (
	"regexp"
)

var (
	curLogLevel = cLogLevelInfo
	remTags     = regexp.MustCompile(`<[^>]*>`)
	oneSpace    = regexp.MustCompile(`\s{2,}`)
	numberRe    = regexp.MustCompile(`^[0-9]+$`)
	reNewline   = regexp.MustCompile(`\r\n|\r|\n`)
)
