package analyzer

import (
	"fmt"
	"testing"
)

func TestTextWriter_run1(t *testing.T) {
	assertSavedCount := func(tw *textWriter, cnt int) bool {
		scnt, err := tw.getSavedCount("*")
		if err != nil {
			t.Errorf("%v", err)
			return false
		}
		if scnt != cnt {
			t.Errorf("want=%d got=%d", cnt, scnt)
			return false
		}
		return true
	}

	testName := "TestTextWriter_run1"
	testDir, err := ensureTestDir(testName)
	if err != nil {
		t.Errorf("%v", err)
		return
	}
	filepath := fmt.Sprintf("%s/sample5.log", testDir)
	if _, err := copyFile("inputs/sample5.log", filepath); err != nil {
		t.Errorf("%v", err)
		return
	}
	r, err := newReader(filepath)
	if err != nil {
		t.Errorf("%v", err)
		return
	}
	defer r.close()

	tw, err := newTextWriter(testDir, 3, 5)
	if err != nil {
		t.Errorf("%v", err)
		return
	}
	id := 0

	tw.db.dropAllTables()

	cnt := 0
	for r.next() {
		if cnt%5 == 0 {
			tw.setID(id)
			id++
			if id >= 3 {
				id = 0
			}
		}

		cnt++
		te := r.text()
		if err := tw.insert(te); err != nil {
			t.Errorf("%v", err)
			return
		}

		if cnt == 3 {
			if !assertSavedCount(tw, 0) {
				return
			}
		}
		if cnt == 5 {
			if !assertSavedCount(tw, 5) {
				return
			}
		}
		if cnt == 15 {
			if !assertSavedCount(tw, 15) {
				return
			}
		}
		if cnt == 18 {
			if !assertSavedCount(tw, 15) {
				return
			}
		}
		if cnt == 20 {
			if !assertSavedCount(tw, 20) {
				return
			}
		}
	}
}
