package analyzer

const (
	cRModePlain      = "plain"
	cRModeGZip       = "gzip"
	cTimestampLayout = "2006-01-02 15:04:05"
	cIPReStr         = `[0-9]+\.[0-9]+\.[0-9]+.[0-9]+`
	cInSenErr        = `(?i)error`
	//wordReStr = `[\pL\p{Mc}\p{Mn}.%]{2,}`
	cWordReStr = `[0-9\pL\p{Mc}\p{Mn}.%]{2,}`
	//wordSegmenter = regexp.MustCompile(`[0-9\pL\p{Mc}\p{Mn}-_'.]+`)
	//wordSegmenter         = regexp.MustCompile(`[\pL\p{Mc}\p{Mn}.%]+`)
	cLogLevelError              = 1
	cLogLevelInfo               = 2
	cLogLevelDebug              = 3
	cDefaultBlockSize           = 100000
	cDefaultMaxBlocks           = 1000
	cDefaultRowsRateToShow      = 0.00005
	cDefaultMinNBlockRateForGap = 0.01
	cMaxRecsToProcessFrq        = 10000
	cEOF                        = -1
	cCountbyScoreLen            = 100
	cMaxTermLength              = 128
	cMinGapToRecord             = 0.2
	cMaxBlockDitigs             = 10
	cMaxRowID                   = int64(9999999999999)
	cMaxStageCounts             = 100
	cRowsToInsertAtOnce         = 10000
	cPrintClosedSetNum          = 100
	cNTopRareRecords            = 5
	cLastTmpBlockID             = -9999
	cLastTmpBlockStr            = "LASTTMPBLOCK"
)
