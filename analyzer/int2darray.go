package analyzer

type int2dArray struct {
	a       [][]int
	pos     int
	incSize int
}

func newInt2dArray() *int2dArray {
	a := new(int2dArray)
	a.a = make([][]int, 1000)
	a.pos = -1
	a.incSize = 1000
	return a
}

func (a *int2dArray) set(i int, val []int) {
	if len(a.a) == 0 {
		a.a = make([][]int, 1000)
	}
	if i+1 > len(a.a) {
		for {
			a.a = append(a.a, make([][]int, a.incSize)...)
			if i < len(a.a) {
				break
			}
		}
	}
	a.a[i] = val
	if i > a.pos {
		a.pos = i
	}
}

func (a *int2dArray) get(i int) []int {
	return a.a[i]
}

func (a *int2dArray) getSlice() [][]int {
	return a.a[:a.pos+1]
}

func (a *int2dArray) setSlice(aa [][]int) {
	a.a = aa
	a.pos = len(aa) - 1
}

func (a *int2dArray) len() int {
	return a.pos + 1
}

func (a *int2dArray) append(val []int) {
	a.pos++
	a.set(a.pos, val)
}

func (a *int2dArray) copy() *int2dArray {
	b := newInt2dArray()
	for i := 0; i <= a.pos; i++ {
		b.set(i, a.get(i))
	}
	return b
}

func (a *int2dArray) finalize() {
	a.a = a.a[:a.pos+1]
}

func (a *int2dArray) toIntArraySlice() []*intArray {
	a2d := make([]*intArray, len(a.getSlice()))
	for i, s := range a.getSlice() {
		a2d[i] = genIntArrayFromSlice(s)
	}
	return a2d
}
