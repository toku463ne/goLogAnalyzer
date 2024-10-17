package logan

var (
	tableDefs = map[string][]string{
		"config": {"logPath", "blockSize", "maxBlocks",
			"keepPeriod", "keepUnit",
			"termCountBorderRate", "termCountBorder",
			"timestampLayout", "logFormat"},
		"lastStatus":       {"lastRowId", "lastFileEpoch", "lastFileRow"},
		"logGroups":        {"groupId", "count", "created", "updated"},
		"logGroupsDetails": {"groupId", "line"},
	}
)
