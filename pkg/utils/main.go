package utils

import (
	"bufio"
	"compress/gzip"
	"encoding/csv"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"syscall"
	"time"
	"unicode"

	"github.com/pkg/errors"
)

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

func Str2Timestamp(dateFormat, dateStr string) (time.Time, error) {
	timestamp, err := time.Parse(dateFormat, dateStr)
	if err != nil {
		return time.Time{}, err
	}
	return timestamp.UTC(), nil
}

func Str2date(dateFormat, dateStr string) (time.Time, error) {
	parsedDate, err := time.Parse(dateFormat, dateStr)
	if err != nil {
		return time.Time{}, err
	}

	now := time.Now()

	// Use parsed date components, but fill in missing parts with defaults
	year := parsedDate.Year()
	month := parsedDate.Month()
	day := parsedDate.Day()

	if month == 0 || (year == 0 && !strings.Contains(dateFormat, "01") && !strings.Contains(dateFormat, "Jan")) {
		month = now.Month()
	}

	if day == 0 || (year == 0 && !strings.Contains(dateFormat, "2")) {
		day = now.Day()
	}

	if year == 0 {
		year = now.Year()
	}

	// Adjust year if the parsed month is in the future compared to the current month
	if month > now.Month() && year == now.Year() {
		year--
	}

	hour, minute, second := parsedDate.Hour(), parsedDate.Minute(), parsedDate.Second()

	finalDate := time.Date(year, month, day, hour, minute, second, 0, time.Local)
	return finalDate, nil
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

func CopyFile(src, dst string) (int64, error) {
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

func QuickSort(a []int64, s []string, i, j int) {
	if i == j {
		return
	}
	p := _pivot(a, i, j)
	if p != -1 {
		k := _partition(a, s, i, j, a[p])
		QuickSort(a, s, i, k-1)
		QuickSort(a, s, k, j)
	}
}

func GetSortedGlob(pathRegex string) ([]int64, []string, error) {
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

	QuickSort(filesEpoch, fileNames, 0, len(fileNames)-1)
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

func EnsureDir(dirPath string) error {
	if err := os.MkdirAll(dirPath, 0755); err != nil && !os.IsExist(err) {
		return errors.WithStack(err)
	}
	return nil
}

func IsInt(s string) bool {
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

func IsRealNumber(s string) bool {
	if len(s) == 0 {
		return false
	}

	dotCount := 0
	for i, c := range s {
		if c == '.' {
			// Only allow one dot and it should not be at the start or end
			dotCount++
			if dotCount > 1 || i == 0 || i == len(s)-1 {
				return false
			}
		} else if !unicode.IsDigit(c) {
			return false
		}
	}
	return true
}

func IsNumeric(s string) bool {
	_, err := strconv.ParseFloat(s, 64)
	return err == nil
}

func GetRegex(reStr string) *regexp.Regexp {
	if reStr == "" {
		return nil
	}

	//return regexp.MustCompile(fmt.Sprintf(".*%s.*", reStr))
	return regexp.MustCompile(`` + reStr)
}

func Re2str(re *regexp.Regexp) string {
	if re == nil {
		return ""
	}
	return re.String()
}

func RemovePath(pathRegex string) error {
	fileNames, _ := filepath.Glob(pathRegex)
	for _, p := range fileNames {
		if err := os.Remove(p); err != nil {
			return errors.WithStack(err)
		}
	}
	return nil
}

func RemoveDirectory(dir string) error {
	// Check if the directory exists
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		// Directory does not exist, nothing to do
		return nil
	}

	// Remove the directory and its contents
	err := os.RemoveAll(dir)
	if err != nil {
		return fmt.Errorf("failed to remove directory %s: %w", dir, err)
	}

	return nil
}

func TimespecToTime(ts syscall.Timespec) time.Time {
	return time.Unix(int64(ts.Sec), int64(ts.Nsec))
}

// Struct to hold the value and its original index
type ValueIndex struct {
	Value float64
	Index int
}

func SortIndexByValue(values []float64, isAsc bool) []int {
	indexes := make([]int, len(values))
	for i, _ := range values {
		indexes[i] = i
	}
	if isAsc {
		sort.Slice(indexes, func(i, j int) bool {
			return values[i] < values[j]
		})
	} else {
		sort.Slice(indexes, func(i, j int) bool {
			return values[i] > values[j]
		})
	}
	return indexes
}

func SortIndexByIntValue(values []int, isAsc bool) []int {
	indexes := make([]int, len(values))
	for i, _ := range values {
		indexes[i] = i
	}
	if isAsc {
		sort.Slice(indexes, func(i, j int) bool {
			return values[i] < values[j]
		})
	} else {
		sort.Slice(indexes, func(i, j int) bool {
			return values[i] > values[j]
		})
	}
	return indexes
}

func AddDaysToEpoch(epoch int64, days int) int64 {
	epochTime := time.Unix(epoch, 0)
	epochTime = epochTime.AddDate(0, 0, days)
	return epochTime.Unix()
}

func StringToInt64(s string) int64 {
	// Use strconv.ParseInt to convert string to int64
	i, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0
	}
	return i
}

func NextDivisibleByN(i, n int) int {
	if i%n == 0 {
		return i
	}
	return ((i / n) + 1) * 10000
}

func GetUnitsecs(frequency int) int64 {
	unitsecs := 3600 * 24
	switch frequency {
	case CFreqDay:
		unitsecs = 3600 * 24
	case CFreqHour:
		unitsecs = 3600
	case CFreqMinute:
		unitsecs = 60
	default:
		unitsecs = 3600 * 24
	}
	return int64(unitsecs)
}

func GetDatetimeFormat(frequency string) string {
	format := "2006-01-02 15:04:05"
	switch frequency {
	case "day":
		format = "2006-01-02"
	case "hour":
		format = "2006-01-02 15"
	case "minute":
		format = "2006-01-02 15:04"
	}
	return format
}

func GetDatetimeFormatFromUnitSecs(unitSecs int64) string {
	format := "2006-01-02 15:04:05"

	if unitSecs >= 3600*24 {
		format = "2006-01-02"
	} else if unitSecs >= 3600 {
		format = "2006-01-02 15"
	} else if unitSecs >= 60 {
		format = "2006-01-02 15:04"
	} else {
		format = "2006-01-02 15:04:03"
	}
	return format
}

func ReadCsv(csfvile string, separator rune, skipHeader bool) ([]string, [][]string, error) {
	file, err := os.Open(csfvile)
	if err != nil {
		return nil, nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.Comma = separator

	var header []string
	if !skipHeader {
		// Read the header line
		header, err = reader.Read()
		if err != nil {
			return nil, nil, err
		}
	}

	var records [][]string

	for {
		line, err := reader.Read()
		if err != nil {
			break // EOF or error
		}
		records = append(records, line)
	}

	return header, records, nil
}

func Slice2File(lines []string, filePath string) error {
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	for _, line := range lines {
		_, err := writer.WriteString(line + "\n")
		if err != nil {
			return err
		}
	}
	return writer.Flush()
}

func ReadFile2Slice(filePath string) ([]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return lines, nil
}

func Base36ToInt64(base36Str string) (int64, error) {
	// Parse the base-36 string into an int64
	result, err := strconv.ParseInt(base36Str, 36, 64)
	if err != nil {
		return 0, err
	}
	return result, nil
}

func Int64Tobase36(n int64) string {
	return strconv.FormatInt(n, 36)
}

func ErrorStack(format string, args ...interface{}) error {
	return errors.WithStack(errors.New(fmt.Sprintf(format, args...)))
}

func Replace(s, target, replacement, separators string) string {
	// Function to determine if a character is a separator
	isSeparator := func(r rune) bool {
		return strings.ContainsRune(separators, r)
	}

	// Result string builder for efficiency
	var result strings.Builder
	part := ""
	for _, ch := range s {
		if isSeparator(ch) {
			// Process the current part before the separator
			if strings.EqualFold(part, target) {
				result.WriteString(replacement)
			} else {
				result.WriteString(part)
			}

			// Add the separator to the result
			result.WriteRune(ch)
			// Reset part for the next word
			part = ""
		} else {
			// Build up the part
			part += string(ch)
		}
	}

	// Handle any remaining part after the loop
	if strings.EqualFold(part, target) { // Use EqualFold for case-insensitive comparison
		result.WriteString(replacement)
	} else {
		result.WriteString(part)
	}
	return result.String()
}

func FloatToStringSlice(floats []float64) []string {
	strs := make([]string, len(floats))
	for i, f := range floats {
		strs[i] = strconv.FormatFloat(f, 'f', -1, 64)
	}
	return strs
}

func IntToStringSlice(ints []int) []string {
	strs := make([]string, len(ints))
	for i, it := range ints {
		strs[i] = strconv.Itoa(it)
	}
	return strs
}

func CountFileLines(filePath string) (int, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return 0, err
	}
	defer file.Close()

	lineCount := 0
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		lineCount++
	}

	if err := scanner.Err(); err != nil {
		return 0, err
	}

	return lineCount, nil
}

func CountGzFileLines(filePath string) (int, error) {
	// Open the gz file
	file, err := os.Open(filePath)
	if err != nil {
		return 0, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Create a gzip reader
	gzReader, err := gzip.NewReader(file)
	if err != nil {
		return 0, fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer gzReader.Close()

	// Use bufio scanner to count lines from the decompressed file
	scanner := bufio.NewScanner(gzReader)
	lineCount := 0
	for scanner.Scan() {
		lineCount++
	}

	// Check for errors during scanning
	if err := scanner.Err(); err != nil {
		return 0, fmt.Errorf("error reading lines: %w", err)
	}

	return lineCount, nil
}

func CalculateStats(values []float64) (mean float64, stdDev float64) {
	sum := 0.0
	for _, v := range values {
		sum += v
	}
	mean = sum / float64(len(values))

	var varianceSum float64
	for _, v := range values {
		varianceSum += math.Pow(v-mean, 2)
	}
	stdDev = math.Sqrt(varianceSum / float64(len(values)))
	return
}

func DetectPeriodicityByThreshold(values []float64, upperThreshold, lowerThreshold float64) bool {
	crossingPattern := make([]int, len(values))

	// Record threshold crossings
	for i, v := range values {
		if v > upperThreshold {
			crossingPattern[i] = 1 // Above upper threshold
		} else if v < lowerThreshold {
			crossingPattern[i] = -1 // Below lower threshold
		} else {
			crossingPattern[i] = 0 // Within thresholds
		}
	}

	// Check for periodicity in the crossing pattern
	n := len(crossingPattern)
	for period := 2; period <= n/2; period++ {
		isPeriodic := true
		for i := 0; i < n-period; i++ {
			if crossingPattern[i] != crossingPattern[i+period] {
				isPeriodic = false
				break
			}
		}
		if isPeriodic {
			return true
		}
	}

	return false
}

func GetNdaysBefore(N int) int64 {
	// Get the current time
	now := time.Now()

	// Calculate the time N days before today
	nDaysBefore := now.AddDate(0, 0, -N)

	// Convert to Unix epoch time
	epoch := nDaysBefore.Unix()

	return epoch
}
