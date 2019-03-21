package analyzer

import (
	"reflect"
	"testing"
)

func Test_tran2BitMatrix(t *testing.T) {
	a, err := newFileAnalyzer("inputs/sample.txt", 0, "", "")
	if err != nil {
		t.Errorf("%+v", err)
	}
	/*[
	bitarray('101011'),
	bitarray('111011'),
	bitarray('100000'),
	bitarray('011111'),
	bitarray('011110')]
	*/
	want := [][]int{
		{1, 0, 1, 0, 1, 1},
		{1, 1, 1, 0, 1, 1},
		{1, 0, 0, 0, 0, 0},
		{0, 1, 1, 1, 1, 1},
		{0, 1, 1, 1, 1, 0},
	}
	if got := tran2BitMatrix(a.trans, a.items); !reflect.DeepEqual(got.toArrays(), want) {
		t.Errorf("tran2BitMatrix() = %v, want %v", got.toArrays(), want)
	}

}
