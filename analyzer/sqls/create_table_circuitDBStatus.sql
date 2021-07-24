CREATE TABLE IF NOT EXISTS circuitDBStatus (
    lastIndex INTEGER PRIMARY KEY,
    blockNo INTEGER UNIQUE,
    blockID TEXT,
    rowNo INTEGER,
    completed INTEGER
);