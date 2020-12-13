package analyzer

import "fmt"

type textWriter struct {
	id           int
	db           *csvDB
	count        int
	rows         [][]string
	pos          int
	rowsInBuffer int
}

func newTextWriter(rootDir string, numPartitions, rowsInBuffer int) (*textWriter, error) {
	t := new(textWriter)
	t.id = -1
	dbDir := fmt.Sprintf("%s/textwDB", rootDir)
	db, err := getTextWriterDB(dbDir, numPartitions)
	if err != nil {
		return nil, err
	}
	t.db = db
	t.count = 0
	t.rows = make([][]string, rowsInBuffer)
	t.pos = 0
	t.rowsInBuffer = rowsInBuffer
	return t, nil
}

func (t *textWriter) setID(id int) {
	t.id = id
}

func (t *textWriter) insert2Buffer(text []string) error {
	//t.rows[t.pos] = []string{text}
	t.rows[t.pos] = text
	t.pos++
	if t.pos >= len(t.rows) {
		//if err := t.db.tables["doc"].insertRows(t.rows, fmt.Sprint(t.id)); err != nil {
		//	return err
		//}
		rows := make([][]string, t.rowsInBuffer)
		//t.pos = 0
		t.rows = append(t.rows, rows...)
	}
	return nil
}

func (t *textWriter) flush() ([][]string, error) {
	if t.pos == 0 && t.rows[0] == nil {
		return nil, nil
	}
	rows := make([][]string, t.pos)
	for i := range rows {
		rows[i] = t.rows[i]
	}
	if t.id == -1 {
		t.id = 0
	}
	err := t.db.tables["doc"].insertRows(rows, fmt.Sprint(t.id))
	t.pos = 0
	t.rows = make([][]string, rowsToInsertAtOnce)

	return rows, err
}

func (t *textWriter) destroy() error {
	return t.db.dropAllTables()
}
