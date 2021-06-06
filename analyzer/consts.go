package analyzer

const (
	cRModePlain          = "plain"
	cRModeGZip           = "gzip"
	cTimestampLayout     = "2006-01-02 15:04:05"
	cIPReStr             = `[0-9]+\.[0-9]+\.[0-9]+.[0-9]+`
	cWordReStr           = `[0-9\pL\p{Mc}\p{Mn}.%]{2,}`
	cWordMaxLen          = 40 // IPv6
	cLogLevelError       = 1
	cLogLevelInfo        = 2
	cLogLevelDebug       = 3
	cDefaultBlockSize    = 10000
	cDefaultMaxBlocks    = 1000
	cMaxRecsToProcessFrq = 10000
	cEOF                 = -1
	cCountbyScoreLen     = 100
	cMaxTermLength       = 128
	cMinGapToRecord      = 0.2
	cMaxBlockDitigs      = 10
	cMaxRowID            = int64(9999999999999)
	cPrintClosedSetNum   = 100
	cNTopRareRecords     = 5
	cLastTmpBlockID      = -9999
	cLastTmpBlockStr     = "LASTTMPBLOCK"
)
