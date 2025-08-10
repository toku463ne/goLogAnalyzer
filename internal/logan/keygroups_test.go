package logan

import (
	"goLogAnalyzer/pkg/utils"
	"testing"
)

func Test_keygroups(t *testing.T) {
	testDir, err := utils.InitTestDir("Test_keygroups")
	if err != nil {
		t.Errorf("%v", err)
		return
	}
	reStr := `key=(?P<keyId>\w+) .*`

	kg, err := newKeyGroups(testDir, []string{reStr}, false, false)
	if err != nil {
		t.Errorf("Error creating keygroups: %v", err)
		return
	}

	matched_lines := []string{
		"key=1234 mached line",
		"key=5678 another matched line",
		"key=1234 mached but key duplicated",
	}

	for _, line := range matched_lines {
		keygroupId, err := kg.findAndRegister(line)
		if err != nil {
			t.Errorf("Error finding and registering keygroup: %v", err)
			return
		}
		if keygroupId == "" {
			t.Errorf("Expected keygroup ID, got empty string for line: %s", line)
			return
		}
	}

	unmatched_lines := []string{
		"no key here",
		"key= not a valid key",
	}
	for _, line := range unmatched_lines {
		keygroupId, err := kg.findAndRegister(line)
		if err != nil {
			t.Errorf("Error finding and registering keygroup: %v", err)
			return
		}
		if keygroupId != "" {
			t.Errorf("Expected no keygroup ID, got %s for line: %s", keygroupId, line)
			return
		}
	}

	// Check if the keygroup IDs are stored correctly
	if len(kg.records) != 2 {
		t.Errorf("Expected 2 unique keygroup IDs, got %d", len(kg.records))
		return
	}

	matched_keys := []string{"1234", "5678"}
	for _, key := range matched_keys {
		if kg.hasMatch([]byte(key)) == false {
			t.Errorf("Expected keygroup ID %s to be registered, but it was not found", key)
			return
		}
	}

	unmatched_keys := []string{"9999", "0000"}
	for _, key := range unmatched_keys {
		if kg.hasMatch([]byte(key)) {
			t.Errorf("Expected keygroup ID %s to not be registered, but it was found", key)
			return
		}
	}

	// Test appending log groups
	kg.appendLogGroup("1234", 1, 1001)
	kg.appendLogGroup("1234", 2, 1002)

	if len(kg.records["1234"]) != 2 {
		t.Errorf("Expected 2 log groups for keygroup ID '1234', got %d", len(kg.records["1234"]))
		return
	}

	// Test flushing the keygroups
	if err := kg.next(3); err != nil {
		t.Errorf("Error in next: %v", err)
		return
	}

	// keys in the old block will have 0 log group IDs
	for _, key := range matched_keys {
		if len(kg.records[key]) != 0 {
			t.Errorf("Expected keygroup ID %s to have 0 log group IDs after loading, but it had %d", key, len(kg.records[key]))
			return
		}
	}

	// register more keys
	kg.findAndRegister("key=8000 new keygroup")
	kg.appendLogGroup("8000", 4, 2001)
	if kg.hasMatch([]byte("8000")) == false {
		t.Errorf("Expected keygroup ID '8000' to be registered, but it was not found")
		return
	}

	if err := kg.flush(); err != nil {
		t.Errorf("Error flushing keygroups: %v", err)
		return
	}

	kg = nil // Clean up
	kg, err = newKeyGroups(testDir, []string{reStr}, false, false)
	if err != nil {
		t.Errorf("Error creating keygroups: %v", err)
		return
	}

	// load
	if err := kg.load(); err != nil {
		t.Errorf("Error loading keygroups: %v", err)
		return
	}

	if len(kg.records) != 3 {
		t.Errorf("Expected 3 unique keygroup IDs after loading, got %d", len(kg.records))
		return
	}

	// all keys should be registered
	for _, key := range matched_keys {
		if kg.hasMatch([]byte(key)) == false {
			t.Errorf("Expected keygroup ID %s to be registered after loading, but it was not found", key)
			return
		}
	}
	// they should have 0 loggroupIds
	for _, key := range matched_keys {
		if len(kg.records[key]) != 0 {
			t.Errorf("Expected keygroup ID %s to have 0 log group IDs after loading, but it had %d", key, len(kg.records[key]))
			return
		}
	}

	// also 8000 should be registered
	if kg.hasMatch([]byte("8000")) == false {
		t.Errorf("Expected keygroup ID '8000' to be registered after loading, but it was not found")
		return
	}
	// it should have 1 loggroupId
	if len(kg.records["8000"]) != 1 {
		t.Errorf("Expected keygroup ID '8000' to have 1 log group ID after loading, but it had %d", len(kg.records["8000"]))
		return
	}

}
