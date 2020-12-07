package analyzer

import (
	"testing"
)

func TestCsvTable_exec(t *testing.T) {
	name := "TestCsvTable_exec"
	rootDir, err := ensureTestDir(name)
	if err != nil {
		t.Errorf("%v", err)
	}

	tb := newCsvTable("test",
		[]string{"id", "name"},
		rootDir, 10)

	if err := tb.dropAll(); err != nil {
		t.Errorf("%v", err)
	}

	if err := tb.insertRows([][]string{
		[]string{"1", "test1"},
		[]string{"2", "test2"},
	}, "001"); err != nil {
		t.Errorf("Error inserting table test : %v", err)
	}
	cur := tb.openCur("001")

	for cur.next() {
		if cur.err != nil {
			t.Errorf("Error getting data : %v", cur.err)
		}
		v := cur.values()
		if v[0] != "1" {
			t.Errorf("data error want=%s got=%s", "1", v[0])
		}
		if v[1] != "test1" {
			t.Errorf("data error want=%s got=%s", "test1", v[1])
		}
		break
	}

	cur.close()

	if err := tb.insertRows([][]string{
		[]string{"3", "test3"},
		[]string{"4", "test4"},
	}, "002"); err != nil {
		t.Errorf("Error inserting table test : %v", err)
	}
	cur = tb.openCur("*")

	for cur.next() {
		if cur.err != nil {
			t.Errorf("Error getting data : %v", cur.err)
		}
		v := cur.values()
		if v[0] != "1" {
			t.Errorf("data error want=%s got=%s", "1", v[0])
		}
		if v[1] != "test1" {
			t.Errorf("data error want=%s got=%s", "test1", v[1])
		}
		break
	}
	cur = tb.openCur("002")
	for cur.next() {
		if cur.err != nil {
			t.Errorf("Error getting data : %v", cur.err)
		}
		v := cur.values()
		if v[0] != "3" {
			t.Errorf("data error want=%s got=%s", "3", v[0])
		}
		if v[1] != "test3" {
			t.Errorf("data error want=%s got=%s", "test3", v[1])
		}
		break
	}

	cur.close()
	err = tb.update(map[string]string{"id": "3"},
		map[string]string{"name": "test3_updated"}, "*", false)
	if err != nil {
		t.Errorf("Error updating : %v", err)
	}

	found, err := tb.query(map[string]string{"id": "3"}, "002")
	if err != nil {
		t.Errorf("Error setting read files : %v", err)
	}
	if found == nil {
		t.Errorf("Found no rows")
	}
	for _, v := range found {
		if v[0] != "3" {
			t.Errorf("data error want=%s got=%s", "3", v[0])
		}
		if v[1] != "test3_updated" {
			t.Errorf("data error want=%s got=%s", "test3_updated", v[1])
		}
		break
	}

	count, err := tb.count(nil, "*")
	if err != nil {
		t.Errorf("%v", err)
	}
	if count != 4 {
		t.Errorf("count error want=%d got=%d", 4, count)
	}
}
