package analyzer

import (
	"fmt"
	"math"
	"regexp"
	"strconv"

	"github.com/damnever/bitarray"
)

func bitAnd(b1 *bitarray.BitArray, b2 *bitarray.BitArray) *bitarray.BitArray {
	a1 := b1.ToArray()
	a2 := b2.ToArray()
	var x, y []int
	if len(a1) <= len(a2) {
		x = a1
		y = a2
	} else {
		x = a2
		y = a1
	}
	c := bitarray.New(len(y))
	for i := range x {
		c.Put(i, x[i]*y[i])
	}
	return c
}

func bitOr(b1 *bitarray.BitArray, b2 *bitarray.BitArray) *bitarray.BitArray {
	a1 := b1.ToArray()
	a2 := b2.ToArray()
	c := bitarray.New(len(a1))
	for i := range a1 {
		if a1[i] == 1 || a2[i] == 1 {
			c.Put(i, 1)
		} else {
			c.Put(i, 0)
		}
	}
	return c
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

func printStatistics(title string,
	count int,
	maxScore float64,
	minScore float64,
	avg float64,
	std float64,
	comment string) {
	fmt.Printf("--- %s ---\n", title)
	fmt.Printf("Counts: %d\n", count)
	fmt.Printf("Max score: %f\n", maxScore)
	fmt.Printf("Min score: %f\n", minScore)
	fmt.Printf("Average: %f\n", avg)
	fmt.Printf("Standard deviation: %f\n", std)
	fmt.Printf("%s\n", comment)
}

// Round 四捨五入
func Round(num, places float64) float64 {
	shift := math.Pow(10, places)
	return roundInt(num*shift) / shift
}

// RoundUp 切り上げ
func RoundUp(num, places float64) float64 {
	shift := math.Pow(10, places)
	return roundUpInt(num*shift) / shift
}

// RoundDown 切り捨て
func RoundDown(num, places float64) float64 {
	shift := math.Pow(10, places)
	return math.Trunc(num*shift) / shift
}

// roundInt 四捨五入(整数)
func roundInt(num float64) float64 {
	t := math.Trunc(num)
	if math.Abs(num-t) >= 0.5 {
		return t + math.Copysign(1, num)
	}
	return t
}

// roundInt 切り上げ(整数)
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

/*
func sortDecSliceBy(slice []interface{},
	ref []interface{},
	less func(i, j int) bool) ([]interface{}, bool) {
	if len(slice) != len(ref) {
		return nil, false
	}
	ref2 := make([]interface{}, len(ref))
	copy(ref2, ref)
	sort.SliceStable(ref2, less)
	slice2 := make([]interface{}, len(slice))
	for i := 0; i < len(ref2); i++ {

	}
}
*/
