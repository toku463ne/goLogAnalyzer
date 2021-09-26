package analyzer

var (
	tableDefs = map[string][]string{
		"config": {"rootDir", "logPathRegex", "linesInBlock",
			"maxBlocks", "maxItemBlocks",
			"filterRe", "xFilterRe", "minGapToRecord",
			"datetimeStartPos", "datetimeLayout"},
		"lastStatus": {"lastRowID", "lastFileEpoch", "lastFileRow"},
		"items":      {"item", "itemCount"},
		"logRecords": {"rowID", "score", "record"},
		"scores": {"seqNo", "blockNo", "rowCount", "scoreSum",
			"scoreSqrSum", "scoreMax", "completed", "lastFileEpoch"},
		"scoresHist": {"seqNo", "blockNo", "avg", "std", "max", "lastFileEpoch"},
		"statistics": {"seqNo", "blockNo", "scoreStage", "itemCount", "lastFileEpoch"},
		"blockInfo": {"lastIndex", "blockNo", "blockID",
			"rowNo", "lastEpoch", "completed"},
		"circuitDBStatus": {"lastIndex", "blockNo", "blockID", "rowNo", "lastEpoch", "completed"},
	}
)
