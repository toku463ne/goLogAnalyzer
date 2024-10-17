package logan

import (
	"fmt"
)

type logTree struct {
	children map[int]*logTree
	depth    int
	groupId  int64
}

func newLogTree(depth int) *logTree {
	lt := new(logTree)
	lt.children = make(map[int]*logTree)
	lt.depth = depth
	return lt
}

// register tokens to the log tree
// leafId must be controlled by external process
// returns a leaf of the logTree
func (lt *logTree) registerTokens(tokens []int) *logTree {
	ltc := lt
	for _, termId := range tokens {
		if lttmp, ok := ltc.children[termId]; ok {
			ltc = lttmp
		} else {
			ltc.children[termId] = newLogTree(ltc.depth + 1)
			ltc = ltc.children[termId]
		}
	}
	return ltc
}

func (lt *logTree) search(tokens []int) int64 {
	ltc := lt
	for _, termId := range tokens {
		if lttmp, ok := ltc.children[termId]; ok {
			ltc = lttmp
		}
	}
	return ltc.groupId
}

// rebuildHelper is a helper function that traverses the logTree and rebuilds it with replacements
func (lt *logTree) rebuildHelper(newTree *logTree, te *terms, termCountBorder int) error {
	if lt.depth != newTree.depth {
		return fmt.Errorf("depth does not match %d != %d. should be a bug", lt.depth, newTree.depth)
	}
	for termId, child := range lt.children {
		newTermId := termId
		if te.counts[termId] < termCountBorder {
			newTermId = cAsteriskItemID
		}
		if _, ok := newTree.children[newTermId]; !ok {
			newTree.children[newTermId] = newLogTree(newTree.depth + 1)
		}
		child.rebuildHelper(newTree.children[newTermId], te, termCountBorder)
	}
	return nil
}
