CREATE TABLE IF NOT EXISTS statistics (
    seqNo INTEGER,
    blockNo INTEGER,
    scoreStage INTEGER,
    itemCount INTEGER,
PRIMARY KEY (seqNo, blockNo, scoreStage)
);
