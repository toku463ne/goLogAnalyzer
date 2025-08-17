package logan

import (
	"goLogAnalyzer/pkg/utils"
	"testing"
)

func Test_patternkeys_detectPatternsByFirstMatch(t *testing.T) {
	testDir, err := utils.InitTestDir("Test_patternkeys_detectPatternsByFirstMatch")
	if err != nil {
		t.Errorf("%v", err)
		return
	}

	reStr := `key=(?P<patternKey>\w+) .*`
	kg, err := newPatternKeys(testDir, []string{reStr}, false, false)
	if err != nil {
		t.Errorf("Error creating patternkeys: %v", err)
		return
	}

	kgId := "1234"
	lgs := []int64{5, 6, 7, 5, 6, 7, 7, 5, 6, 7, 5}
	matches := []int{1, 0, 0, 1, 0, 0, 0, 1, 0, 0, 1}
	epochs := []int64{1, 1, 2, 11, 11, 12, 12, 21, 21, 23, 30}
	for i, groupId := range lgs {
		matched := matches[i] == 1
		epoch := epochs[i]
		kg.appendLogGroup(kgId, epoch, matched, groupId)
	}

	kgId = "5678"
	lgs = []int64{5, 6, 7, 7, 5, 6, 7, 9}
	matches = []int{1, 0, 0, 0, 1, 0, 0, 1}
	epochs = []int64{15, 15, 16, 16, 25, 25, 26, 35}
	for i, groupId := range lgs {
		matched := matches[i] == 1
		epoch := epochs[i]
		kg.appendLogGroup(kgId, epoch, matched, groupId)
	}

	patterns := kg.detectPatternsByFirstMatch()
	if len(patterns) != 4 {
		t.Errorf("Expected 4 patterns, got %d", len(patterns))
		return
	}
	expectedPatterns := map[string]map[string]*pattern{
		"5 6 7": {
			"1234": {startEpoch: 1, count: 2},
			"5678": {startEpoch: 25, count: 1},
		},
		"5 6 7 7": {
			"1234": {startEpoch: 11, count: 1},
			"5678": {startEpoch: 15, count: 1},
		},
		"5": {
			"1234": {startEpoch: 30, count: 1},
		},
		"9": {
			"5678": {startEpoch: 35, count: 1},
		},
	}
	for patternId, subPatterns := range expectedPatterns {
		if _, ok := patterns[patternId]; !ok {
			t.Errorf("Expected pattern %s not found", patternId)
			continue
		}
		for kgId, pat := range subPatterns {
			if _, ok := patterns[patternId][kgId]; !ok {
				t.Errorf("Expected keygroup ID %s for pattern %s not found", kgId, patternId)
				continue
			}
			if patterns[patternId][kgId].startEpoch != pat.startEpoch || patterns[patternId][kgId].count != pat.count {
				t.Errorf("Pattern %s for keygroup ID %s has unexpected values: got (%d, %d), want (%d, %d)",
					patternId, kgId, patterns[patternId][kgId].startEpoch, patterns[patternId][kgId].count,
					pat.startEpoch, pat.count)
			}
		}
	}
}

func Test_patternkeys_detectPatternsByPatternKeys(t *testing.T) {
	testDir, err := utils.InitTestDir("Test_patternkeys_detectPatternsByPatternKeys")
	if err != nil {
		t.Errorf("%v", err)
		return
	}

	reStr := `key=(?P<patternKey>\w+) relation=(?P<relationKey>\w+) .*`
	pk, err := newPatternKeys(testDir, []string{reStr}, false, false)
	if err != nil {
		t.Errorf("Error creating patternkeys: %v", err)
		return
	}

	pk.findAndRegister("key=p1 relation=1234 first match")
	pk.findAndRegister("p1 1st line including key")
	pk.findAndRegister("p2 2nd line including key")
	pk.findAndRegister("not related line")
	pk.appendLogGroup("p1", 1, true, 5)
	pk.appendLogGroup("p1", 1, false, 6)
	pk.appendLogGroup("p1", 2, false, 7)

	pk.findAndRegister("key=p2 relation=1234 first match")
	pk.findAndRegister("p2 1st line including key")
	pk.findAndRegister("not related line2")
	pk.findAndRegister("p2 2nd line including key")
	pk.appendLogGroup("p2", 11, true, 5)
	pk.appendLogGroup("p2", 11, false, 6)
	pk.appendLogGroup("p2", 11, false, 7)

	pk.findAndRegister("key=p3 relation=5678 first match")
	pk.findAndRegister("p3 1st line including key")
	pk.findAndRegister("not related line3")
	pk.findAndRegister("p3 2nd line including key")
	pk.appendLogGroup("p3", 15, true, 5)
	pk.appendLogGroup("p3", 15, false, 6)
	pk.appendLogGroup("p3", 16, false, 7)

	pk.findAndRegister("key=p4 relation=5678 first match")
	pk.findAndRegister("p4 1st line including key")
	pk.findAndRegister("p4 2nd line including key")
	pk.appendLogGroup("p4", 25, true, 5)
	pk.appendLogGroup("p4", 25, false, 6)
	pk.appendLogGroup("p4", 26, false, 7)
	pk.appendLogGroup("p4", 28, false, 7)

	patterns := pk.detectPatternsByPatternKeys()
	if len(patterns) != 2 {
		t.Errorf("Expected 2 patterns, got %d", len(patterns))
		return
	}

	expectedPatterns := map[string]map[string]*pattern{
		"5 6 7": {
			"relationKey:1234": {startEpoch: 1, count: 2},
			"relationKey:5678": {startEpoch: 15, count: 1},
		},
		"5 6 7 7": {
			"relationKey:5678": {startEpoch: 25, count: 1},
		},
	}
	for patternId, subPatterns := range expectedPatterns {
		if _, ok := patterns[patternId]; !ok {
			t.Errorf("Expected pattern %s not found", patternId)
			continue
		}
		for kgId, pat := range subPatterns {
			if _, ok := patterns[patternId][kgId]; !ok {
				t.Errorf("Expected keygroup ID %s for pattern %s not found", kgId, patternId)
				continue
			}
			if patterns[patternId][kgId].startEpoch != pat.startEpoch || patterns[patternId][kgId].count != pat.count {
				t.Errorf("Pattern %s for keygroup ID %s has unexpected values: got (%d, %d), want (%d, %d)",
					patternId, kgId, patterns[patternId][kgId].startEpoch, patterns[patternId][kgId].count,
					pat.startEpoch, pat.count)
			}
		}
	}

}
