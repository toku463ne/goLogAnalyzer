package analyzer

import (
	"reflect"
	"testing"
)

func TestLargeDCIClosed_Run(t *testing.T) {
	bolShowProgress = false
	a, err := newFileAnalyzer("inputs/sample.txt", 0, "")
	if err != nil {
		t.Errorf("%+v", err)
	}
	ldci := newLargeDCIClosed(2, &a.trans, &a.items, true)
	err = ldci.run()
	if err != nil {
		t.Errorf("%+v", err)
	}
	want1 := [][]int{
		{3}, {1}, {0, 1}, {3, 4}, {1, 3}, {0, 1, 3}, {1, 3, 4}, {0, 1, 3, 4},
	}
	//got1 := ldci.closedSets.getSlice()
	//closedsets, supb, ftid, ltid, err := ldci.getSortedClosedSets()
	closedsets, supb, ftid, ltid, err := ldci.getSortedClosedSets()
	if err != nil {
		t.Errorf("%+v", err)
	}
	if !reflect.DeepEqual(closedsets, want1) {
		t.Errorf("closedSets = %v, want %v", closedsets, want1)
	}
	want2 := []int{5, 5, 4, 4, 4, 3, 3, 2}
	if !reflect.DeepEqual(supb, want2) {
		t.Errorf("supports = %v, want %v", supb, want2)
	}
	want3 := []int{1, 0, 0, 1, 1, 2, 1, 2}
	if !reflect.DeepEqual(ftid, want3) {
		t.Errorf("supports = %v, want %v", ftid, want3)
	}
	want4 := []int{5, 5, 5, 4, 5, 5, 4, 4}
	if !reflect.DeepEqual(ltid, want4) {
		t.Errorf("supports = %v, want %v", ltid, want4)
	}
}
