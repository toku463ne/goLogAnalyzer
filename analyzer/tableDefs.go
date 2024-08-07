package analyzer

var (
	tableDefs = map[string][]string{
		"config": {"rootDir", "logPathRegex", "blockSize",
			"maxBlocks", "maxItemBlocks",
			"filterRe", "xFilterRe", "minGapToRecord", "minGapToDetect",
			"datetimeStartPos", "datetimeLayout", "scoreStyle"},
		"lastStatus": {"lastRowID", "lastFileEpoch", "lastFileRow"},
		"items":      {"item", "itemCount", "tranScoreAvg"},
		"phrases":    {"phrase", "phraseCount", "tranScoreAvg"},
		"logRecords": {"rowID", "score", "epoch", "record"},
		"scores": {"seqNo", "blockNo", "rowCount", "scoreSum",
			"scoreSqrSum", "scoreMax", "completed", "lastFileEpoch"},
		"statistics": {"seqNo", "blockNo", "scoreStage", "itemCount", "lastFileEpoch"},
		"blockInfo": {"lastIndex", "blockNo", "blockID",
			"rowNo", "lastEpoch", "completed"},
		"lastTopN":        {"rowid", "score", "maxScore", "count", "lastDate", "record"},
		"circuitDBStatus": {"lastIndex", "blockNo", "blockID", "rowNo", "lastEpoch", "completed"},
	}
)
