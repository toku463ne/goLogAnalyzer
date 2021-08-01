package analyzer

import "io"

func (r *circuitRows) next() bool {
	if r.pos >= len(r.tableNames) {
		r.err = io.EOF
		r.rows = nil
		return false
	}

	if r.rows == nil {
		completed := true
		if err := r.statusTable.Select1Row(func(v []string) bool {
			return v[r.blockIDIdx] == r.tableNames[r.pos]
		}, []string{"completed"}, &completed); err != nil {
			r.err = err
			return false
		}

		t, err := r.csvdbObj.GetTable(r.tableNames[r.pos])
		if err != nil {
			r.err = err
			return false
		}
		rows, err := t.SelectRows(r.conditionCheckFunc, r.columns)
		if err != nil {
			r.err = err
			return false
		}
		r.rows = rows
		r.blockCompleted = completed

	}

	r.rows.Next()
	err := r.rows.Err()
	r.err = err

	if err != nil && err.Error() == "EOF" {
		r.pos++
		r.rows = nil
		return r.next()
	} else if err != nil {
		r.err = err
		return false
	}
	return true
}

func (r *circuitRows) scan(a ...interface{}) error {
	return r.rows.Scan(a...)
}
