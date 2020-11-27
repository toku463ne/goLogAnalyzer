package analyzer

func getRarityAnalDB(baseDir string, maxPartitions int) (*csvDB, error) {
	d := map[string]csvTableDef{
		"lastStatus": csvTableDef{
			"lastStatus",
			[]string{"lastRowID", "lastBlockID", "fileName", "lastRow",
				"modifiedEpoch", "modifiedUtcTime"},
			0},
		"logBlocks": csvTableDef{
			"logBlocks",
			[]string{"blockID", "lastRowID", "blockCnt", "scoreSum",
				"scoreSqrSum", "createdAt"},
			maxPartitions},
		"items": csvTableDef{
			"items",
			[]string{"word", "cnt"},
			maxPartitions},
	}
	return newCsvDB(baseDir, d)
}

func getDCIClosedDB(baseDir string, maxPartitions int) (*csvDB, error) {
	d := map[string]csvTableDef{
		"frequentItemSets": csvTableDef{
			"frequentItemSets",
			[]string{"key", "itemSets", "support"},
			maxPartitions},
		"frequentItemFirstLines": csvTableDef{
			"frequentItemFirstLines",
			[]string{"line"},
			maxPartitions},
		"frequentItemLastLines": csvTableDef{
			"frequentItemLastLines",
			[]string{"line"},
			maxPartitions},
		"frequentItemSetsDotted": csvTableDef{
			"frequentItemSetsDotted",
			[]string{"line"},
			maxPartitions},
		"frequentItemSetsAbsent": csvTableDef{
			"frequentItemSetsAbsent",
			[]string{"line"},
			maxPartitions},
	}
	return newCsvDB(baseDir, d)
}
