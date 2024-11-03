package main

import (
	"goLogAnalyzer/pkg/utils"
	"os"
	"testing"
)

func runGroups(t *testing.T, testName, config string) [][]string {
	testDir, err := utils.InitTestDir(testName)
	if err != nil {
		t.Errorf("%v", err)
		return nil
	}
	dataDir := testDir + "/data"
	os.Setenv("DATADIR", dataDir)
	if err := os.Setenv("DATADIR", dataDir); err != nil {
		t.Errorf("%v", err)
		return nil
	}

	// run test
	os.Args = []string{"logan", "groups", "-c", config, "-o", testDir}
	main()

	_, records, err := utils.ReadCsv(testDir + "/logGroups.csv")
	if err != nil {
		t.Errorf("%v", err)
		return nil
	}
	return records
}

func getGroupCount(records [][]string, displayString string) string {
	for _, row := range records {
		if row[2] == displayString {
			return row[1]
		}
	}
	return "0"
}

func Test_groups_001_filter(t *testing.T) {
	tbl := runGroups(t, "Test_groups_001_filter", "groups_001_filter.yml")

	if err := utils.GetGotExpErr("Test_groups_001_filter:rows", len(tbl), 3); err != nil {
		t.Errorf("%v", err)
		return
	}

	ds := "Com1, grpb10 Com2 * grpa50 * <coM3> * grpa20 *"
	cntstr := getGroupCount(tbl, ds)
	if err := utils.GetGotExpErr("Test_groups_001_filter:group count", cntstr, "10"); err != nil {
		t.Errorf("%v", err)
		return
	}
	//Com1, grpa10 Com2 uniq0001 grpa50 uniq0101 <coM3> uniq0201 grpa20 uniq0301
	ds = "Com1, grpa10 Com2 * grpa50 * <coM3> * grpa20 *"
	cntstr = getGroupCount(tbl, ds)
	if err := utils.GetGotExpErr("Test_groups_001_filter:group count", cntstr, "0"); err != nil {
		t.Errorf("%v", err)
		return
	}

	//Com1, grpe10 Com2 uniq0041 grpa50 uniq0141 <coM3> uniq0241 grpc20 uniq0341
	ds = "Com1, grpe10 Com2 * grpa50 * <coM3> * grpc20 *"
	cntstr = getGroupCount(tbl, ds)
	if err := utils.GetGotExpErr("Test_groups_001_filter:group count", cntstr, "0"); err != nil {
		t.Errorf("%v", err)
		return
	}
}
