package analyzer

import (
	"testing"
)

func Test_topNItems(t *testing.T) {
	h := newTopNItems(3)
	h.register(1, 1)
	if err := getGotExpErr("first", h.scores[0], 1.0); err != nil {
		t.Errorf("%v", err)
		return
	}
	if err := getGotExpErr("second", h.scores[1], 0.0); err != nil {
		t.Errorf("%v", err)
		return
	}

	h.register(2, 2)
	h.register(3, 3)
	if err := getGotExpErr("first", h.scores[0], 3.0); err != nil {
		t.Errorf("%v", err)
		return
	}
	if err := getGotExpErr("last", h.scores[2], 1.0); err != nil {
		t.Errorf("%v", err)
		return
	}

	h.register(4, 2.5)
	if err := getGotExpErr("not moved", h.scores[0], 3.0); err != nil {
		t.Errorf("%v", err)
		return
	}
	if err := getGotExpErr("inserted", 2.5, h.scores[1]); err != nil {
		t.Errorf("%v", err)
		return
	}
	if err := getGotExpErr("moved", 2.0, h.scores[2]); err != nil {
		t.Errorf("%v", err)
		return
	}

	h.register(5, 1)
	if err := getGotExpErr("not registered", 2, h.itemIDs[2]); err != nil {
		t.Errorf("%v", err)
		return
	}
	h.register(4, 4.0)
	if err := getGotExpErr("duplicate", 2, h.itemIDs[2]); err != nil {
		t.Errorf("%v", err)
		return
	}

}
