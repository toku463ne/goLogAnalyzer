package analyzer

import "testing"

func Test_idfScores_yuStatistics(t *testing.T) {
	bolShowProgress = false

	type args struct {
		filepath        string
		timeStampEndCol int
	}
	type want struct {
		count int
		maxYu float64
		minYu float64
		mean  float64
		std   float64
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		// TODO: Add test cases.
		{"small", args{"inputs/sample.txt", 0},
			want{6, 0.264394, 0.073473, 0.121029, 0.075504}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a, err := newFileAnalyzer(tt.args.filepath, tt.args.timeStampEndCol, "", "")
			if err != nil {
				t.Errorf("%+v", err)
			}
			//a.loadMatrix()
			matrix := tran2BitMatrix(a.trans, a.items)
			idf := newIdfScores(matrix)
			count, maxYu, minYu, mean, std := idf.yuStatistics(a.trans)
			if count != tt.want.count {
				t.Errorf("yuStatistics count: got=%d want=%d", count, tt.want.count)
			}
			if Round(maxYu, 6) != tt.want.maxYu {
				t.Errorf("yuStatistics maxYu: got=%f want=%f", maxYu, tt.want.maxYu)
			}
			if Round(minYu, 6) != tt.want.minYu {
				t.Errorf("yuStatistics minYu: got=%f want=%f", minYu, tt.want.minYu)
			}
			if Round(mean, 6) != tt.want.mean {
				t.Errorf("yuStatistics mean: got=%f want=%f", mean, tt.want.mean)
			}
			if Round(std, 6) != tt.want.std {
				t.Errorf("yuStatistics std: got=%f want=%f", std, tt.want.std)
			}
		})
	}
}
