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

func getClosedItemsDB(baseDir string, maxPartitions int) (*csvDB, error) {
	d := map[string]csvTableDef{
		"closedItemSets": csvTableDef{
			"closedItemSets",
			[]string{"key", "itemSets", "support", "lastLine"},
			maxPartitions},
		"closedItemKeys": csvTableDef{
			"closedItemKeys",
			[]string{"key"},
			maxPartitions},
	}
	return newCsvDB(baseDir, d)
}
