package analyzer

import (
	"reflect"
	"testing"
)

func TestDCIClosed_Run(t *testing.T) {
	trans1, items1, err := tokenizeFile("inputs/sample.txt")
	matrix := tran2BitMatrix(trans1, items1)
	dci, err := newDCIClosed(matrix, 2, true)
	if err != nil {
		t.Errorf("newDCIClosed() error = %v", err)
		return
	}
	if err := dci.run(); err != nil {
		t.Errorf("%+v", err)
	}

	want1 := [][]int{
		{0, 1},
		{0, 1, 4, 3},
		{0, 1, 3},
		{4, 3},
		{4, 3, 1},
		{1},
		{1, 3},
		{3},
	}
	got1 := dci.closedSetsToArray()
	if !reflect.DeepEqual(got1, want1) {
		t.Errorf("closedSets = %v, want %v", got1, want1)
	}
	want2 := []int{4, 2, 3, 4, 3, 5, 4, 5}
	got2 := dci.closedSupp.getSlice()
	if !reflect.DeepEqual(got2, want2) {
		t.Errorf("closedSup = %v, want %v", got2, want2)
	}

	want3 := [][]string{{"apple", "melon"},
		{"apple", "melon", "orange", "lemon"},
		{"apple", "melon", "lemon"},
		{"orange", "lemon"},
		{"orange", "lemon", "melon"},
		{"melon"},
		{"melon", "lemon"},
		{"lemon"}}
	want4 := []int{4, 2, 3, 4, 3, 5, 4, 5}
	got3, got4 := dci.getClosedWords(items1)
	if !reflect.DeepEqual(got3, want3) {
		t.Errorf("GetClosedWords = %v, want %v", got3, want3)
	}
	if !reflect.DeepEqual(got4, want4) {
		t.Errorf("GetClosedWords  support = %v, want %v", got4, want4)
	}

	want5 := [][]string{
		{"lemon"},
		{"melon"},
		{"apple", "melon"},
		{"orange", "lemon"},
		{"melon", "lemon"},
		{"apple", "melon", "lemon"},
		{"orange", "lemon", "melon"},
		{"apple", "melon", "orange", "lemon"},
	}
	want6 := []int{5, 5, 4, 4, 4, 3, 3, 2}
	got5, got6, got7, got8 := dci.getClosedWordsSorted(items1)
	if !reflect.DeepEqual(got5, want5) {
		t.Errorf("GetClosedWordsSorted trans got = %v, want %v", got5, want5)
	}
	if !reflect.DeepEqual(got6, want6) {
		t.Errorf("GetClosedWordsSorted support got = %v, want %v", got6, want6)
	}

	want7 := []int{1, 0, 0, 1, 1, 2, 1, 2}
	want8 := []int{5, 5, 5, 4, 5, 5, 4, 4}
	if !reflect.DeepEqual(got7, want7) {
		t.Errorf("GetClosedWordsSorted first appared tid got = %v, want %v", got7, want7)
	}
	if !reflect.DeepEqual(got8, want8) {
		t.Errorf("GetClosedWordsSorted last appared tid got = %v, want %v", got8, want8)
	}
}

/*
func TestDCIClosed_Run(t *testing.T) {
	bolShowProgress = false

	a := newFileAnalyzer(0, "", "")
	err := a.tokenizeFile("inputs/sample.txt")
	if err != nil {
		t.Errorf("%+v", err)
	}
	//a.loadMatrix()
	matrix := tran2BitMatrix(a.trans, a.items)
	dci, err := newDCIClosed(matrix, 2, true)
	if err != nil {
		t.Errorf("newDCIClosed() error = %v", err)
		return
	}
	if err := dci.run(); err != nil {
		t.Errorf("%+v", err)
	}

	want1 := [][]int{
		{0, 1},
		{0, 1, 4, 3},
		{0, 1, 3},
		{4, 3},
		{4, 3, 1},
		{1},
		{1, 3},
		{3},
	}
	got1 := dci.closedSetsToArray()
	if !reflect.DeepEqual(got1, want1) {
		t.Errorf("closedSets = %v, want %v", got1, want1)
	}
	want2 := []int{4, 2, 3, 4, 3, 5, 4, 5}
	got2 := dci.closedSupp.getSlice()
	if !reflect.DeepEqual(got2, want2) {
		t.Errorf("closedSup = %v, want %v", got2, want2)
	}

	want3 := [][]string{{"apple", "melon"},
		{"apple", "melon", "orange", "lemon"},
		{"apple", "melon", "lemon"},
		{"orange", "lemon"},
		{"orange", "lemon", "melon"},
		{"melon"},
		{"melon", "lemon"},
		{"lemon"}}
	want4 := []int{4, 2, 3, 4, 3, 5, 4, 5}
	got3, got4 := dci.getClosedWords(a.items)
	if !reflect.DeepEqual(got3, want3) {
		t.Errorf("GetClosedWords = %v, want %v", got3, want3)
	}
	if !reflect.DeepEqual(got4, want4) {
		t.Errorf("GetClosedWords  support = %v, want %v", got4, want4)
	}

	want5 := [][]string{
		{"lemon"},
		{"melon"},
		{"apple", "melon"},
		{"orange", "lemon"},
		{"melon", "lemon"},
		{"apple", "melon", "lemon"},
		{"orange", "lemon", "melon"},
		{"apple", "melon", "orange", "lemon"},
	}
	want6 := []int{5, 5, 4, 4, 4, 3, 3, 2}
	got5, got6, got7, got8 := dci.getClosedWordsSorted(a.items)
	if !reflect.DeepEqual(got5, want5) {
		t.Errorf("GetClosedWordsSorted trans got = %v, want %v", got5, want5)
	}
	if !reflect.DeepEqual(got6, want6) {
		t.Errorf("GetClosedWordsSorted support got = %v, want %v", got6, want6)
	}

	want7 := []int{1, 0, 0, 1, 1, 2, 1, 2}
	want8 := []int{5, 5, 5, 4, 5, 5, 4, 4}
	if !reflect.DeepEqual(got7, want7) {
		t.Errorf("GetClosedWordsSorted first appared tid got = %v, want %v", got7, want7)
	}
	if !reflect.DeepEqual(got8, want8) {
		t.Errorf("GetClosedWordsSorted last appared tid got = %v, want %v", got8, want8)
	}
}

func TestDCIClosed_Run3(t *testing.T) {
	bolShowProgress = false

	a := newFileAnalyzer(0, "", "")
	err := a.tokenizeFile("inputs/sample1.txt")
	if err != nil {
		t.Errorf("%+v", err)
	}
	//a.loadMatrix()
	matrix := tran2BitMatrix(a.trans, a.items)
	dci, err := newDCIClosed(matrix, 2, true)
	if err != nil {
		t.Errorf("newDCIClosed() error = %v", err)
		return
	}
	if err := dci.run(); err != nil {
		t.Errorf("%+v", err)
	}

	want1 := [][]int{
		{4, 3, 0, 1},
		{3, 0, 1},
		{0, 1},
	}
	got1 := dci.closedSetsToArray()
	if !reflect.DeepEqual(got1, want1) {
		t.Errorf("closedSets = %v, want %v", got1, want1)
	}
	want2 := []int{2, 3, 4}
	got2 := dci.closedSupp.getSlice()
	if !reflect.DeepEqual(got2, want2) {
		t.Errorf("closedSup = %v, want %v", got2, want2)
	}
}

func TestDCIClosed_Run2(t *testing.T) {
	bolShowProgress = false

	a := newFileAnalyzer(0, "apple", "")
	err := a.tokenizeFile("inputs/sample.txt")
	if err != nil {
		t.Errorf("%+v", err)
	}
	//a.loadMatrix()
	matrix := tran2BitMatrix(a.trans, a.items)
	dci, err := newDCIClosed(matrix, 2, false)
	if err != nil {
		t.Errorf("newDCIClosed() error = %v", err)
		return
	}
	if err := dci.run(); err != nil {
		t.Errorf("%+v", err)
	}

	want1 := [][]int{
		{4, 3, 0, 1},
		{3, 0, 1},
		{0, 1},
	}
	got1 := dci.closedSetsToArray()
	if !reflect.DeepEqual(got1, want1) {
		t.Errorf("closedSets = %v, want %v", got1, want1)
	}
	want2 := []int{2, 3, 4}
	got2 := dci.closedSupp.getSlice()
	if !reflect.DeepEqual(got2, want2) {
		t.Errorf("closedSup = %v, want %v", got2, want2)
	}
}

func TestDCIClosed4_Run(t *testing.T) {
	bolShowProgress = false

	//a, err := newFileAnalyzer("/var/log/syslog", 10, "", "")
	a := newFileAnalyzer(10, "", "")
	err := a.tokenizeFile("/var/log/syslog")
	if err != nil {
		t.Errorf("%+v", err)
	}
	//a.loadMatrix()
	matrix := tran2BitMatrix(a.trans, a.items)
	dci, err := newDCIClosed(matrix, 2, true)
	if err != nil {
		t.Errorf("newDCIClosed() error = %v", err)
		return
	}
	if err := dci.run(); err != nil {
		t.Errorf("%+v", err)
	}
}
*/
