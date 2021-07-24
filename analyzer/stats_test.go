package analyzer

import (
	"fmt"
	"testing"
)

func Test_stats(t *testing.T) {
	assertScore := func(st *stats, expCnt int64, expSum, expSqrSum float64) error {
		if err := getGotExpErr("scoreCount", st.scoreCount, expCnt); err != nil {
			return err
		}
		if err := getGotExpErr("scoreSum", st.scoreSum, expSum); err != nil {
			return err
		}
		if err := getGotExpErr("scoreSqrSum", st.scoreSqrSum, expSqrSum); err != nil {
			return err
		}
		return nil
	}

	assertCurrBlockScore := func(st *stats, expCnt int64, expSum, expSqrSum float64) error {
		if err := getGotExpErr("scoreCount", st.currBlock.scoreCount, expCnt); err != nil {
			return err
		}
		if err := getGotExpErr("scoreSum", st.currBlock.scoreSum, expSum); err != nil {
			return err
		}
		if err := getGotExpErr("scoreSqrSum", st.currBlock.scoreSqrSum, expSqrSum); err != nil {
			return err
		}
		return nil
	}

	checkScoresTable := func(st *stats, blockNo int,
		expected ...interface{}) error {
		var rowCount int
		var scoreSum float64
		var scoreSqrSum float64
		var completed bool
		if err := st.select1rec(fmt.Sprintf(`SELECT 
rowCount, scoreSum, scoreSqrSum, completed 
FROM scores WHERE blockNo = %d 
AND seqNo = (SELECT MAX(seqNo) FROM scores WHERE blockNo = %d);`, blockNo, blockNo),
			&rowCount, &scoreSum, &scoreSqrSum, &completed); err != nil {
			return err
		}
		if err := getGotExpErr("rowCount", rowCount, expected[0]); err != nil {
			return err
		}
		if err := getGotExpErr("scoreSum", scoreSum, expected[1]); err != nil {
			return err
		}
		if err := getGotExpErr("scoreSqrSum", scoreSqrSum, expected[2]); err != nil {
			return err
		}
		if err := getGotExpErr("completed", completed, expected[3]); err != nil {
			return err
		}
		return nil
	}

	dataDir, err := ensureTestDir("statstest")
	if err != nil {
		t.Errorf("%v", err)
		return
	}
	err = dropDB(dataDir, "main")
	if err != nil {
		t.Errorf("%v", err)
		return
	}

	maxBlocks := 3
	maxRowsInBlock := 5
	st, err := newStats(dataDir, maxBlocks, maxRowsInBlock)
	if err != nil {
		t.Errorf("%v", err)
		return
	}

	for _, score := range []float64{1.0, 1.0, 1.0} {
		st.registerScore(score)
	}

	if err := assertScore(st, 3, 3.0, 3.0); err != nil {
		t.Errorf("%v", err)
		return
	}
	if err := assertCurrBlockScore(st, 3, 3.0, 3.0); err != nil {
		t.Errorf("%v", err)
		return
	}

	for _, score := range []float64{1.0, 1.0, 2.0} {
		if err := st.registerScore(score); err != nil {
			t.Errorf("%v", err)
			return
		}
	}

	if err := assertCurrBlockScore(st, 1, 2.0, 4.0); err != nil {
		t.Errorf("%v", err)
		return
	}
	if err := assertScore(st, 6, 7.0, 9.0); err != nil {
		t.Errorf("%v", err)
		return
	}
	//blockNo, rowCount, scoreSum, scoreSqrSum, completed
	if err := checkScoresTable(st, 0, 5, 5.0, 5.0, true); err != nil {
		t.Errorf("%v", err)
		return
	}

	for _, score := range []float64{2.0, 2.0, 2.0, 2.0,
		3.0, 3.0, 3.0, 3.0, 3.0,
		4.0} {
		if err := st.registerScore(score); err != nil {
			t.Errorf("%v", err)
			return
		}
	}

	if err := assertCurrBlockScore(st, 1, 4.0, 16.0); err != nil {
		t.Errorf("%v", err)
		return
	}
	if err := assertScore(st, 16, 34.0, 86.0); err != nil {
		t.Errorf("%v", err)
		return
	}
	//blockNo, rowCount, scoreSum, scoreSqrSum, completed
	if err := checkScoresTable(st, 0, 5, 5.0, 5.0, true); err != nil {
		t.Errorf("%v", err)
		return
	}
	if err := checkScoresTable(st, 1, 5, 10.0, 20.0, true); err != nil {
		t.Errorf("%v", err)
		return
	}
	if err := checkScoresTable(st, 2, 5, 15.0, 45.0, true); err != nil {
		t.Errorf("%v", err)
		return
	}

	if err := st.commit(false); err != nil {
		t.Errorf("%v", err)
		return
	}

	//blockNo, rowCount, scoreSum, scoreSqrSum, completed
	if err := checkScoresTable(st, 0, 1, 4.0, 16.0, false); err != nil {
		t.Errorf("%v", err)
		return
	}

	st.close()

	st, err = newStats(dataDir, maxBlocks, maxRowsInBlock)
	if err != nil {
		t.Errorf("%v", err)
		return
	}

	if err := assertCurrBlockScore(st, 1, 4.0, 16.0); err != nil {
		t.Errorf("%v", err)
		return
	}
	if err := assertScore(st, 11, 29.0, 81.0); err != nil {
		t.Errorf("%v", err)
		return
	}

	for _, score := range []float64{4.0, 4.0, 4.0, 4.0, 5.0} {
		if err := st.registerScore(score); err != nil {
			t.Errorf("%v", err)
			return
		}
	}

	if err := assertCurrBlockScore(st, 1, 5.0, 25.0); err != nil {
		t.Errorf("%v", err)
		return
	}

	//blockNo, rowCount, scoreSum, scoreSqrSum, completed
	if err := checkScoresTable(st, 0, 5, 20.0, 80.0, true); err != nil {
		t.Errorf("%v", err)
		return
	}

}
