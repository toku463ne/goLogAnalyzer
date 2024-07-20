package analyzer

const (
	CDefaultTopNToShow      = 10
	CDefaultBlockSize       = 0
	CDefaultNBlocks         = 0
	CDefaultNItemBlocks     = 0
	CDefaultScoreStyle      = cScoreConstSizeAvg
	CDefaultScoreNSize      = 10
	CDefaultTimestampLayout = "2006-01-02 15:04:05"
	CDefaultMinGap          = 0
	CDefaultDaysToReport    = 3
	CDefaultNRareTerms      = 20
	CDefaultIgnoreCount     = 1

	cDefaultGroupName = "default"
	cIntTrue          = 1
	cIntFalse         = 2

	cMinNTopItemCount          = 2
	cMinNTopItemTermLen        = 3
	cMinTermApparenceInPhrases = 10
	cMinTermInPhrases          = 3

	cErrPathNotExists     = "path not exists"
	cRModePlain           = "plain"
	cRModeGZip            = "gzip"
	cIPReStr              = `[0-9]+\.[0-9]+\.[0-9]+.[0-9]+`
	cWordReStr            = `[0-9\pL\p{Mc}\p{Mn}.%]{2,}`
	cWordMaxLen           = 40 // IPv6
	cNumMaxDigits         = 3  // HTTP codes
	cLogLevelError        = 1
	cLogLevelInfo         = 2
	cLogLevelDebug        = 3
	cCountbyScoreLen      = 100
	cEOF                  = -1
	cMaxTermLength        = 128
	cMaxBlockDitigs       = 10
	cLogCycle             = 14
	cMaxRowID             = int64(9223372036854775806)
	cNTopRareRecords      = 5
	cLogPerLines          = 1000000
	cDefaultBuffSize      = 10000
	cCountBorderRate      = 0.01
	cErrorKeywords        = "failure|failed|error|down|crit"
	cFormatHtml           = "html"
	cFormatText           = "txt"
	cScoreSimpleAvg       = 1
	cScoreNAvg            = 2
	cScoreNDistAvg        = 3
	cScoreConstSizeAvg    = 4
	cMaxCharsToShowInTopN = 400
	cNTopMultiplier       = 10
	cNFilesToCheckCount   = 5

	cMaxCountToShowInDigest = 50

	cHtmlRareEmphTag = "font color='blue'"
)
