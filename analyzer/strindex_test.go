package analyzer

import "testing"

func TestStrIndex_run1(t *testing.T) {
	s := []string{
		"abc",
	}
	si := newStrIndex(0)
	for i, t := range s {
		si.register(i, t)
	}

	if si.n[tob("a")].n[tob("b")].n[tob("c")].itemID != 0 {
		t.Errorf("register not correct")
	}

	s = []string{
		"abc", "abc", "abc",
		"abd", "abd",
		"ab", "ab",
		"a",
	}

	si = newStrIndex(0)
	for i, t := range s {
		si.register(i, t)
	}

	if si.n[tob("a")].n[tob("b")].n[tob("c")].itemID != 2 {
		t.Errorf("register not correct")
	}

	if !si.unRegister("abc") {
		t.Errorf("unRegister not correct")
	}
	if si.getItemID("abc") != -1 {
		t.Errorf("unRegister not correct")
	}
	if si.unRegister("abc") {
		t.Errorf("unRegister not correct")
	}

}
