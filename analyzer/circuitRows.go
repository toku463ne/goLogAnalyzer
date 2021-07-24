package analyzer

import "fmt"

func (r *circuitRows) next() bool {
	if r.pos >= len(r.tableNames) {
		r.err = nil
		r.rows = nil
		return false
	}

	if r.rows == nil {
		sqlstr := fmt.Sprintf(`SELECT %s FROM %s;`, r.fields, r.tableNames[r.pos])
		if r.conds != "" {
			sqlstr = `WHERE ` + r.conds
		}
		rows, err := r.db.query(sqlstr)
		if err != nil {
			r.err = err
			return false
		}
		completed := true
		err = r.db.select1rec(fmt.Sprintf(`SELECT completed FROM circuitDBStatus
WHERE blockID = '%s'`, r.tableNames[r.pos]), &completed)
		if err != nil {
			r.err = err
			return false
		}
		r.blockCompleted = completed
		r.rows = rows
	}

	res := r.rows.Next()
	err := r.rows.Err()
	if err != nil {
		r.err = err
		return false
	}

	if !res {
		r.pos++
		r.rows = nil
		return r.next()
	}
	return true
}

func (r *circuitRows) scan(a ...interface{}) error {
	return r.rows.Scan(a...)
}
