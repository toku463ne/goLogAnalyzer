package analyzer

func getTableDef(dbName, tableName string) {

}

var (
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
		"items": map[string]string{
			"block": `CREATE TABLE IF NOT EXISTS {{ blockName }} (
				item INTEGER,
				itemCount INTEGER
			);`,
		},
		"logRecords": map[string]string{
			"block": `CREATE TABLE IF NOT EXISTS {{ blockName }} (
				rowId INTEGER,
				score REAL,
				record TEXT
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
