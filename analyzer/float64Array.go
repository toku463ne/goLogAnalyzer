package analyzer

type float64Array struct {
	a       []float64
	pos     int
	incSize int
}

func newFloat64Array() *float64Array {
	a := new(float64Array)
	a.a = make([]float64, 1000)
	a.pos = -1
	a.incSize = 1000
	return a
}

func (a *float64Array) set(i int, val float64) {
	if len(a.a) == 0 {
		a.a = make([]float64, 1000)
	}
	if i+1 > len(a.a) {
		for {
			a.a = append(a.a, make([]float64, a.incSize)...)
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

func (a *float64Array) has(i int) bool {
	if a.pos >= i {
		return true
	}
	return false
}

func (a *float64Array) get(i int) float64 {
	return a.a[i]
}

func (a *float64Array) getSlice() []float64 {
	return a.a[:a.pos+1]
}

func (a *float64Array) setSlice(aa []float64) {
	a.a = aa
	a.pos = len(aa) - 1
}

func (a *float64Array) len() int {
	return a.pos + 1
}

func (a *float64Array) append(val float64) {
	a.pos++
	a.set(a.pos, val)
}

func (a *float64Array) copy() *float64Array {
	b := newFloat64Array()
	for i := 0; i <= a.pos; i++ {
		b.set(i, a.get(i))
	}
	return b
}

func (a *float64Array) finalize() {
	a.a = a.a[:a.pos+1]
}
