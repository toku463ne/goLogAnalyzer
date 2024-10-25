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
	CDefaultMinMatchRate        = 0.6
	CDefaultTermCountBorderRate = 0.999
	CDefaultTermCountBorder     = 0

	cAsteriskItemID          = -1
	cMaxNumDigits            = 3 // HTTP codes
	cDelimiters              = `[\r\n\t"'\\,;[\]<>{}=\(\)|:&\?+/\s!.]+`
	cLogPerLines             = 1000000
	cStageRegisterTerms      = 1
	cStageRegisterLogStrings = 2
)
