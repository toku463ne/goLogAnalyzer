package analyzer

import (
	"testing"
)

func Test_newArray(t *testing.T) {
	tests := []struct {
		name string
		val1 interface{}
		val2 interface{}
	}{
		// TODO: Add test cases.
		{"int", 1, 2},
		{"[]int", &([]int{1, 2}), &([]int{3, 4})},
		{"different type", 1, "a"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := newArray()
			a.set(0, tt.val1)
			a.set(1, tt.val2)
			if a.get(1) != tt.val2 {
				t.Errorf("unexpected")
			}
			a.set(1999, tt.val1)
			if a.size() != 2000 {
				t.Errorf("unexpected")
			}
		})
	}
}
