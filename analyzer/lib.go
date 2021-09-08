package analyzer

import (
	"fmt"
	"io"
	"log"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"time"
	"unicode"

	"github.com/pkg/errors"
)

func InitLog(rootDir string) {
	if rootDir != "" {
		logFile := fmt.Sprintf("%s/analyzer.log", rootDir)
		w, _ := os.OpenFile(logFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		multiLogFile := io.MultiWriter(os.Stdout, w)
		log.SetOutput(multiLogFile)
	}
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
}

func searchReg(s, reStr string) bool {
	re := regexp.MustCompile(fmt.Sprintf(".*%s.*", reStr))
	return re.MatchString(s)
}

// Round ...
func Round(num, places float64) float64 {
	shift := math.Pow(10, places)
	return roundInt(num*shift) / shift
}

// RoundUp ...
func RoundUp(num, places float64) float64 {
	shift := math.Pow(10, places)
	return roundUpInt(num*shift) / shift
}

// RoundDown ...
func RoundDown(num, places float64) float64 {
	shift := math.Pow(10, places)
	return math.Trunc(num*shift) / shift
}

// roundInt
func roundInt(num float64) float64 {
	t := math.Trunc(num)
	if math.Abs(num-t) >= 0.5 {
		return t + math.Copysign(1, num)
	}
	return t
}

// roundInt
func roundUpInt(num float64) float64 {
	t := math.Trunc(num)
	return t + math.Copysign(1, num)
}

func copyFile(src, dst string) (int64, error) {
	sourceFileStat, err := os.Stat(src)
	if err != nil {
		return 0, err
	}

	if !sourceFileStat.Mode().IsRegular() {
		return 0, fmt.Errorf("%s is not a regular file", src)
	}

	source, err := os.Open(src)
	if err != nil {
		return 0, err
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		return 0, err
	}
	defer destination.Close()
	nBytes, err := io.Copy(destination, source)
	return nBytes, err
}

func _pivot(a []int64, i, j int) int {
	k := i + 1
	for k <= j && a[i] == a[k] {
		k++
	}
	if k > j {
		return -1
	}
	if a[i] >= a[k] {
		return i
	}
	return k
}

func _partition(a []int64, s []string, i, j int, x int64) int {
	l := i
	r := j

	for l <= r {
		for l <= j && a[l] < x {
			l++
		}
		for r >= i && a[r] >= x {
			r--
		}
		if l > r {
			break
		}
		t := a[l]
		s1 := s[l]
		a[l] = a[r]
		s[l] = s[r]
		a[r] = t
		s[r] = s1
		l++
		r--
	}
	return l
}

func quickSort(a []int64, s []string, i, j int) {
	if i == j {
		return
	}
	p := _pivot(a, i, j)
	if p != -1 {
		k := _partition(a, s, i, j, a[p])
		quickSort(a, s, i, k-1)
		quickSort(a, s, k, j)
	}
}

func getSortedGlob(pathRegex string) ([]int64, []string, error) {
	fileNames, err := filepath.Glob(pathRegex)
	if err != nil {
		return nil, nil, err
	}
	if fileNames == nil {
		return nil, nil, errors.New("No filed found")
	}
	filesEpoch := make([]int64, len(fileNames))

	for i, fileName := range fileNames {
		file, _ := os.Stat(fileName)
		//ts, _ := times.Stat(fileName)
		t := file.ModTime()
		//t := ts.BirthTime()
		filesEpoch[i] = t.Unix()
	}

	quickSort(filesEpoch, fileNames, 0, len(fileNames)-1)
	return filesEpoch, fileNames, nil
}

// PathExist ..
func pathExist(path string) bool {
	if _, err := os.Stat(path); err != nil {
		return false
	}
	return true
}

func ensureDir(dirPath string) error {
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		os.MkdirAll(dirPath, 0755)
	} else if err != nil {
		return errors.WithStack(err)
	}
	return nil
}

func isInt(s string) bool {
	if len(s) > 1 && string(s[0]) == "0" {
		return false
	}
	for _, c := range s {
		if !unicode.IsDigit(c) {
			return false
		}
	}
	return true
}

func getRegex(reStr string) *regexp.Regexp {
	if reStr == "" {
		return nil
	}

	return regexp.MustCompile(fmt.Sprintf(".*%s.*", reStr))
}

func re2str(re *regexp.Regexp) string {
	if re == nil {
		return ""
	}
	return re.String()
}

func removePath(pathRegex string) error {
	fileNames, _ := filepath.Glob(pathRegex)
	for _, p := range fileNames {
		if err := os.Remove(p); err != nil {
			return errors.WithStack(err)
		}
	}
	return nil
}

func registerNTopRareRec(
	nTopRareLogs []*colLogRecords,
	minTopRareScore float64,
	rowID int64,
	score float64, text string) ([]*colLogRecords, float64) {
	if minTopRareScore > 0 && score <= minTopRareScore {
		return nTopRareLogs, minTopRareScore
	}
	newTopN := make([]*colLogRecords, len(nTopRareLogs))
	logr2 := new(colLogRecords)
	logr2.rowid = rowID
	logr2.score = score
	logr2.record = text
	for i, logr := range nTopRareLogs {
		if logr == nil {
			newTopN[i] = logr2
			break
		} else if score == logr.score && rowID == logr.rowid {
			return nTopRareLogs, minTopRareScore
		} else if score > logr.score {
			newTopN[i] = logr2
			oldScore2 := 0.0
			for j := i + 1; j < len(nTopRareLogs); j++ {
				if nTopRareLogs[j-1] == nil {
					break
				}
				score2 := nTopRareLogs[j-1].score
				if nTopRareLogs[j-1].record == "" {
					minTopRareScore = oldScore2
					break
				}
				if j >= cNTopRareRecords-1 {
					minTopRareScore = score2
				}
				newTopN[j] = nTopRareLogs[j-1]
				oldScore2 = score2
			}
			break
		} else {
			newTopN[i] = logr
		}
	}
	return newTopN, minTopRareScore
}

func getColIdx(tableName, colName string) int {
	cols, ok := tableDefs[tableName]
	if !ok {
		return -1
	}
	for i, col := range cols {
		if col == colName {
			return i
		}
	}
	return -1
}

func timeToString(t time.Time) string {
	str := t.Format(cTimestampLayout)
	return str
}

func epochToString(epoch int64) string {
	str := timeToString(time.Unix(epoch, 0).UTC())
	return str
}

func getCurrentEpoch() int64 {
	now := time.Now()
	return now.Unix()
}

func getWeight(idx int) int {
	return int(math.Log(float64(idx))) + 1
}

func getBottoms(n []int, baseLine int) []int {
	if baseLine == 0 {
		sum := 0.0
		cnt := 0
		for _, v := range n {
			if v > 0 {
				sum += float64(v)
				cnt++
			}
		}
		mean := float64(sum) / float64(cnt)
		baseLine = int(mean)
	}

	bottoms := make([]int, 0)
	min := 0
	mini := 0
	passedBottom := false
	prev := 0
	for i := len(n) - 1; i >= 0; i-- {
		if n[i] == 0 {
			continue
		}

		if min == 0 || n[i] < min {
			min = n[i]
			mini = i
			if n[i] < prev {
				passedBottom = false
			}
		}

		if (!passedBottom && n[i] > min) || len(bottoms) == 0 {
			bottoms = append(bottoms, mini)
			passedBottom = true
		}
		if n[i] > min {
			min = baseLine
		}
		prev = n[i]
	}
	return bottoms
}
