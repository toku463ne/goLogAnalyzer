package analyzer

type strindex struct {
	n      map[byte]*strindex
	itemID int
	depth  int
}

func newStrIndex(depth int) *strindex {
	si := new(strindex)
	si.n = make(map[byte]*strindex, 10)
	si.depth = depth
	si.itemID = -1
	return si
}

func (si *strindex) register(itemID int, s string) {
	if len(s) == 0 || si.depth >= cMaxTermLength {
		si.itemID = itemID
		return
	}

	code := s[0]
	if _, ok := si.n[code]; !ok {
		si.n[code] = newStrIndex(si.depth + 1)
	}

	s = s[1:]
	si.n[code].register(itemID, s)
}

func (si *strindex) getItemID(s string) int {
	if len(s) == 0 {
		return si.itemID
	}
	code := s[0]
	if _, ok := si.n[code]; !ok {
		return -1
	}
	return si.n[code].getItemID(s[1:])
}

func (si *strindex) unRegister(s string) bool {
	if len(s) == 0 {
		return false
	}
	code := s[0]
	if _, ok := si.n[code]; !ok {
		return false
	}
	if len(s) == 1 {
		delete(si.n, code)
		return true
	}

	return si.n[code].unRegister(s[1:])
}
