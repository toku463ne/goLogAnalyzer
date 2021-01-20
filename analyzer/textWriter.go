package analyzer

import "fmt"

type textWriter struct {
	id       int
	db       *csvDB
	count    int
	rows     [][]string
	pos      int
	buffSize int
}

func newTextWriter(rootDir string, numPartitions, buffSize int) (*textWriter, error) {
	t := new(textWriter)
	t.id = -1
	if buffSize <= 0 {
		t.buffSize = cRowsToInsertAtOnce
	} else {
		t.buffSize = buffSize
	}
	dbDir := t.getDBDir(rootDir)
	db, err := getTextWriterDB(dbDir, numPartitions)
	if err != nil {
		return nil, err
	}
	t.db = db
	t.count = 0
	t.rows = make([][]string, t.buffSize)
	t.pos = -1
	return t, nil
}

func (t *textWriter) getDBDir(rootDir string) string {
	return fmt.Sprintf("%s/textwriter", rootDir)
}

func (t *textWriter) setID(id int) error {
	if t.pos >= 0 {
		if _, err := t.flush(); err != nil {
			return err
		}
	}
	t.id = id
	return nil
}

func (t *textWriter) insert(text string) error {
	t.pos++
	if t.pos >= len(t.rows) {
		if err := t.db.tables["doc"].insertRows(t.rows, fmt.Sprint(t.id), 0); err != nil {
			return err
		}
		t.rows = make([][]string, t.buffSize)
		t.pos = 0
		//t.rows = append(t.rows, rows...)
	}
	t.rows[t.pos] = []string{text}

	return nil
}

func (t *textWriter) flush() ([][]string, error) {
	if t.pos < 0 || t.rows == nil {
		return nil, nil
	}
	rows := make([][]string, t.pos+1)
	for i := range rows {
		rows[i] = t.rows[i]
	}
	if t.id == -1 {
		t.id = 0
	}
	err := t.db.tables["doc"].insertRows(rows, fmt.Sprint(t.id), 0)
	t.pos = -1
	t.rows = make([][]string, t.buffSize)

	return rows, err
}

func (t *textWriter) dropPartition(idstr string) error {
	return t.db.tables["doc"].dropPartition(idstr)
}

func (t *textWriter) destroy() error {
	return t.db.dropAllTables()
}

func (t *textWriter) getSavedCount(idstr string) (int, error) {
	return t.db.tables["doc"].count(nil, idstr)
}
