package analyzer

var dbTables = []string{"logs", "logBlocks", "items"}
var dbTablesDefs = map[string]string{
	"test": `CREATE TABLE IF NOT EXISTS test (
		blockID int auto_increment, 
		score float,
		PRIMARY KEY (blockID)
		);`,
	"logs": `CREATE TABLE IF NOT EXISTS logs (
		logID int auto_increment,
		name varchar(50) unique NOT NULL,
		host varchar(50) NOT NULL,
		pathRegex varchar(100) NOT NULL,
		filterRe varchar(100) DEFAULT '',
		xFilterRe varchar(100) DEFAULT '',
		createdAt datetime DEFAULT CURRENT_TIMESTAMP,
		PRIMARY KEY (logID)
		);`,
	"logBlocks": `CREATE TABLE IF NOT EXISTS logBlocks(
		logID int NOT NULL,
		blockID int NOT NULL, 
		blockCnt int DEFAULT 0,
		lastRow int default 0,
		scoreSum float,
		scoreSqrSum float,
		lastEpoch int default 0,
		lastUtcTime datetime DEFAULT '0000-00-00 00:00:00',
		createdAt datetime DEFAULT CURRENT_TIMESTAMP,
		PRIMARY KEY (logID, blockID),
		FOREIGN KEY fk_logBlocks_logID(logID) 
		REFERENCES logs(logID)
		ON DELETE CASCADE	
	)`,
	"items": `CREATE TABLE IF NOT EXISTS items (
		logID int NOT NULL,
		blockID int NOT NULL, 
		word varchar(100)  NOT NULL,
		cnt int DEFAULT 0,
		PRIMARY KEY (logId, blockID, word),
		FOREIGN KEY fk_items_blockID(logID, blockID) 
		REFERENCES logBlocks(logID, blockID)
		ON DELETE CASCADE
		);`,
}
