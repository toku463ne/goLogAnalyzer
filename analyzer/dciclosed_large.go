package analyzer

import (
	"fmt"
	"sort"
)

// LargeDCIClosed ... struct to keep Closed Sets from large transactions
type LargeDCIClosed struct {
	minSupp    int
	trans      *trans
	items      *items
	closedSets *int2dArray
	closedSupp *intArray
	firstTIDs  *intArray
	lastTIDs   *intArray
	showTIDs   bool
}

func newLargeDCIClosed(minSupp int, trans1 *trans,
	items1 *items, showTIDs bool) *LargeDCIClosed {
	ldci := new(LargeDCIClosed)
	ldci.trans = trans1
	ldci.items = items1
	ldci.minSupp = minSupp
	ldci.showTIDs = showTIDs
	//ldci.closedSets = make([]*intArray, 0)
	ldci.closedSets = newInt2dArray()
	ldci.closedSupp = newIntArray()
	ldci.firstTIDs = newIntArray()
	ldci.lastTIDs = newIntArray()
	return ldci
}

func (ldci LargeDCIClosed) run() error {
	closedsetsMap := make(map[string][]int)
	ltrans := ldci.trans.tranList.len()
	logInfo("DCIClosed algorithm...")
	for pos := 0; pos < ltrans; pos += maxBitMatrixXLen {
		xLen := 0
		if maxBitMatrixXLen > ltrans-pos {
			xLen = ltrans - pos
		} else {
			xLen = maxBitMatrixXLen
		}
		matrix, _, items := tranPart2BitMatrix(ldci.trans, ldci.items,
			pos, xLen)
		dci, err := newDCIClosed(matrix, ldci.minSupp, true)
		if err != nil {
			return err
		}
		err = dci.run()
		if err != nil {
			return err
		}
		a := dci.closedSetsToArray()
		for _, b := range a {
			k := ""
			//b2 := newIntArray()
			b2 := make([]int, len(b))
			for j, v := range b {
				//b2.append(items[v])
				b2[j] = items[v]
			}
			sort.Ints(b2)
			for _, c := range b2 {
				if k == "" {
					k = fmt.Sprintf("%d", c)
				} else {
					k = fmt.Sprintf("%s,%d", k, c)
				}
			}
			if _, ok := closedsetsMap[k]; !ok {
				closedsetsMap[k] = b2
			}
		}
		matrix = nil
	}
	logInfo("Calculating support of closed sets")
	suppMap := make(map[string]int)
	lfirstTIDsMap := make(map[string]int)
	llastTIDsMap := make(map[string]int)
	for k := range closedsetsMap {
		cs := closedsetsMap[k]
		for tranID, tran := range ldci.trans.getSlice() {
			csInTran := true
			for _, citemID := range cs {
				citemIDInTran := false
				for _, titemID := range tran {
					if citemID == titemID {
						citemIDInTran = true
						break
					}
				}
				if citemIDInTran == false {
					csInTran = false
					break
				}
			}
			if csInTran {
				if _, ok := suppMap[k]; ok {
					suppMap[k]++
				} else {
					suppMap[k] = 1
				}
				if lfTID, ok := lfirstTIDsMap[k]; ok {
					if tranID < lfTID {
						lfirstTIDsMap[k] = tranID
					}
				} else {
					lfirstTIDsMap[k] = tranID
				}
				if llTID, ok := llastTIDsMap[k]; ok {
					if tranID > llTID {
						llastTIDsMap[k] = tranID
					}
				} else {
					llastTIDsMap[k] = tranID
				}
			}
		}

		ldci.closedSets.append(cs)
		//ldci.closedSets = append(ldci.closedSets, closedsetsMap[k])
		ldci.closedSupp.append(suppMap[k])
		ldci.firstTIDs.append(lfirstTIDsMap[k])
		ldci.lastTIDs.append(llastTIDsMap[k])
	}
	return nil
}

func (ldci *LargeDCIClosed) getSortedClosedSets() ([][]int, []int, []int, []int, error) {
	dci, err := newDCIClosed(nil, ldci.minSupp, ldci.showTIDs)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	dci.closedSets = ldci.closedSets.toIntArraySlice()
	dci.firstTIDs = ldci.firstTIDs
	dci.lastTIDs = ldci.lastTIDs
	dci.closedSupp = ldci.closedSupp
	closedsets, supb, ftid, ltid := dci.getSortedClosedSets()
	return closedsets, supb, ftid, ltid, nil
}

func (ldci *LargeDCIClosed) outLargeDCIClosed(filepath string,
	rowNum int, regStr string, mask *intArray) error {
	dci, err := newDCIClosed(nil, ldci.minSupp, ldci.showTIDs)
	if err != nil {
		return err
	}
	dci.closedSets = ldci.closedSets.toIntArraySlice()
	dci.firstTIDs = ldci.firstTIDs
	dci.lastTIDs = ldci.lastTIDs
	dci.closedSupp = ldci.closedSupp

	//dci.closedSets = ldci.closedSets.getSlice()
	return dci.outDCIClosed(filepath, ldci.items, rowNum, regStr, mask)
}
