package utils

import "testing"

func Test_aho_corasik(t *testing.T) {
	ac := NewAC()
	patterns := []string{"cat", "car", "at"}
	for _, p := range patterns {
		ac.Register(p)
	}

	line := []byte("my catcar is here")
	matches := ac.Match(line)

	if len(matches) != 2 {
		t.Errorf("Expected 2 matches, got %d", len(matches))
	}
	expectedMatches := []int{0, 1} // indices of "cat" and "car"
	for i, match := range matches {
		if match != expectedMatches[i] {
			t.Errorf("Expected match at index %d, got %d", expectedMatches[i], match)
		}
	}

	// Test with exact match
	exactMatches := ac.MatchExact(line)
	if len(exactMatches) != 0 {
		t.Errorf("Expected 0 exact match, got %d", len(exactMatches))
	}

	// Test with exact match on a different line
	exactLine := []byte("my cat is here")
	exactMatches = ac.MatchExact(exactLine)
	if len(exactMatches) != 1 {
		t.Errorf("Expected 1 exact match, got %d", len(exactMatches))
	}
	if exactMatches[0] != 0 {
		t.Errorf("Expected exact match at index 0, got %d", exactMatches[0])
	}

	line = []byte("cat")
	matches = ac.MatchExact(line)
	if len(matches) != 1 {
		t.Errorf("Expected 1 match, got %d", len(matches))
	}
	line = []byte("catcar")
	matches = ac.MatchExact(line)
	if len(matches) != 0 {
		t.Errorf("Expected 0 matches, got %d", len(matches))
	}

	// register an existing pattern
	ac.Register("cat")
	if len(ac.patterns) != 3 {
		t.Errorf("Expected 3 patterns, got %d", len(ac.patterns))
	}
}
