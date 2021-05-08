package analyzer

import "fmt"

func getRarityAnalDB(baseDir string, maxPartitions int) (*csvDB, error) {
	countPerScoreH := make([]string, cCountbyScoreLen+1)
	//countPerGapH[0] = "blockID"
	for i := 0; i < cCountbyScoreLen; i++ {
		countPerScoreH[i] = fmt.Sprint(i - 1)
	}

	d := map[string]csvTableDef{
		"lastStatus": {
			"lastStatus",
			[]string{"lastRowID", "lastBlockID", "fileName", "lastRow",
				"modifiedEpoch", "modifiedUtcTime"},
			0},
		"logBlocks": {
			"logBlocks",
			[]string{"blockID", "lastRowID", "blockCnt", "scoreSum",
				"scoreSqrSum", "lastEpoch", "createdAt", "completed"},
			maxPartitions},
		"countPerScore": {
			"countPerScore", countPerScoreH,
			maxPartitions},
		"items": {
			"items",
			[]string{"word", "cnt"},
			maxPartitions},
		"logRecords": {
			"logRecords",
			[]string{"rowID", "score", "scoreGap", "text"},
			maxPartitions},
		"nTopRareLogs": {
			"nTopRareLogs",
			[]string{"rowID", "score", "text"},
			maxPartitions},
	}
	return newCsvDB(baseDir, d)
}

func getClosedItemsDB(baseDir string, maxPartitions int) (*csvDB, error) {
	d := map[string]csvTableDef{
		"closedItemSets": {
			"closedItemSets",
			[]string{"support", "key", "itemSets", "lastLine"},
			maxPartitions},
		"closedItemKeys": {
			"closedItemKeys",
			[]string{"key"},
			maxPartitions},
	}
	return newCsvDB(baseDir, d)
}

func getTextWriterDB(baseDir string, maxPartitions int) (*csvDB, error) {
	d := map[string]csvTableDef{
		"doc": {
			"doc",
			[]string{"text"},
			maxPartitions},
	}
	return newCsvDB(baseDir, d)
}
