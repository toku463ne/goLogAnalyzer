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
		if err := rt.createCircuitDBStatus(); err != nil {
			return nil, err
		}
	}
	rt.maxBlocks = maxBlocks
	rt.maxRowsInBlock = maxRowsInBlock
	rt.blockNo = 0
	rt.rowNo = 0
	if err := rt.loadCircuitDBStatus(); err != nil {
		return nil, err
	}
	//rt.rows = make([][]interface{}, maxRowsInBlock)
	//rt.nextBlock()
	return rt, nil
}

func (rt *circuitDB) createCircuitDBStatus() error {
	if _, err := rt.exec(dbDefVar["default"]["circuitDBStatus"]); err != nil {
		return err
	}
	return nil
}

func (rt *circuitDB) loadCircuitDBStatus() error {
	if cnt := rt.db.count("circuitDBStatus", ""); cnt <= 0 {
		return nil
	}

	var completed bool
	if err := rt.db.select1rec(`SELECT lastIndex, blockNo, rowNo, lastEpoch, completed 
FROM circuitDBStatus
WHERE lastIndex = (SELECT MAX(lastIndex) FROM circuitDBStatus);`,
		&rt.lastIndex, &rt.blockNo, &rt.rowNo, &rt.lastEpoch, &completed); err != nil {
		return err
	}

	if completed {
		if err := rt.nextBlock(); err != nil {
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
	_, err := rt.exec(strings.ReplaceAll(dbDefVar[rt.dbName]["block"],
		"{{ blockName }}", rt.getBlockTableName(rt.blockNo)))
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
circuitDBStatus(lastIndex, blockNo, blockID, rowNo, lastEpoch, completed) 
 VALUES(?,?,?,?,?,?)`)
	if err != nil {
		return errors.WithStack(err)
	}
	blockID := rt.getBlockTableName(rt.blockNo)
	_, err = stmt.Exec(rt.lastIndex, rt.blockNo, blockID, rt.rowNo, rt.lastEpoch, completed)
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}

func (rt *circuitDB) setLastEpoch(lastEpoch int64) {
	rt.lastEpoch = lastEpoch
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
	rows, err := rt.db.query(`SELECT blockNo FROM circuitDBStatus;`)
	if err != nil {
		return nil, err
	}
	i := 0
	for rows.Next() {
		if err := rows.Scan(&blockNos[i]); err != nil {
			return nil, err
		}
		i++
	}
	return blockNos, nil
}

func (rt *circuitDB) selectRows(fields []string,
	conds string, blockNos []int) (*circuitRows, error) {
	var err error
	if blockNos == nil {
		blockNos, err = rt.getBlockNos()
		if err != nil {
			return nil, err
		}
	}

	r := new(circuitRows)
	r.tableNames = make([]string, len(blockNos))
	pos := 0
	for _, blockNo := range blockNos {
		r.tableNames[pos] = rt.getBlockTableName(blockNo)
		pos++
	}
	r.pos = 0
	r.fields = strings.Join(fields, ", ")
	r.conds = conds
	r.db = rt.db
	return r, nil
}

func (rt *circuitDB) selectRows2(fields []string, conds string,
	orderby string, limit int, blockNos []int) (*sql.Rows, error) {
	var err error
	fieldstr := "*"
	if fields != nil {
		fieldstr = strings.Join(fields, ",")
	}
	condstr := ""
	if conds != "" {
		condstr = "WHERE " + conds
	}
	orderbystr := ""
	if orderby != "" {
		orderbystr = "ORDER BY " + orderby
	}
	if blockNos == nil {
		blockNos, err = rt.getBlockNos()
		if err != nil {
			return nil, err
		}
	}
	limitstr := ""
	if limit > 0 {
		limitstr = fmt.Sprintf("limit %d", limit)
	}
	sqlstr := ""
	for _, blockNo := range blockNos {
		if sqlstr != "" {
			sqlstr += "\n UNION \n"
		}
		tableName := rt.getBlockTableName(blockNo)
		sqlstr += fmt.Sprintf(`SELECT %s FROM %s %s`, fieldstr, tableName, condstr)
	}
	sqlstr = fmt.Sprintf("SELECT * FROM \n(%s)\n %s %s", sqlstr, orderbystr, limitstr)
	//print(sqlstr)
	return rt.db.query(sqlstr)
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
