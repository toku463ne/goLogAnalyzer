package analyzer

import (
	"database/sql"
	"fmt"
	"strings"
)

type cursor struct {
	rows     *sql.Rows
	cols     []string
	v        []sql.RawBytes
	scanArgs []interface{}
	e        error
}

func cond2WhereStr(conditions []string) string {
	if conditions == nil || len(conditions) == 0 {
		return ""
	}
	whereStr := fmt.Sprintf("where %s", strings.Join(conditions, " and "))
	return whereStr
}

func (cur *cursor) next() bool {
	return cur.rows.Next()
}

func (cur *cursor) values() []string {
	if err := cur.rows.Scan(cur.scanArgs...); err != nil {
		cur.e = err
		return nil
	}

	v := make([]string, len(cur.cols))
	for i, col := range cur.v {
		v[i] = string(col)
	}
	return v
}

func (cur *cursor) close() {
	if cur.rows != nil {
		cur.rows.Close()
	}
	cur.rows = nil
}

func (cur *cursor) err() error {
	return cur.e
}
