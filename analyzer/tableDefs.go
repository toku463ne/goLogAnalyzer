package analyzer

var (
	tableDefs = map[string][]string{
		"config": {"rootDir", "logPathRegex", "linesInBlock",
			"maxBlocks", "maxItemBlocks", "filterRe", "xFilterRe", "minGapToRecord"},
		"lastStatus": {"lastRowID", "lastFileEpoch", "lastFileRow"},
		"items":      {"item", "itemCount"},
		"logRecords": {"rowID", "score", "record"},
		"scores":     {"seqNo", "blockNo", "rowCount", "scoreSum", "scoreSqrSum", "completed"},
		"statistics": {"lastRowID", "lastFileEpoch", "lastFileRow"},
		"blockInfo": {"lastIndex", "blockNo", "blockID",
			"rowNo", "lastEpoch", "completed"},
		"circuitDBStatus": {"lastIndex", "blockNo", "blockID", "rowNo", "lastEpoch", "completed"},
	}

	dbDefVar = map[string](map[string]string){
		"testdb": map[string]string{
			"item": `CREATE TABLE IF NOT EXISTS item_{{ blockname }} (
					id INTEGER,
					comment TEXT
					);`,
			"customer": `CREATE TABLE IF NOT EXISTS customer (
				id INTEGER,
				comment TEXT
				);`,
		},
		"main": map[string]string{
			"config": `CREATE TABLE IF NOT EXISTS config (
				rootDir TEXT PRIMARY KEY,
				logPathRegex TEXT,
				linesInBlock INTEGER,
				maxBlocks INTEGER,
				maxItemBlocks INTEGER,
				filterRe TEXT,
				xFilterRe TEXT,
				minGapToRecord FLOAT
			);`,
			"lastStatus": `CREATE TABLE IF NOT EXISTS lastStatus (
				lastRowID INTEGER,
				lastFileEpoch INTEGER,
				lastFileRow INTEGER
			);`,
		},
		"stats": map[string]string{
			"scores": `CREATE TABLE IF NOT EXISTS scores (
				seqNo INTEGER PRIMARY KEY,
				blockNo INTEGER UNIQUE,
				rowCount INTEGER,
				scoreSum REAL,
				scoreSqrSum REAL,
				completed INTEGER
			);`,
			"statistics": `CREATE TABLE IF NOT EXISTS statistics (
				seqNo INTEGER,
				blockNo INTEGER,
				scoreStage INTEGER,
				itemCount INTEGER,
			PRIMARY KEY (seqNo, blockNo, scoreStage)
			);`,
		},
		"default": map[string]string{
			"circuitDBStatus": `CREATE TABLE IF NOT EXISTS circuitDBStatus (
				lastIndex INTEGER PRIMARY KEY,
				blockNo INTEGER UNIQUE,
				blockID TEXT,
				rowNo INTEGER,
				lastEpoch INTEGER,
				completed INTEGER
			);`,
		},
	}
)
