package analyzer

import (
	"fmt"
	"testing"
	"time"
)

func TestFilePointer_run1(t *testing.T) {
	testName := "TestFileRarityAnalyzer_run1"
	testDir, err := ensureTestDir(testName)
	if err != nil {
		t.Errorf("%v", err)
		return
	}
	if _, err := copyFile("inputs/sample1.log.1.gz",
		fmt.Sprintf("%s/sample1.log.1.gz", testDir)); err != nil {
		t.Errorf("%v", err)
		return
	}
	time.Sleep(time.Second * 2)
	if _, err := copyFile("inputs/sample1.log",
		fmt.Sprintf("%s/sample1.log", testDir)); err != nil {
		t.Errorf("%v", err)
		return
	}
	logPathRegex := fmt.Sprintf("%s/sample1.log*", testDir)

	fp := newFilePointer(logPathRegex, 0, 0)
	if err := fp.open(); err != nil {
		t.Errorf("%v", err)
		return
	}
	s := []string{"001", "002", "003", "004", "005",
		"006", "007", "008", "009", "010", "011", "012"}

	i := 0
	for fp.next() {
		te := fp.text()
		if s[i] != te {
			t.Errorf("want=%s got=%s", s[i], te)
			return
		}
		i++
	}
}
