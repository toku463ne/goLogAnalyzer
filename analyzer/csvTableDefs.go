package analyzer

import "fmt"

func getRarityAnalDB(baseDir string, maxPartitions int) (*csvDB, error) {
	countPerGapH := make([]string, cCountbyScoreLen+1)
	//countPerGapH[0] = "blockID"
	for i := 0; i < cCountbyScoreLen; i++ {
		countPerGapH[i] = fmt.Sprint(i - 1)
	}

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
		"countPerGap": csvTableDef{
			"countPerGap", countPerGapH,
			maxPartitions},
		"items": csvTableDef{
			"items",
			[]string{"word", "cnt"},
			maxPartitions},
		"logRecords": csvTableDef{
			"logRecords",
			[]string{"rowID", "score", "text"},
			maxPartitions},
		"nTopRareLogs": csvTableDef{
			"nTopRareLogs",
			[]string{"rowID", "score", "text"},
			maxPartitions},
	}
	return newCsvDB(baseDir, d)
}

func getClosedItemsDB(baseDir string, maxPartitions int) (*csvDB, error) {
	d := map[string]csvTableDef{
		"closedItemSets": csvTableDef{
			"closedItemSets",
			[]string{"support", "key", "itemSets", "lastLine"},
			maxPartitions},
		"closedItemKeys": csvTableDef{
			"closedItemKeys",
			[]string{"key"},
			maxPartitions},
	}
	return newCsvDB(baseDir, d)
}

func getTextWriterDB(baseDir string, maxPartitions int) (*csvDB, error) {
	d := map[string]csvTableDef{
		"doc": csvTableDef{
			"doc",
			[]string{"text"},
			maxPartitions},
	}
	return newCsvDB(baseDir, d)
}
