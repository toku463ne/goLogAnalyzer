package analyzer

import (
	"fmt"
	"reflect"
	"testing"
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
	ntr := make([]*logRec, 3)
	m := 0.0
	for i := 0; i < 3; i++ {
		rowID := int64(i)
		score := float64(i) * 2.0
		text := fmt.Sprintf("%03d", i)
		ntr, m = registerNTopRareRec(ntr, m, rowID, score, text)
	}

	if ntr[0].rowID != 2 {
		t.Errorf("rowID does not match!")
		return
	}
	if ntr[2].rowID != 0 {
		t.Errorf("rowID does not match!")
		return
	}

	for i := 3; i < 5; i++ {
		rowID := int64(i)
		score := float64(i) * 2.0
		text := fmt.Sprintf("%03d", i)
		ntr, m = registerNTopRareRec(ntr, m, rowID, score, text)
	}

	if ntr[0].rowID != 4 {
		t.Errorf("rowID does not match!")
		return
	}
	if ntr[2].rowID != 2 {
		t.Errorf("rowID does not match!")
		return
	}

}
