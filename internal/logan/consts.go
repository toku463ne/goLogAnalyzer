package logan

import (
	"goLogAnalyzer/pkg/utils"
)

const (
	CDefaultBlockSize           = 0
	CDefaultNBlocks             = 0
	CDefaultNItemBlocks         = 0
	CDefaultKeepPeriod          = 30
	CDefaultKeepUnit            = int64(utils.CFreqDay)
	CDefaultMinMatchRate        = 0.7
	CDefaultTermCountBorderRate = 0.999
	CDefaultTermCountBorder     = 0
	CDefaultSeparators          = ` "'\\,;[]<>{}=()|:&?/+!@`
	CDefaultStdThreshold        = 2
	CDefaultMinOccurrences      = 10
	CFileFormatJson             = "json"
	CFileFormatCsv              = "csv"

	cAsteriskItemID          = -1
	cMaxNumDigits            = 3 // HTTP codes
	cLogPerLines             = 1000000
	cStageRegisterTerms      = 1
	cStageRegisterLogStrings = 2
	cMaxLogGroups            = 1000000
)
