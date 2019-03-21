package analyzer

import (
	"reflect"
	"testing"
)

func TestFileAnalyzer_tokenizeFile(t *testing.T) {
	type fields struct {
		filepath        string
		timeStampEndCol int
	}
	tests := []struct {
		name    string
		fields  fields
		want    [][]int
		want2   []string
		wantErr bool
	}{
		// TODO: Add test cases.
		{"normal1", fields{"inputs/test1.txt", 0}, [][]int{{0, 1}}, []string{""}, false},
		{"normal2", fields{"inputs/test2.txt", 15}, [][]int{{0, 1}}, []string{"Feb 28 22:50:31"}, false},
		{"normal3", fields{"inputs/sample.txt", 0},
			[][]int{{0, 1, 2}, {3, 1, 4}, {0, 3, 1, 4}, {3, 4}, {0, 3, 1, 4}, {0, 3, 1, 1}},
			[]string{"", "", "", "", "", ""}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a, err := newFileAnalyzer(tt.fields.filepath, tt.fields.timeStampEndCol, "", "")
			if (err != nil) != tt.wantErr {
				t.Errorf("FileAnalyzer.tokenizeFile() error = %v, wantErr %v", err, tt.wantErr)
			} else {
				tokens := a.getTokens()
				if !reflect.DeepEqual(tt.want, tokens) {
					t.Errorf("tokens: not match. want=%+v got=%+v", tt.want, tokens)
				}
				timstamp := a.getTimeStamps()
				if !reflect.DeepEqual(tt.want2, timstamp) {
					t.Errorf("timeStamp: not match. want=%+v got=%+v", tt.want2, timstamp)
				}
			}
		})
	}
}
