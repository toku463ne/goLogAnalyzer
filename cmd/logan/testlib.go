package main

import (
	"goLogAnalyzer/pkg/utils"
	"os"
	"testing"
)

func initTestDir(t *testing.T, testName string) (string, error) {
	testDir, err := utils.InitTestDir(testName)
	if err != nil {
		t.Errorf("%v", err)
		return "", nil
	}
	dataDir := testDir + "/data"
	os.Setenv("DATADIR", dataDir)
	if err := os.Setenv("DATADIR", dataDir); err != nil {
		t.Errorf("%v", err)
		return "", nil
	}
	return dataDir, nil
}
