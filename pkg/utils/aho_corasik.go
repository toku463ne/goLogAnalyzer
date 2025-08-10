package utils

type node struct {
	ch     map[byte]int // transitions
	fail   int          // failure link
	outIdx []int        // pattern indexes that end here
}

type AC struct {
	nodes    []node
	patterns [][]byte
	patIndex map[string]int // maps pattern string to its index in patterns
}

// create a new Aho-Corasick automaton
func NewAC() *AC {
	ac := &AC{
		nodes:    make([]node, 1),
		patterns: make([][]byte, 0),
		patIndex: make(map[string]int),
	}
	ac.nodes[0].ch = make(map[byte]int)
	ac.nodes[0].fail = 0 // root node has no failure link
	return ac
}

// register a pattern in the Aho-Corasick automaton
func (ac *AC) Register(pattern string) {
	// Do nothing if pattern already registered
	if _, exists := ac.patIndex[pattern]; exists {
		return
	}

	b := []byte(pattern)
	idx := len(ac.patterns)
	ac.patterns = append(ac.patterns, b)
	ac.patIndex[pattern] = idx

	state := 0
	for _, c := range b {
		nxt, ok := ac.nodes[state].ch[c]
		if !ok {
			nxt = len(ac.nodes)
			ac.nodes = append(ac.nodes, node{ch: make(map[byte]int), fail: 0})
			ac.nodes[state].ch[c] = nxt
		}
		state = nxt
	}
	ac.nodes[state].outIdx = append(ac.nodes[state].outIdx, idx)
}

// Match returns the list of matched pattern indexes in the line
func (ac *AC) Match(line []byte) []int {
	state := 0
	var hits []int

	for _, c := range line {
		for {
			if nxt, ok := ac.nodes[state].ch[c]; ok {
				state = nxt
				break
			}
			if state == 0 {
				break
			}
			state = ac.nodes[state].fail
		}
		if len(ac.nodes[state].outIdx) > 0 {
			hits = append(hits, ac.nodes[state].outIdx...)
		}
	}
	return hits
}

// Match returns the list of matched pattern indexes in the line.
// Patterns are not part of another words, so "cat" will not match in "catalog".
func (ac *AC) MatchExact(line []byte) []int {
	state := 0
	var hits []int

	for i := 0; i < len(line); i++ {
		c := line[i]
		for {
			if nxt, ok := ac.nodes[state].ch[c]; ok {
				state = nxt
				break
			}
			if state == 0 {
				break
			}
			state = ac.nodes[state].fail
		}
		if len(ac.nodes[state].outIdx) > 0 {
			for _, idx := range ac.nodes[state].outIdx {
				if (i+1-len(ac.patterns[idx]) >= 0 && (i+1 == len(line) || line[i+1] < 'a' || line[i+1] > 'z')) &&
					(i-len(ac.patterns[idx]) < 0 || line[i-len(ac.patterns[idx])] < 'a' || line[i-len(ac.patterns[idx])] > 'z') {
					hits = append(hits, idx)
				}
			}
		}
	}
	return hits
}

// GetPatterns returns the registered patterns
func (ac *AC) GetPatterns() []string {
	patterns := make([]string, len(ac.patterns))
	for i, p := range ac.patterns {
		patterns[i] = string(p)
	}
	return patterns
}
