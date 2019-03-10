package analyzer

type stringArray struct {
	a       []string
	pos     int
	incSize int
}

func newStringArray() *stringArray {
	a := new(stringArray)
	a.a = make([]string, 1000)
	a.pos = -1
	a.incSize = 1000
	return a
}

func (a *stringArray) set(i int, val string) {
	if len(a.a) == 0 {
		a.a = make([]string, 1000)
	}
	if i+1 > len(a.a) {
		for {
			a.a = append(a.a, make([]string, a.incSize)...)
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

func (a *stringArray) get(i int) string {
	return a.a[i]
}

func (a *stringArray) getSlice() []string {
	return a.a[:a.pos+1]
}

func (a *stringArray) setSlice(aa []string) {
	a.a = aa
	a.pos = len(aa) - 1
}

func (a *stringArray) len() int {
	return a.pos + 1
}

func (a *stringArray) append(val string) {
	a.pos++
	a.set(a.pos, val)
}

func (a *stringArray) copy() *stringArray {
	b := newStringArray()
	for i := 0; i <= a.pos; i++ {
		b.set(i, a.get(i))
	}
	return b
}

func (a *stringArray) finalize() {
	a.a = a.a[:a.pos+1]
}
