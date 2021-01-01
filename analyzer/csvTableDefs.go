package analyzer

import "fmt"

func getRarityAnalDB(baseDir string, maxPartitions int) (*csvDB, error) {
	countPerGapH := make([]string, cCountbyGapLen+1)
	countPerGapH[0] = "blockID"
	for i := 0; i < cCountbyGapLen; i++ {
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
			"text",
			[]string{"key"},
			maxPartitions},
	}
	return newCsvDB(baseDir, d)
}
