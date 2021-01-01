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
	cLogLevelError          = 1
	cLogLevelInfo           = 2
	cLogLevelDebug          = 3
	cDefaultBlockSize       = 10000
	cDefaultMaxBlocks       = 1000
	cDefaultBlockSizeNoDb   = 1000
	cDefaultMaxBlocksNoDb   = 100
	cDefaultRarityThreshold = 0.8
	cMaxRecsToProcessFrq    = 10000
	cEOF                    = -1
	cCountbyGapLen          = 100
)
