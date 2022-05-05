package analyzer

var (
	useGzipInCircuitTables = true
	tranMatchRates         = []tranMatchRate{
		{
			matchLen:  1,
			matchRate: 1,
		},
		{
			matchLen:  5,
			matchRate: 0.8,
		},
		{
			matchLen:  20,
			matchRate: 0.7,
		},
	}
	IsDebug = false
)
