package analyzer

import "regexp"

const ()

var (
	remTags   = regexp.MustCompile(`<[^>]*>`)
	oneSpace  = regexp.MustCompile(`\s{2,}`)
	numberRe  = regexp.MustCompile(`^[0-9]+$`)
	ipReStr   = `[0-9]+\.[0-9]+\.[0-9]+.[0-9]+`
	inSenErr  = `(?i)error`
	wordReStr = `[\pL\p{Mc}\p{Mn}.%]{2,}`
	//wordSegmenter = regexp.MustCompile(`[0-9\pL\p{Mc}\p{Mn}-_'.]+`)
	//wordSegmenter         = regexp.MustCompile(`[\pL\p{Mc}\p{Mn}.%]+`)

	autoIncreaseSize      = 1000
	autoIncreaseSizeSmall = 1000
	bolShowProgress       = true
	printClosedSetNum     = 20
)
