package analyzer

const (
	cRModePlain      = "plain"
	cRModeGZip       = "gzip"
	cTimestampLayout = "2006-01-02 15:04:05"
	cIPReStr         = `[0-9]+\.[0-9]+\.[0-9]+.[0-9]+`
	cWordReStr       = `[0-9\pL\p{Mc}\p{Mn}.%]{2,}`
	cWordMaxLen      = 40 // IPv6
	cNumMaxDigits    = 3  // HTTP codes
	cLogLevelError   = 1
	cLogLevelInfo    = 2
	cLogLevelDebug   = 3
	cCountbyScoreLen = 100
	cEOF             = -1
	cMaxTermLength   = 128
	cMaxBlockDitigs  = 10
	cMaxRowID        = int64(9999999999999)
	cNTopRareRecords = 5
	cLogPerLines     = 1000000
	cDefaultBuffSize = 10000
	cDefaultHistSize = 5
	cCountBorderRate = 0.01
)
