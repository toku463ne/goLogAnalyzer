package analyzer

type intArray struct {
	a       []int
	pos     int
	incSize int
}

func newIntArray() *intArray {
	a := new(intArray)
	a.a = make([]int, 1000)
	a.pos = -1
	a.incSize = 1000
	return a
}

func genIntArrayFromSlice(s []int) *intArray {
	a := new(intArray)
	a.a = s
	a.pos = len(s) - 1
	a.incSize = 1000
	return a
}

func (a *intArray) set(i, val int) {
	if len(a.a) == 0 {
		a.a = make([]int, 1000)
	}
	if i+1 > len(a.a) {
		for {
			a.a = append(a.a, make([]int, a.incSize)...)
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

func (a *intArray) get(i int) int {
	if i >= len(a.a)-1 {
		a.set(i, 0)
	}
	return a.a[i]
}

func (a *intArray) getSlice() []int {
	return a.a[:a.pos+1]
}

func (a *intArray) setSlice(aa []int) {
	a.a = aa
	a.pos = len(aa) - 1
}

func (a *intArray) len() int {
	return a.pos + 1
}

func (a *intArray) append(val int) {
	a.pos++
	a.set(a.pos, val)
}

func (a *intArray) copy() *intArray {
	b := newIntArray()
	for i := 0; i <= a.pos; i++ {
		b.set(i, a.get(i))
	}
	return b
}

func (a *intArray) finalize() {
	a.a = a.a[:a.pos+1]
}
