CREATE TABLE IF NOT EXISTS config (
    rootDir TEXT PRIMARY KEY,
    logPathRegex TEXT,
    linesInBlock INTEGER,
    maxBlocks INTEGER,
    maxItemBlocks INTEGER,
    filterRe TEXT,
    xFilterRe TEXT,
    minGapToRecord FLOAT
);