package csvdb

func newCsvTableDef(groupName, tableName, path string) *CsvTableDef {
	td := new(CsvTableDef)
	td.groupName = groupName
	td.tableName = tableName
	td.path = path
	return td
}
