package analyzer

import (
	"errors"
	"fmt"
	"sort"

	"github.com/damnever/bitarray"
)

// DCIClosed ... Struct to run DCI Closed algorithm
//      Lucchese, C., Orlando, S. & Perego, Raffaele (2004), DCI_Closed: a fast and memory efficient
//       *   algorithm to mine frequent closed itemsets, Proc. 2nd IEEE ICDM Workshop on Frequent Itemset
//       *   Mining Implementations at ICDM 2004.
type DCIClosed struct {
	closedCount int
	maxItemID   int
	minSupp     int
	matrix      *bitMatrix
	closedSets  []*intArray
	closedSupp  *intArray
	firstTIDs   *intArray
	lastTIDs    *intArray
	showTIDs    bool
}

// NewDCIClosed ... Constructor of DCIClosed
//    tranList is list of transactions; ex) [[1,2], [3,5], [3,7,5,3]]
//func NewDCIClosed(trans1 *trans, items1 *items, minSupp int) (*DCIClosed, error) {
func newDCIClosed(matrix *bitMatrix, minSupp int, showTIDs bool) (*DCIClosed, error) {
	dci := new(DCIClosed)
	dci.closedCount = 0
	dci.closedSets = make([]*intArray, 0)
	dci.minSupp = minSupp
	dci.closedSupp = newIntArray()
	dci.firstTIDs = newIntArray()
	dci.lastTIDs = newIntArray()
	dci.showTIDs = showTIDs

	if matrix != nil {
		dci.matrix = matrix
		dci.maxItemID = dci.matrix.yLen - 1
	}

	//printMatrix(dci.matrix)
	return dci, nil
}

func (dci *DCIClosed) smallerAccordingToTotalOrder(i, j int, matrix *bitMatrix) bool {
	//return matrix.getSupport(i) < matrix.getSupport(j)
	si := matrix.getSupport(i)
	sj := matrix.getSupport(j)
	if si == sj {
		return i < j
	}
	return si < sj
}

// Run ... Run the algorithm
func (dci *DCIClosed) run() error {
	var closedsetTIDs *bitarray.BitArray
	closedset := newIntArray()
	preset := newIntArray()
	postset := newIntArray()
	m := dci.matrix

	for i := 0; i <= dci.maxItemID; i++ {
		supp := m.getSupportFirstTime(i)
		if supp >= dci.minSupp {
			postset.append(i)
		}
	}

	// sort post set by support acending order
	a := ([]int)(postset.getSlice())
	sort.SliceStable(a, func(i, j int) bool {
		return m.getSupport(a[i]) < m.getSupport(a[j])
	})
	postset.setSlice(a)
	//printMatrix(m)
	dci.dciClosed(true, closedset, closedsetTIDs, postset, preset, m, m, -1)

	return nil
}

func (dci *DCIClosed) isDuplicate(newgenTIDs *bitarray.BitArray,
	preset *intArray, matrix *bitMatrix) bool {
	for _, j := range preset.getSlice() {
		if newgenTIDs.Eq(matrix.getBitSetOf(j)) {
			return true
		}
	}
	return false
}

func (dci *DCIClosed) projectMatrix(matrix *bitMatrix,
	bitset *bitarray.BitArray, projectedsize int) *bitMatrix {
	newMatrix := newBitMatrix(projectedsize, dci.maxItemID+1)
	newBit := 0
	for bit := 0; bit < bitset.Len(); bit++ {
		if b, _ := bitset.Get(bit); b == 1 {
			for item := 0; item <= dci.maxItemID; item++ {
				if c, _ := dci.matrix.getBitSetOf(item).Get(bit); c == 1 {
					newMatrix.set(newBit, item)
				}
			}
			newBit++
		}
	}
	//printMatrix(newMatrix)
	return newMatrix
}

func (dci *DCIClosed) deProjectMatrix(bitset *bitarray.BitArray, itemID int) *bitarray.BitArray {
	matrix := dci.matrix
	orgBitset := matrix.getBitSetOf(itemID)
	newBitset := bitarray.New(orgBitset.Len())
	bit := 0
	for i := 0; i < orgBitset.Len(); i++ {
		if v, _ := orgBitset.Get(i); v == 1 {
			if w, _ := bitset.Get(bit); w == 1 {
				newBitset.Set(i, i)
			}
			bit++
		}
	}
	return newBitset
}

func printMatrix(matrix *bitMatrix) {
	m := matrix.matrix
	for _, b := range m {
		fmt.Printf("%+v\n", b.ToArray())
	}
}

func getBitarrayCopy(src *bitarray.BitArray) (*bitarray.BitArray, error) {
	c := bitarray.New(src.Len())
	d := src
	for i := 0; i < src.Len(); i++ {
		bit, err := d.Get(i)
		if err != nil {
			return nil, err
		}
		c.Put(i, bit)
	}
	return c, nil
}

func (dci *DCIClosed) dciClosed(firstTime bool, closedset *intArray, bitset *bitarray.BitArray,
	postset *intArray, preset *intArray, matrix *bitMatrix, originalMatrix *bitMatrix,
	originalItemID int) error {
	var err error
	postsetSlice := postset.getSlice()
	for _, i := range postsetSlice {
		var newgenTIDs *bitarray.BitArray
		if firstTime {
			newgenTIDs = matrix.getBitSetOf(i)
		} else {
			newgenTIDs = bitAnd(matrix.getBitSetOf(i), bitset)
			//newgenTIDs = intersectTIDset(closedsetTIDs, database.get(i));
		}
		if newgenTIDs.Count() >= dci.minSupp {
			closedSlice := closedset.getSlice()
			newgen := make([]int, len(closedSlice))
			copy(newgen, closedSlice)
			newgen = append(newgen, i)

			if dci.isDuplicate(newgenTIDs, preset, matrix) == false {
				closedsetNew := newIntArray()
				closedsetNew.setSlice(newgen)
				var closedsetNewTIDs *bitarray.BitArray
				if firstTime {
					closedsetNewTIDs, err = matrix.getBitSetCopyOf(i)
					if err != nil {
						return err
					}
				} else {
					closedsetNewTIDs, err = getBitarrayCopy(newgenTIDs)
					if err != nil {
						return err
					}
					//deepcopier.Copy(newgenTIDs).To(closedsetNewTIDs)
				}
				postsetNew := newIntArray()
				postsetArr := postset.getSlice()
				for _, j := range postsetArr {
					if dci.smallerAccordingToTotalOrder(i, j, originalMatrix) {
						//if i != j && originalMatrix.getSupport(i) < originalMatrix.getSupport(j) {
						if newgenTIDs.Leq(matrix.getBitSetOf(j)) {
							closedsetNew.append(j)
							// recalculate TIDS of closedsetNEW by intersection
							closedsetNewTIDs = bitAnd(closedsetNewTIDs, matrix.getBitSetOf(j))
						} else {
							// otherwise add j to the new postset
							postsetNew.append(j)
						}
					}
				}
				support := closedsetNewTIDs.Count()
				dci.closedSupp.append(support)
				dci.closedSets = append(dci.closedSets, closedsetNew)
				if dci.showTIDs {
					//a := closedsetNewTIDs.ToArray()
					//b := closedsetNew.getSlice()
					//fmt.Printf("%+v %+v\n", a, b)
					ftid, _ := dci.nextBit(closedsetNewTIDs, 0, originalItemID)
					dci.firstTIDs.append(ftid)
					ltid, _ := dci.prevBit(closedsetNewTIDs, -1, originalItemID)
					dci.lastTIDs.append(ltid)
				}

				presetNew := preset.copy()
				if firstTime {
					// THIS IS THE "Dataset projection" optimization described in the TKDE paper.
					//t := closedsetNewTIDs.ToArray()
					//fmt.Printf("%+v", t)
					projectedMatrix := dci.projectMatrix(matrix, closedsetNewTIDs, support)
					replacement := bitarray.New(support)
					replacement.Set(0, support-1)
					//printMatrix(projectedMatrix)
					dci.dciClosed(false, closedsetNew, replacement, postsetNew,
						presetNew, projectedMatrix, matrix, i)
				} else {
					dci.dciClosed(false, closedsetNew, closedsetNewTIDs, postsetNew, presetNew, matrix, originalMatrix, originalItemID)
				}
				preset.append(i)
			}
		}
	}
	//for _, cl := range dci.closedSets {
	//	cl.finalize()
	//}

	return nil
}

// ClosedSetsToArray ... convert closed sets to array
func (dci *DCIClosed) closedSetsToArray() [][]int {
	l := len(dci.closedSets)
	a := make([][]int, l)
	for i := 0; i < l; i++ {
		a[i] = dci.closedSets[i].getSlice()
	}
	return a
}

func (dci *DCIClosed) getSortedClosedSets() ([][]int, []int, []int, []int) {
	var ftid []int
	var ltid []int
	a := dci.closedSetsToArray()
	b := make([][]int, len(a))
	//copy(b, a)
	supa := dci.closedSupp.getSlice()
	supb := make([]int, len(supa))
	copy(supb, supa)

	supi := make([]int, len(supa))
	for i := 0; i < len(supi); i++ {
		supi[i] = i
	}
	sort.Slice(supi, func(i, j int) bool {
		return supb[supi[i]] > supb[supi[j]]
	})
	for i := 0; i < len(b); i++ {
		b[i] = a[supi[i]]
	}
	sort.Slice(supb, func(i, j int) bool {
		return supb[i] > supb[j]
	})

	if dci.showTIDs {
		ftido := dci.firstTIDs.getSlice()
		ltido := dci.lastTIDs.getSlice()
		ftid = make([]int, len(ftido))
		ltid = make([]int, len(ltido))
		for i := 0; i < len(supi); i++ {
			ftid[i] = ftido[supi[i]]
			ltid[i] = ltido[supi[i]]
		}
	}

	return b, supb, ftid, ltid
}

// getClosedWords ... Convert closed set to word sets, sorted by support
func (dci *DCIClosed) getClosedWords(items1 *items) ([][]string, []int) {
	s := dci.closedSets
	sup := dci.closedSupp.getSlice()
	cw := make([][]string, len(s))
	for i, t := range s {
		t2 := t.getSlice()
		tw := make([]string, len(t2))
		for j, itemID := range t2 {
			tw[j] = items1.getWord(itemID)
		}
		cw[i] = tw
	}
	return cw, sup
}

// getClosedWordsSorted ... Convert closed set to word sets, sorted by support
func (dci *DCIClosed) getClosedWordsSorted(items1 *items) ([][]string,
	[]int, []int, []int) {
	s, sup, ftid, ltid := dci.getSortedClosedSets()
	cw := make([][]string, len(s))
	for i, t := range s {
		tw := make([]string, len(t))
		for j, itemID := range t {
			tw[j] = items1.getWord(itemID)
		}
		cw[i] = tw
	}
	return cw, sup, ftid, ltid
}

func (dci *DCIClosed) getClosedWordsSortedSearched(items1 *items, regStr string, mask *intArray) ([][]string,
	[]int, []int, []int) {
	s, sup, ftid, ltid := dci.getSortedClosedSets()
	cw := make([][]string, 0)
	newsup := make([]int, 0)
	newftid := make([]int, 0)
	newltid := make([]int, 0)
	pos := 0
	for i, t := range s {
		hasReg := false
		tw := make([]string, len(t))
		for j, itemID := range t {
			w := items1.getWord(itemID)
			if searchReg(w, regStr) {
				hasReg = true
			}
			tw[j] = w
		}
		for {
			if pos >= mask.len() || mask.get(pos) == 1 {
				break
			}
			pos++
		}

		if hasReg {
			cw = append(cw, tw)
			newsup = append(newsup, sup[i])
			newftid = append(newftid, mask.get(ftid[i]))
			newltid = append(newltid, mask.get(ltid[i]))
		}
	}
	return cw, newsup, newftid, newltid
}

func (dci *DCIClosed) output(items1 *items,
	rowNum int, mask *intArray) error {
	var tokens [][]string
	var sup, ftid, ltid []int
	tokens, sup, ftid, ltid = dci.getClosedWordsSorted(items1)

	total := fmt.Sprintf("Total lines: %d\n", rowNum)
	fmt.Printf(total)
	header := fmt.Sprintf("   support      start        end closedset\n")
	fmt.Printf(header)
	for i, t := range tokens {
		line := fmt.Sprintf("%10d %10d %10d %v\n", sup[i], ftid[i]+1, ltid[i]+1, t)
		if i <= printClosedSetNum {
			fmt.Printf(line)
		}
		if _, err := fmt.Println(line); err != nil {
			return err
		}
	}
	return nil
}

func (dci *DCIClosed) nextBit(b *bitarray.BitArray, startIdx int, itemID int) (int, error) {
	if dci.matrix.xLen > b.Len() {
		b = dci.deProjectMatrix(b, itemID)
	}

	for i := startIdx; i < b.Len(); i++ {
		bit, err := b.Get(i)
		if err != nil {
			return -1, err
		}
		if bit == 1 {
			return i, nil
		}
	}
	return -1, errors.New("Next bit does not exist")
}

func (dci *DCIClosed) prevBit(b *bitarray.BitArray, lastIdx int, itemID int) (int, error) {
	if dci.matrix.xLen > b.Len() {
		b = dci.deProjectMatrix(b, itemID)
	}

	if lastIdx == -1 {
		lastIdx = b.Len() - 1
	}
	for i := lastIdx; i >= 0; i-- {
		bit, err := b.Get(i)
		if err != nil {
			return -1, err
		}
		if bit == 1 {
			return i, nil
		}
	}
	return -1, errors.New("Previous bit does not exist")
}
