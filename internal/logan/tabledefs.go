package logan

var (
	tableDefs = map[string][]string{
		"config": {"logPath", "blockSize", "maxBlocks",
			"keepPeriod", "unitSecs",
			"termCountBorderRate", "termCountBorder", "minMatchRate",
			"timestampLayout", "useUtcTime", "ignoreNumbers", "separators", "logFormat"},
		"lastStatus":       {"lastRowId", "lastFileEpoch", "lastFileRow"},
		"logGroups":        {"groupId", "retentionPos", "count", "created", "updated"},
		"logGroupsDetails": {"groupId", "line"},
		"terms":            {"term", "count"},
	}
)
