package analyzer

import "fmt"

func getCountPerScoreH() []string {
	countPerScoreH := make([]string, cCountbyScoreLen)
	//countPerGapH[0] = "blockID"
	for i := 0; i < cCountbyScoreLen; i++ {
		countPerScoreH[i] = fmt.Sprint(i - 1)
	}
	return countPerScoreH
}

func getRarityAnalDB(baseDir string, maxPartitions int) (*csvDB, error) {
	countPerScoreH := getCountPerScoreH()

	d := map[string]csvTableDef{
		"lastStatus": {
			"lastStatus",
			[]string{"lastRowID", "lastBlockID", "fileName", "lastRow",
				"modifiedEpoch", "modifiedUtcTime"},
			0, false},
		"logBlocks": {
			"logBlocks",
			[]string{"blockID", "lastRowID", "blockCnt", "scoreSum",
				"scoreSqrSum", "lastEpoch", "createdAt", "completed"},
			maxPartitions, false},
		"countPerScore": {
			"countPerScore", countPerScoreH,
			maxPartitions, false},
		"items": {
			"items",
			[]string{"word", "cnt"},
			maxPartitions, true},
		"logRecords": {
			"logRecords",
			[]string{"rowID", "score", "text"},
			maxPartitions, true},
		"nTopRareLogs": {
			"nTopRareLogs",
			[]string{"rowID", "score", "text"},
			maxPartitions, true},
	}
	return newCsvDB(baseDir, d)
}

func getClosedItemsDB(baseDir string, maxPartitions int) (*csvDB, error) {
	d := map[string]csvTableDef{
		"closedItemSets": {
			"closedItemSets",
			[]string{"support", "key", "itemSets", "lastLine"},
			maxPartitions, false},
		"closedItemKeys": {
			"closedItemKeys",
			[]string{"key"},
			maxPartitions, false},
	}
	return newCsvDB(baseDir, d)
}

func getTextWriterDB(baseDir string, maxPartitions int) (*csvDB, error) {
	d := map[string]csvTableDef{
		"doc": {
			"doc",
			[]string{"text"},
			maxPartitions, false},
	}
	return newCsvDB(baseDir, d)
}
