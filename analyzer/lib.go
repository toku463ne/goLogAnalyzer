package analyzer

import (
	"fmt"
	"io"
	"log"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
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

func ShowDebug(msg string) {
	if IsDebug {
		log.Printf("[DEBUG] %s", msg)
	}
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

func Str2Epoch(dateFormat, str string) (int64, error) {
	if dt, err := time.Parse(dateFormat, str); err != nil {
		return 0, errors.WithStack(err)
	} else {
		return dt.Unix(), nil
	}
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
		return nil, nil, errors.New(fmt.Sprintf("No files found at %s", pathRegex))
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

func UniqueStringSplit(s []string) []string {
	m := make(map[string]bool, 0)
	for _, v := range s {
		m[v] = true
	}
	u := make([]string, 0)
	for k := range m {
		u = append(u, k)
	}
	return u
}

// PathExist ..
func PathExist(path string) bool {
	if _, err := os.Stat(path); err != nil {
		return false
	}
	return true
}

func ensureDir(dirPath string) error {
	if err := os.MkdirAll(dirPath, 0755); err != nil && !os.IsExist(err) {
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

func checkMatchRate(s1, s2 []int) float64 {
	i := 0
	j := 0
	cnt := 0
	base := math.Max(float64(len(s1)), float64(len(s2)))
	for {
		if i >= len(s1) || j >= len(s2) {
			break
		}

		if s1[i] < s2[j] {
			i++
		} else if s1[i] > s2[j] {
			j++
		} else {
			cnt++
			i++
			j++
		}
	}
	return float64(cnt) / base
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

func TimeToString(t time.Time) string {
	str := t.Format(CDefaultTimestampLayout)
	return str
}

func EpochToString(epoch int64) string {
	str := TimeToString(time.Unix(epoch, 0).UTC())
	return str
}

func DateStringToEpoch(date string) (int64, error) {
	t, err := time.Parse(time.RFC3339, date+"T00:00:00Z")
	if err != nil {
		return -1, err
	}
	return t.Unix(), nil
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

// scores: list of term scores in the tran
func calcNAvgScore(scores []float64, scoreStyle, scoreNSize int) float64 {
	score := 0.0
	tranSize := len(scores)
	if tranSize >= scoreNSize {
		sort.Slice(scores, func(i, j int) bool { return scores[i] > scores[j] })
		mid := int(float64(scoreNSize) / 2.0)
		j := 0
		for i := 0; i < mid; i++ {
			score += scores[i]
			j++
		}
		inc := 1
		if scoreStyle == cScoreNDistAvg {
			inc = int(float64(tranSize*2) / float64(scoreNSize))
		}
		i := mid
		for j < scoreNSize && i < tranSize {
			score += scores[i]
			i += inc
			if i >= tranSize {
				score += scores[tranSize-1]
				break
			}
			j++
		}

	} else {
		minScore := 0.0
		scoreTotal := 0.0
		for _, s := range scores {
			if minScore == 0 || s < minScore {
				minScore = s
			}
			scoreTotal += s
		}
		score = scoreTotal + minScore*float64(scoreNSize-tranSize)/2.0
	}
	score /= float64(scoreNSize)
	return score
}

func isNumeric(s string) bool {
	_, err := strconv.ParseFloat(s, 64)
	return err == nil
}
