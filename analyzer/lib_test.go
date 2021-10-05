package analyzer

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/pkg/errors"
)

func Test_quickSort(t *testing.T) {
	tests := []struct {
		name   string
		input  []int64
		inputs []string
		want   []int64
		wants  []string
	}{
		// TODO: Add test cases.
		{"test1", []int64{10, 15, 3, 9},
			[]string{"10", "15", "3", "9"},
			[]int64{3, 9, 10, 15},
			[]string{"3", "9", "10", "15"}},
		{"test2", []int64{10, 15, 10, 9},
			[]string{"10", "15", "11", "9"},
			[]int64{9, 10, 10, 15},
			[]string{"9", "10", "11", "15"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			quickSort(tt.input, tt.inputs, 0, len(tt.input)-1)
			if !reflect.DeepEqual(tt.want, tt.input) {
				t.Errorf("array: not match. want=%+v got=%+v", tt.want, tt.input)
			}
			if !reflect.DeepEqual(tt.wants, tt.inputs) {
				t.Errorf("arrays: not match. want=%+v got=%+v", tt.wants, tt.inputs)
			}
		})
	}
}

func Test_searchReg(t *testing.T) {
	tests := []struct {
		name string
		s    string
		re   string
		want bool
	}{
		// TODO: Add test cases.
		{"test4", "abcdevg", "(cdg|cdf)", false},
		{"test1", "abcdevg", "cde", true},
		{"test2", "abcdevg", "cdf", false},
		{"test3", "abcdevg", "(cde|cdf)", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := searchReg(tt.s, tt.re)
			if tt.want != got {
				t.Errorf("not match. s=%s re=%s", tt.s, tt.re)
			}
		})
	}
}

func Test_registerNTopRareRec(t *testing.T) {
	ntr := make([]*colLogRecords, 3)
	m := 0.0
	for i := 0; i < 3; i++ {
		rowID := int64(i)
		score := float64(i) * 2.0
		text := fmt.Sprintf("%03d", i)
		ntr, m = registerNTopRareRec(ntr, m, rowID, score, text)
	}

	if ntr[0].rowid != 2 {
		t.Errorf("rowID does not match!")
		return
	}
	if ntr[2].rowid != 0 {
		t.Errorf("rowID does not match!")
		return
	}

	for i := 3; i < 5; i++ {
		rowID := int64(i)
		score := float64(i) * 2.0
		text := fmt.Sprintf("%03d", i)
		ntr, m = registerNTopRareRec(ntr, m, rowID, score, text)
	}

	if ntr[0].rowid != 4 {
		t.Errorf("rowID does not match!")
		return
	}
	if ntr[2].rowid != 2 {
		t.Errorf("rowID does not match!")
		return
	}

}

func Test_getBottoms(t *testing.T) {
	n := []int{0, 0, 1, 62217, 2608, 9, 3, 0, 0, 7386, 9720, 13, 3, 0, 0, 0}

	b := getBottoms(n, 0)
	if !reflect.DeepEqual(b, []int{12, 6}) {
		t.Errorf("does not match!")
		return
	}

	n = []int{0, 0, 0, 0, 15790, 265776, 3362, 32, 3, 72189, 47250, 2077, 26, 3, 0, 0, 0}
	b = getBottoms(n, 0)
	if !reflect.DeepEqual(b, []int{13, 8}) {
		t.Errorf("does not match!")
		return
	}

	n = []int{0, 0, 0, 0, 6830779, 2981682, 106500, 860, 84, 28, 3458406,
		595297, 595297, 0, 315283, 242247, 12579, 51, 0, 0, 0}
	b = getBottoms(n, 0)
	if !reflect.DeepEqual(b, []int{17, 9}) {
		t.Errorf("does not match!")
		return
	}

}

func Test_calcNAvgScore(t *testing.T) {
	assertScore := func(title string, scores []float64, scoreStyle int, want float64) error {
		got := calcNAvgScore(scores, scoreStyle)
		if got != want {
			return errors.New(fmt.Sprintf("%s got=%f want=%f", title, got, want))
		}
		return nil
	}

	//calcNAvgScore(scores []float64, scoreStyle int)
	scores := []float64{0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0}
	if err := assertScore("10 scores", scores, cScoreNDistAvg, 0.0); err != nil {
		t.Errorf("%+v", err)
		return
	}

	scores = []float64{0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0,
		1.0, 1.0, 1.0, 1.0, 1.0, 1.0, 1.0, 1.0, 1.0, 1.0}
	if err := assertScore("20 scores", scores, cScoreNDistAvg, 0.5); err != nil {
		t.Errorf("%+v", err)
		return
	}

	scores = []float64{0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0,
		1.0, 1.0, 1.0, 1.0, 1.0, 1.0, 1.0, 1.0, 1.0, 1.0,
		10.0, 10.0, 10.0, 10.0, 10.0, 10.0, 10.0, 10.0, 10.0, 10.0}
	if err := assertScore("30 scores", scores, cScoreNDistAvg, 5.2); err != nil {
		t.Errorf("%+v", err)
		return
	}

	scores = []float64{10.0, 10.0, 10.0, 10.0, 1.0}
	if err := assertScore("4 scores", scores, cScoreNDistAvg, 2.425); err != nil {
		t.Errorf("%+v", err)
		return
	}
}
