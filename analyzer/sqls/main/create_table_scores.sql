CREATE TABLE IF NOT EXISTS scores (
    seqNo INTEGER PRIMARY KEY,
    blockNo INTEGER UNIQUE,
    rowCount INTEGER,
    scoreSum REAL,
    scoreSqrSum REAL,
    completed INTEGER
);