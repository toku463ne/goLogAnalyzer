package csvdb

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/go-ini/ini"
	"github.com/pkg/errors"
)

func newCsvTableGroup(groupName, rootDir string,
	columns []string,
	useGzip bool, bufferSize, readBufferSize int) (*CsvTableGroup, error) {
	g := new(CsvTableGroup)
	g.groupName = groupName
	g.rootDir = rootDir
	g.dataDir = fmt.Sprintf("%s/%s", rootDir, groupName)
	if err := ensureDir(g.dataDir); err != nil {
		return nil, err
	}
	//g.dataDir = dataDir
	//g.iniFile = fmt.Sprintf("%s/%s.%s", g.dataDir, groupName, cTblIniExt)
	g.iniFile = fmt.Sprintf("%s/%s.%s", g.rootDir, groupName, cTblIniExt)
	g.tableDefs = make(map[string]*CsvTableDef)
	g.init(columns, useGzip, bufferSize, readBufferSize)
	return g, nil
}

func (g *CsvTableGroup) getTablePath(tableName string) string {
	path := fmt.Sprintf("%s/%s.csv", g.dataDir, tableName)

	if g.useGzip {
		path += ".gz"
	}
	return path
}

func (g *CsvTableGroup) init(columns []string,
	useGzip bool, bufferSize, readBufferSize int) {

	g.columns = columns
	g.useGzip = useGzip
	g.bufferSize = bufferSize
	g.readBufferSize = readBufferSize

}

func (g *CsvTableGroup) load(iniFile string) error {
	g.iniFile = iniFile
	pos := strings.LastIndex(iniFile, "/")
	if pos == -1 {
		pos = strings.LastIndex(iniFile, "\\")
		if pos == -1 {
			return errors.New("Not a proper path : " + iniFile)
		}
	}

	fileName := iniFile[pos+1:]
	tokens := strings.Split(fileName, ".")
	if len(tokens) != 3 {
		return errors.New("Not a proper filename format : " + iniFile)
	}
	pos = strings.Index(iniFile, cTblIniExt)
	if pos == -1 {
		return errors.New("Not a proper extension : " + iniFile)
	}
	g.dataDir = iniFile[:pos-1]

	g.groupName = tokens[0]

	cfg, err := ini.Load(iniFile)
	if err != nil {
		return errors.WithStack(err)
	}
	tableNames := make([]string, 0)
	columns := make([]string, 0)
	useGzip := false
	bufferSize := cDefaultBuffSize
	readBufferSize := 0
	for _, k := range cfg.Section("conf").Keys() {
		switch k.Name() {
		case "tableNames":
			tableNameStr := k.MustString("")
			if tableNameStr == "" {
				return errors.New("Not available ini file")
			}
			tableNames = strings.Split(tableNameStr, ",")
		case "columns":
			columns = strings.Split(k.MustString(""), ",")
		case "useGzip":
			useGzip = k.MustBool(false)
		case "bufferSize":
			bufferSize = k.MustInt(cDefaultBuffSize)
		case "readBufferSize":
			readBufferSize = k.MustInt(0)
		}
	}

	g.init(columns, useGzip, bufferSize, readBufferSize)
	tableDefs := make(map[string]*CsvTableDef, len(tableNames))
	for _, tableName := range tableNames {
		tableDefs[tableName] = newCsvTableDef(g.groupName,
			tableName, g.getTablePath(tableName))
	}
	g.tableDefs = tableDefs

	return nil
}

func (g *CsvTableGroup) save() error {
	if len(g.tableDefs) == 0 {
		return nil
	}

	tableNames := make([]string, len(g.tableDefs))
	i := 0
	for tableName, _ := range g.tableDefs {
		tableNames[i] = tableName
		i++
	}

	file, err := os.OpenFile(g.iniFile, os.O_CREATE, 0640)
	if err != nil {
		return errors.WithStack(err)
	}
	defer file.Close()

	cfg, err := ini.Load(g.iniFile)
	if err != nil {
		return errors.WithStack(err)
	}
	cfg.Section("conf").Key("groupName").SetValue(g.groupName)
	cfg.Section("conf").Key("columns").SetValue(strings.Join(g.columns, ","))
	cfg.Section("conf").Key("tableNames").SetValue(strings.Join(tableNames, ","))
	cfg.Section("conf").Key("useGzip").SetValue(strconv.FormatBool(g.useGzip))
	cfg.Section("conf").Key("bufferSize").SetValue(strconv.Itoa(g.bufferSize))
	cfg.Section("conf").Key("readBufferSize").SetValue(strconv.Itoa(g.readBufferSize))

	if err := cfg.SaveTo(g.iniFile); err != nil {
		return errors.WithStack(err)
	}

	if _, err := os.Stat(g.dataDir); os.IsNotExist(err) {
		if err := os.MkdirAll(g.dataDir, 0755); err != nil {
			return errors.WithStack(err)
		}
	}
	return nil
}

func (g *CsvTableGroup) DropTable(tableName string) error {
	t, err := g.GetTable(tableName)
	if err != nil {
		return err
	}
	if t == nil {
		return nil
	}
	return t.Drop()
}

func (g *CsvTableGroup) Drop() error {
	if pathExist(g.dataDir) {
		if err := os.RemoveAll(g.dataDir); err != nil {
			return errors.WithStack(err)
		}
	}
	if pathExist(g.iniFile) {
		if err := os.Remove(g.iniFile); err != nil {
			return errors.WithStack(err)
		}
	}
	return nil
}

func (g *CsvTableGroup) TableExists(tableName string) bool {
	if g == nil || g.iniFile == "" {
		return false
	}
	if !pathExist(g.getTablePath(tableName)) {
		return false
	}
	if !pathExist(g.iniFile) {
		return false
	}
	_, ok := g.tableDefs[tableName]
	return ok
}

func (g *CsvTableGroup) GetTable(tableName string) (*CsvTable, error) {
	if err := ensureDir(g.dataDir); err != nil {
		return nil, err
	}
	if _, ok := g.tableDefs[tableName]; ok {
		path := g.getTablePath(tableName)
		return newCsvTable(g.groupName, tableName,
			path, g.columns, g.useGzip, g.bufferSize, g.readBufferSize)
	} else {
		return g.CreateTable(tableName)
	}
}

func (g *CsvTableGroup) CreateTable(tableName string) (*CsvTable, error) {

	if g.TableExists(tableName) {
		return nil, errors.New(fmt.Sprintf("The table %s exists", tableName))
	}
	t, err := newCsvTable(g.groupName, tableName, g.getTablePath(tableName),
		g.columns, g.useGzip, g.bufferSize, g.readBufferSize)
	if err != nil {
		return nil, err
	}

	g.tableDefs[tableName] = t.CsvTableDef
	if err := t.Flush(); err != nil {
		return nil, err
	}
	if err := g.save(); err != nil {
		return nil, err
	}
	return t, nil
}

func (g *CsvTableGroup) CreateTableIfNotExists(tableName string) (*CsvTable, error) {
	if g.TableExists(tableName) {
		t, err := g.GetTable(tableName)
		if err != nil {
			return nil, err
		}
		return t, err
	}
	return g.CreateTable(tableName)
}

func (g *CsvTableGroup) Count(conditionCheckFunc func([]string) bool) int {
	cnt := 0
	for tableName := range g.tableDefs {
		tb, err := g.GetTable(tableName)
		if err != nil {
			return -1
		}
		cnt += tb.Count(conditionCheckFunc)
	}
	return cnt
}
