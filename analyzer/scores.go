package analyzer

import (
	"bufio"
	"fmt"
	"math"
	"os"

	"github.com/gonum/stat"
)

type idfScores struct {
	scores *float64Array
	maxIDF float64
	matrix *bitMatrix
}

func newIdfScores(matrix *bitMatrix) *idfScores {
	idf := new(idfScores)
	idf.scores = newFloat64Array()
	idf.maxIDF = 0
	idf.matrix = matrix
	return idf
}

func (idf *idfScores) getIDF(itemID int) float64 {
	if idf.scores.has(itemID) {
		return idf.scores.get(itemID)
	}

	s := idf.matrix.count(itemID)
	if s == 0 {
		return 0
	}
	score := math.Log(float64(idf.matrix.xLen) / float64(s))
	idf.scores.set(itemID, score)

	if score > idf.maxIDF {
		idf.maxIDF = score
	}
	return score
}

func calcTF(itemID int, tran []int) float64 {
	cnt := 0
	for _, titem := range tran {
		if itemID == titem {
			cnt++
		}
	}
	return float64(cnt) / float64(len(tran))
}

func (idf *idfScores) calcYuScore(tran []int) float64 {
	var idfScore, yuScore float64
	for _, itemID := range tran {
		if itemID >= 0 {
			idfScore = idf.getIDF(itemID)
		} else {
			idfScore = idf.maxIDF
		}
		tf := calcTF(itemID, tran)
		yuScore += idfScore * tf
	}
	return yuScore / float64(len(tran))
}

func (idf *idfScores) outYuScoreByTime(filepath string, trans1 *trans,
	timeStampLen int) error {
	ts, idxs, scores, counts := idf.calcMaxYuScoreByTime(trans1, timeStampLen)
	ou, err := os.Create(filepath)
	if err != nil {
		return err
	}
	w := bufio.NewWriter(ou)
	defer ou.Close()
	fmt.Fprint(w, "time,count,lineNo,yuScore\n")
	for i, idx := range idxs {
		fmt.Fprintf(w, "%s,%d,%d,%f\n", ts[i], counts[i], idx, scores[i])
	}
	w.Flush()
	return nil
}

func (idf *idfScores) calcMaxYuScoreByTime(trans1 *trans,
	timeStampLen int) ([]string, []int, []float64, []int) {
	counts := newIntArray()
	scores := newFloat64Array()
	tsIdxs := newIntArray()
	times := newStringArray()
	var maxYu float64
	var maxI int
	oldT := ""
	t := ""
	cnt := 0
	j := 0
	for i, tran := range trans1.tranList.getSlice() {
		ts := trans1.tranTimeStamps.get(i)
		cnt++
		if timeStampLen > 0 && timeStampLen < len(ts) {
			t = ts[:timeStampLen]
		} else {
			t = ts
		}
		if t != oldT {
			times.append(t)
			if oldT != "" {
				tsIdxs.append(maxI)
				scores.append(maxYu)
				counts.append(cnt)
				cnt = 0
				j++
			}
			maxYu = 0.0
			maxI = j
		}
		yu := idf.calcYuScore(tran)
		if yu > maxYu {
			maxYu = yu
			maxI = trans1.mask.get(i)
		}
		oldT = t
	}
	times.append(t)
	tsIdxs.append(maxI)
	scores.append(maxYu)
	counts.append(cnt)
	return times.getSlice(), tsIdxs.getSlice(), scores.getSlice(), counts.getSlice()
}

func (idf *idfScores) yuStatistics(trans1 *trans) (int, float64, float64, float64, float64) {
	trans2 := trans1.getSlice()
	maxYu := 0.0
	maxI := -1
	minYu := -0.1
	yuScores := make([]float64, len(trans2))
	for i, tran := range trans2 {
		yu := idf.calcYuScore(tran)
		if minYu < 0 || yu < minYu {
			minYu = yu
		}
		if yu > maxYu {
			maxYu = yu
			maxI = i
		}
		yuScores[i] = yu
	}
	mean, std := stat.MeanStdDev(yuScores, nil)
	count := len(trans2)
	maxDoc := trans1.doc.get(maxI)
	printStatistics("Yu score", count, maxYu, minYu,
		mean, std, fmt.Sprintf("Max scored doc: %s", maxDoc))
	return count, maxYu, minYu, mean, std
}
