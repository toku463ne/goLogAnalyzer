package analyzer

//go get github.com/damnever/bitarray
import (
	"github.com/damnever/bitarray"
)

type bitMatrix struct {
	matrix       []*bitarray.BitArray
	xLen         int
	yLen         int
	support1item []int
}

func newBitMatrix(xLen, yLen int) *bitMatrix {
	b := new(bitMatrix)
	b.matrix = make([]*bitarray.BitArray, yLen)
	for i := 0; i < yLen; i++ {
		b.matrix[i] = bitarray.New(xLen)
	}
	b.support1item = make([]int, yLen)
	b.xLen = xLen
	b.yLen = yLen
	return b
}

func (b *bitMatrix) set(x, y int) error {
	_, err := b.matrix[y].Put(x, 1)
	return err
}

func (b *bitMatrix) count(y int) int {
	if y+1 > len(b.matrix) {
		return 0
	}
	return b.matrix[y].Count()
}

func (b *bitMatrix) toArray(y int) []int {
	return b.matrix[y].ToArray()
}

func (b *bitMatrix) toArrays() [][]int {
	a := make([][]int, b.yLen)
	for i := 0; i < b.yLen; i++ {
		a[i] = b.toArray(i)
	}
	return a
}

func (b *bitMatrix) getSupportFirstTime(i int) int {
	ret := b.count(i)
	b.support1item[i] = ret
	return ret
}

func (b *bitMatrix) getSupport(i int) int {
	return b.support1item[i]
}

func (b *bitMatrix) getBitSetOf(y int) *bitarray.BitArray {
	return b.matrix[y]
}

func (b *bitMatrix) getBitSetCopyOf(y int) (*bitarray.BitArray, error) {
	c := bitarray.New(b.xLen)
	d := b.matrix[y]
	for i := 0; i < b.xLen; i++ {
		bit, err := d.Get(i)
		if err != nil {
			return nil, err
		}
		c.Put(i, bit)
	}
	return c, nil
}

func tran2BitMatrix(trans1 *trans, items1 *items) *bitMatrix {
	matrix := newBitMatrix(trans1.maxTranID+1, items1.maxItemID+1)
	for i := 0; i <= trans1.maxTranID; i++ {
		for _, itemID := range trans1.get(i) {
			matrix.set(i, itemID)
		}
	}
	return matrix
}

func tranPart2BitMatrix(trans1 *trans, items1 *items,
	tranStart, tranCnt int) (*bitMatrix, map[int]int, []int) {
	subItemsMap := make(map[int]int)
	subItems := newIntArray()
	j := 0
	for i := 0; i < tranCnt; i++ {
		tran := trans1.get(tranStart + i)
		for _, item := range tran {
			if _, ok := subItemsMap[item]; !ok {
				subItemsMap[item] = j
				subItems.append(item)
				j++
			}
		}
	}
	matrix := newBitMatrix(tranCnt, j+1)
	for i := 0; i < tranCnt; i++ {
		tran := trans1.get(tranStart + i)
		for _, item := range tran {
			matrix.set(i, subItemsMap[item])
		}
	}
	return matrix, subItemsMap, subItems.getSlice()
}
