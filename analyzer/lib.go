package analyzer

import (
	"fmt"
	"io"
	"log"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"syscall"
	"time"
	"unicode"

	"github.com/pkg/errors"

	_ "github.com/mattn/go-sqlite3"
)

func setLogLevelByStr(logLevelStr string) {
	switch logLevelStr {
	case "error":
		curLogLevel = cLogLevelError
	case "debug":
		curLogLevel = cLogLevelDebug
	default:
		curLogLevel = cLogLevelInfo
	}
}

func logmsg(logLevel int, msg string) {
	if curLogLevel >= logLevel {
		log.Printf("[%d] %s\n", os.Getpid(),
			msg)
	}
}

func logDebug(msg string) {
	logmsg(cLogLevelDebug, fmt.Sprintf("DEBUG - %s", msg))
}

func logInfo(msg string) {
	logmsg(cLogLevelInfo, fmt.Sprintf("INFO - %s", msg))
}

func logError(msg string) {
	logmsg(cLogLevelError, fmt.Sprintf("ERROR - %s", msg))
}

func searchReg(s, reStr string) bool {
	re := regexp.MustCompile(fmt.Sprintf(".*%s.*", reStr))
	if re.MatchString(s) {
		return true
	}
	return false
}

func getMapKeys(m map[string]interface{}) []string {
	keys := make([]string, len(m))

	i := 0
	for k := range m {
		keys[i] = k
		i++
	}
	return keys
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

func argParseANum(args map[string]string, key string) (int, error) {
	v, ok := args[key]
	if ok == false {
		v = "0"
	}
	vs, err := strconv.Atoi(v)
	if err != nil {
		return -1, fmt.Errorf("%s must be integer", key)
	}
	return vs, nil
}

func timeToString(t time.Time) string {
	str := t.Format(cTimestampLayout)
	return str
}

func epochToString(epoch int64) string {
	str := timeToString(time.Unix(epoch, 0).UTC())
	return str
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

func _pivotFloatInt(a []float64, i, j int) int {
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

func _partitionFloatInt(a []float64, s []int, i, j int, x float64) int {
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

func quickSortFloatInt(a []float64, s []int, i, j int) {
	if i == j {
		return
	}
	p := _pivotFloatInt(a, i, j)
	if p != -1 {
		k := _partitionFloatInt(a, s, i, j, a[p])
		quickSortFloatInt(a, s, i, k-1)
		quickSortFloatInt(a, s, k, j)
	}
}

func timespecToTime(ts syscall.Timespec) time.Time {
	return time.Unix(int64(ts.Sec), int64(ts.Nsec))
}

func getSortedGlob(pathRegex string) ([]int64, []string) {
	fileNames, _ := filepath.Glob(pathRegex)
	if fileNames == nil {
		return nil, nil
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
	return filesEpoch, fileNames
}

func extError(err error, msg string) error {
	return errors.WithStack(errors.Wrapf(err, msg))
}

func countDigits(i int) (count int) {
	for i != 0 {

		i /= 10
		count = count + 1
	}
	return count
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

func tob(a string) byte {
	return []byte(a)[0]
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

func getCurrentEpoch() int64 {
	now := time.Now()
	return now.Unix()
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
