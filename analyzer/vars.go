package analyzer

var (
	useGzipInCircuitTables = true
	tranMatchRates         = []tranMatchRate{
		{
			matchLen:  1,
			matchRate: 0.8,
		},
		{
			matchLen:  5,
			matchRate: 0.7,
		},
		{
			matchLen:  10,
			matchRate: 0.6,
		},
		{
			matchLen:  20,
			matchRate: 0.5,
		},
	}
	IsDebug                 = false
	Name                    string
	termAppearenceInPhrases = []int{10, 100, 1000}
)
