package analyzer

import (
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
