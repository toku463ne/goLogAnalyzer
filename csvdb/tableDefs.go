package csvdb

var (
	csvTableInfoDef = `CREATE TABLE IF NOT EXISTS csvTableInfo (
name TEXT,
columns TEXT,
useGzip NUMBER,
bufferSize NUMBER
);`
)
