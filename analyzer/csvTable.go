package analyzer

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

type csvTableDef struct {
	name          string
	columns       []string
	maxPartitions int
}

type csvTable struct {
	name          string
	colMap        map[string]int
	baseDir       string
	maxPartitions int
}

type csvCursor struct {
	filenames          []string
	currReadingFileIdx int
	currReader         *csvReader
	err                error
}

func newCsvTable(name string, columns []string,
	baseDir string, maxPartitions int) *csvTable {

	t := new(csvTable)
	t.name = name
	colMap := map[string]int{}
	for i, col := range columns {
		colMap[col] = i
	}
	t.colMap = colMap
	t.maxPartitions = maxPartitions

	t.baseDir = baseDir
	return t
}

func (t *csvTable) getPath(partitionID string) string {
	path := ""
	filename := ""
	if partitionID == "*" {
		filename = fmt.Sprintf("%s_*.csv", t.name)
	} else if partitionID != "" {
		filename = fmt.Sprintf("%s_%0"+fmt.Sprint(maxBlockDitigs)+"s.csv", t.name, partitionID)
	} else {
		filename = fmt.Sprintf("%s.csv", t.name)
	}
	path = fmt.Sprintf("%s/%s", t.baseDir, filename)
	return path
}

func (t *csvTable) getPartitionID(path string) string {
	tokens := strings.Split(path, "_")
	last := tokens[len(tokens)-1]
	tokens = strings.Split(last, ".")
	nstr := tokens[0]
	return nstr
}

func (t *csvTable) getPartitionIDs() []string {
	cur := t.openCur("*")
	partitionIDs := []string{}
	for _, filename := range cur.filenames {
		partitionIDs = append(partitionIDs, t.getPartitionID(filename))
	}
	return partitionIDs
}

func (t *csvTable) openW(partitionID string) (*csvWriter, error) {
	path := t.getPath(partitionID)
	writer, err := newCsvWriter(path)
	if err != nil {
		return nil, err
	}
	return writer, nil
}

func (t *csvTable) openCur(partitionID string) *csvCursor {
	if t.maxPartitions == 0 {
		partitionID = ""
	}
	path := t.getPath(partitionID)
	_, filenames := getSortedGlob(path)
	cur := new(csvCursor)
	cur.filenames = filenames
	cur.currReadingFileIdx = -1
	return cur
	//return nil
}

func (cur *csvCursor) values() []string {
	if cur.currReader != nil {
		return cur.currReader.values
	}
	return nil
}

func (cur *csvCursor) next() bool {
	if cur == nil {
		return false
	}
	if cur.currReader != nil {
		ret := cur.currReader.next()
		cur.err = cur.currReader.err
		return ret
	}
	if cur.filenames == nil {
		cur.err = nil
		return false
	}
	cur.currReadingFileIdx++
	if cur.currReadingFileIdx >= len(cur.filenames) {
		cur.err = nil
		return false
	}

	reader, err := newCsvReader(cur.filenames[cur.currReadingFileIdx])
	if err != nil {
		cur.err = err
		return false
	}

	ret := reader.next()
	cur.err = reader.err
	cur.currReader = reader
	return ret
}

func (t *csvTable) insertRows(rows [][]string, partitionID string) error {
	writer, err := t.openW(partitionID)
	if err != nil {
		return err
	}
	defer writer.close()
	for _, row := range rows {
		if err := writer.write(row); err != nil {
			return err
		}
	}
	writer.flush()
	//writer.close()
	return nil
}

func (t *csvTable) drop(partitionID string) error {
	cur := t.openCur(partitionID)
	filenames := cur.filenames
	for _, filename := range filenames {
		if _, err := os.Stat(filename); err == nil {
			if err := os.Remove(filename); err != nil {
				return err
			}
		}
	}
	return nil
}

func (t *csvTable) dropPartition(partitionID string) error {
	if t.maxPartitions == 0 {
		partitionID = ""
	}
	path := t.getPath(partitionID)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil
	}
	err := os.Remove(path)
	return err
}

func (t *csvTable) minmax(fieldname string,
	conds map[string]string,
	partitionID string) (float64, float64, []string, []string, error) {
	rows, err := t.query(conds, partitionID)
	if err != nil {
		return 0.0, 0.0, nil, nil, err
	}
	if rows == nil {
		return 0.0, 0.0, nil, nil, nil
	}
	maxVal := float64(0.0)
	minVal := float64(0.0)
	var maxRow []string
	var minRow []string
	fIdx := t.colMap[fieldname]
	for i, row := range rows {
		vstr := row[fIdx]
		v, err := strconv.ParseFloat(vstr, 64)
		if err != nil {
			continue
		}
		if i == 0 {
			maxVal = v
			minVal = v
			maxRow = row
			minRow = row
		} else {
			if v > maxVal {
				maxVal = v
				maxRow = row
			}
			if v < minVal {
				minVal = v
				minRow = row
			}
		}
	}
	return minVal, maxVal, minRow, maxRow, nil
}

func (t *csvTable) min(fieldname string,
	conds map[string]string,
	partitionID string) (float64, []string, error) {
	mi, _, miR, _, err := t.minmax(fieldname, conds, partitionID)
	return mi, miR, err
}

func (t *csvTable) max(fieldname string,
	conds map[string]string,
	partitionID string) (float64, []string, error) {
	_, ma, _, maR, err := t.minmax(fieldname, conds, partitionID)
	return ma, maR, err
}

func (t *csvTable) count(conds map[string]string, partitionID string) (int, error) {
	rows, err := t.query(conds, partitionID)
	if err != nil {
		return -1, err
	}
	if rows == nil {
		return 0, nil
	}
	cnt := 0
	for _ = range rows {
		cnt++
	}
	return cnt, nil
}

func (t *csvTable) select1rec(conds map[string]string, partitionID string) ([]string, error) {
	rows, err := t.query(conds, partitionID)
	if err != nil {
		return nil, err
	}
	if rows == nil {
		return nil, nil
	}
	for _, v := range rows {
		return v, nil
	}
	return nil, nil
}

func (t *csvTable) query(conds map[string]string, partitionID string) ([][]string, error) {
	cur := t.openCur(partitionID)
	filenames := cur.filenames
	found := [][]string{}
	defer cur.close()
	for _, filename := range filenames {
		reader, err := newCsvReader(filename)
		if err != nil {
			return nil, err
		}
		defer reader.close()
		for reader.next() {
			v := reader.values
			isCondOk := true
			for col, cond := range conds {
				if v[t.colMap[col]] != cond {
					isCondOk = false
					break
				}
			}
			if isCondOk {
				found = append(found, v)
			}
		}
	}
	return found, nil
}

func (t *csvTable) update(
	conds map[string]string,
	updates map[string]string,
	partitionID string,
	isUpsert bool,
) error {
	if isUpsert && partitionID == "*" {
		return errors.Errorf("partitionID cannot be '*' when isUpsert=true")
	}

	cur := t.openCur(partitionID)
	defer cur.close()
	isUpdated := false
	var err error
	for _, filename := range cur.filenames {
		newFilename := filename + ".tmp"
		isUpdated, err = t.updatePartition(filename, newFilename,
			conds, updates)
		if err != nil {
			return err
		}
		if isUpdated {
			if err := os.Rename(newFilename, filename); err != nil {
				return errors.WithStack(err)
			}
		}
	}
	if isUpdated == false && isUpsert {
		v := make([]string, len(t.colMap))
		for col, val := range conds {
			v[t.colMap[col]] = val
		}
		for col, val := range updates {
			v[t.colMap[col]] = val
		}
		t.insertRows([][]string{v}, partitionID)
	}
	return nil
}

func (t *csvTable) updatePartition(rpath string, wpath string,
	conds map[string]string,
	updates map[string]string) (bool, error) {

	reader, err := newCsvReader(rpath)
	if err != nil {
		return false, err
	}
	defer reader.close()
	wvalues := [][]string{}
	isUpdated := false
	for reader.next() {
		v := reader.values
		isCondOk := true
		if conds == nil {
			isCondOk = true
		} else {
			for col, cond := range conds {
				if v[t.colMap[col]] != cond {
					isCondOk = false
					break
				}
			}
		}
		if isCondOk {
			isUpdated = true
			for col, updv := range updates {
				v[t.colMap[col]] = updv
			}

		}
		if updates != nil || isCondOk == false {
			wvalues = append(wvalues, v)
		}
	}
	if isUpdated {
		writer, err := newCsvWriter(wpath)
		if err != nil {
			return false, err
		}
		defer writer.close()

		for _, v := range wvalues {
			if err := writer.write(v); err != nil {
				return isUpdated, err
			}
		}
		writer.flush()
	}
	return isUpdated, nil
}

func (t *csvTable) delete(
	conds map[string]string,
	partitionID string,
) error {
	return t.update(conds, nil, partitionID, false)
}

func (t *csvTable) dropAll() error {
	parts := []string{"*", ""}
	for _, p := range parts {
		cur := t.openCur(p)
		for _, filename := range cur.filenames {
			if err := os.Remove(filename); err != nil {
				return errors.WithStack(err)
			}
		}
	}
	return nil
}

func (cur *csvCursor) close() {
	if cur == nil {
		return
	}
	if cur.currReader != nil {
		cur.currReader.close()
	}
	cur.currReader = nil
	cur.currReadingFileIdx = -1
	cur.err = nil
}
