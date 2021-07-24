package analyzer

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

func newCircuitDB(dataDir, dbName string,
	maxBlocks, maxRowsInBlock int) (*circuitDB, error) {
	rt := new(circuitDB)
	if dataDir != "" {
		d, err := newDB(dataDir, dbName)
		if err != nil {
			return nil, err
		}
		rt.db = d
		if err := rt.createcircuitDBStatus(); err != nil {
			return nil, err
		}
	}
	rt.maxBlocks = maxBlocks
	rt.maxRowsInBlock = maxRowsInBlock
	rt.blockNo = 0
	rt.rowNo = 0
	if err := rt.loadcircuitDBStatus(); err != nil {
		return nil, err
	}
	//rt.rows = make([][]interface{}, maxRowsInBlock)
	//rt.nextBlock()
	return rt, nil
}

func (rt *circuitDB) createcircuitDBStatus() error {
	if _, err := rt.db.execFromFile("create_table_circuitDBStatus.sql"); err != nil {
		return err
	}
	return nil
}

func (rt *circuitDB) loadcircuitDBStatus() error {
	cnt := rt.db.count("circuitDBStatus", "completed = 0")
	if cnt <= 0 {
		return nil
	}

	if err := rt.db.select1rec(`SELECT MAX(lastIndex) 
FROM circuitDBStatus;`, &rt.lastIndex); err != nil {
		return err
	}

	if err := rt.db.select1rec(fmt.Sprintf(`SELECT rowNo FROM circuitDBStatus
WHERE lastIndex =%d;`, rt.lastIndex), &rt.rowNo); err != nil {
		return err
	}

	var completed bool
	if err := rt.db.select1rec(fmt.Sprintf(`SELECT completed 
FROM circuitDBStatus
WHERE lastIndex = %d;`, rt.lastIndex),
		&completed); err != nil {
		return err
	}

	if completed {
		if _, err := rt.db.exec(`UPDATE circuitDBStatus SET completed = 1;`); err != nil {
			return err
		}
	}

	return nil
}

func (rt *circuitDB) getBlockTableName(blockNo int) string {
	return fmt.Sprintf("BLK%s", rt.getBlockID(blockNo))
}

func (rt *circuitDB) execFromFile(sqlFile string, blockNo int) (sql.Result, error) {
	sqlstr, err := rt.getSqlFileContents(sqlFile)
	if err != nil {
		return nil, err
	}
	sqlstr = strings.ReplaceAll(sqlstr, "{{ blockName }}", rt.getBlockTableName(rt.blockNo))
	return rt.exec(sqlstr)
}

func (rt *circuitDB) createBlock(blockNo int) error {
	sqlFile := fmt.Sprintf("%s/create_table.sql", rt.dbName)
	_, err := rt.execFromFile(sqlFile, rt.blockNo)
	return err
}

func (rt *circuitDB) dropBlock(blockNo int) error {
	_, err := rt.db.exec(`DELETE FROM 
circuitDBStatus WHERE blockNo =` + strconv.Itoa(rt.blockNo))
	if err != nil {
		return err
	}

	return rt.dropTable(rt.getBlockTableName(rt.blockNo))
}

func (rt *circuitDB) dropBlockIfCompeted(blockNo int) error {
	cnt := rt.count("circuitDBStatus", fmt.Sprintf("blockNo = %d", blockNo))
	if cnt <= 0 {
		return nil
	}

	var completed bool
	if err := rt.db.select1rec(fmt.Sprintf(`SELECT completed 
FROM circuitDBStatus
WHERE blockNo = %d;`, blockNo),
		&completed); err != nil {
		return err
	}
	if completed {
		if err := rt.dropBlock(rt.blockNo); err != nil {
			return err
		}
	}
	return nil
}

func (rt *circuitDB) updateBlockStatus(completed bool) error {
	_, err := rt.db.exec(`DELETE FROM 
circuitDBStatus WHERE blockNo =` + strconv.Itoa(rt.blockNo))
	if err != nil {
		return err
	}

	stmt, err := rt.db.conn.Prepare(`INSERT INTO 
circuitDBStatus(lastIndex, blockNo, blockID, rowNo, completed) 
 VALUES(?,?,?,?,?)`)
	if err != nil {
		return errors.WithStack(err)
	}
	blockID := rt.getBlockTableName(rt.blockNo)
	_, err = stmt.Exec(rt.lastIndex, rt.blockNo, blockID, rt.rowNo, completed)
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}

func (rt *circuitDB) nextBlock() error {
	rt.rowNo = 0
	rt.blockNo++
	if rt.blockNo >= rt.maxBlocks {
		rt.blockNo = 0
	}
	rt.lastIndex++
	//rt.rows = make([][]interface{}, rt.maxRowsInBlock)
	return nil
}

func (rt *circuitDB) getBlockID(blockNo int) string {
	return fmt.Sprintf("%0"+strconv.Itoa(cMaxBlockDitigs)+"d", blockNo)
}

func (rt *circuitDB) getBlockNos() ([]int, error) {
	cnt := rt.count(`circuitDBStatus`, "")
	blockNos := make([]int, cnt)
	rows, err := rt.db.query(`SELECT BlockNo FROM circuitDBStatus;`)
	if err != nil {
		return nil, err
	}
	i := 0
	for rows.Next() {
		if err := rows.Scan(&blockNos[i]); err != nil {
			return nil, err
		}
	}
	return blockNos, nil
}

func (rt *circuitDB) query(fields []string, conds string) (*circuitRows, error) {
	r := new(circuitRows)
	cnt := rt.count("circuitDBStatus", "")
	if cnt < 0 {
		return nil, nil
	}
	r.tableNames = make([]string, cnt)
	rows, err := rt.db.query(`SELECT blockID FROM circuitDBStatus;`)
	if err != nil {
		return nil, err
	}
	pos := 0
	for rows.Next() {
		name := ""
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		r.tableNames[pos] = name
		pos++
	}
	r.pos = 0
	r.fields = strings.Join(fields, ", ")
	r.conds = conds
	r.db = rt.db
	return r, nil
}

func (rt *circuitDB) countAll(cond string) int {
	cnt := rt.count("circuitDBStatus", "")
	if cnt < 0 {
		return -1
	}
	tableNames := make([]string, cnt)
	rows, err := rt.db.query(`SELECT blockID FROM circuitDBStatus;`)
	if err != nil {
		return -1
	}
	pos := 0
	for rows.Next() {
		name := ""
		if err := rows.Scan(&name); err != nil {
			return -1
		}
		tableNames[pos] = name
		pos++
	}
	cnt = 0
	for _, name := range tableNames {
		cnt += rt.db.count(name, cond)
	}
	return cnt
}

//func (rt *circuitDB) setInsertCols(cols []string) {
//	rt.cols = cols
//}
