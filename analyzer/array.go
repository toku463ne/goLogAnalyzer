package analyzer

type array struct {
	a       []interface{}
	pos     int
	incSize int
}

func newArray() *array {
	a := new(array)
	a.a = make([]interface{}, 1000)
	a.pos = -1
	a.incSize = 1000
	return a
}

func (a *array) set(i int, val interface{}) {
	if len(a.a) == 0 {
		a.a = make([]interface{}, 1000)
	}
	if i+1 > len(a.a) {
		for {
			a.a = append(a.a, make([]interface{}, a.incSize)...)
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

func (a *array) get(i int) interface{} {
	return a.a[i]
}

func (a *array) getSlice() []interface{} {
	return a.a[:a.pos+1]
}

func (a *array) setSlice(aa []interface{}) {
	a.a = aa
	a.pos = len(aa) - 1
}

func (a *array) len() int {
	return a.pos + 1
}

func (a *array) size() int {
	return len(a.a)
}

func (a *array) append(val int) {
	a.pos++
	a.set(a.pos, val)
}

func (a *array) copy() *array {
	b := newArray()
	for i := 0; i <= a.pos; i++ {
		b.set(i, a.get(i))
	}
	return b
}

func (a *array) finalize() {
	a.a = a.a[:a.pos+1]
}
