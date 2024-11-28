package main

import (
	"goLogAnalyzer/pkg/utils"
	"os"
	"testing"
)

func runGroups(t *testing.T, testName, config string, extraArgs []string) [][]string {
	testDir, err := initTestDir(t, testName)
	if err != nil {
		t.Errorf("%v", err)
		return nil
	}

	// run test
	os.Args = []string{"logan", "groups", "-c", config, "-o", testDir}
	if extraArgs != nil && len(extraArgs) > 0 {
		os.Args = append(os.Args, extraArgs...)
	}
	main()

	_, records, err := utils.ReadCsv(testDir+"/logGroups.csv", ',', false)
	if err != nil {
		t.Errorf("%v", err)
		return nil
	}
	return records
}

func getGroupCount(records [][]string, displayString string) string {
	for _, row := range records {
		if row[3] == displayString {
			return row[1]
		}
	}
	return "0"
}

func Test_groups_001_filter(t *testing.T) {
	tbl := runGroups(t, "Test_groups_001_filter", "../../testdata/loganal/groups_001_filter.yml", nil)

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

func Test_groups_002_kw_igw(t *testing.T) {
	tbl := runGroups(t, "Test_groups_002_kw_igw", "../../testdata/loganal/groups_002_kw_igw.yml", nil)

	if err := utils.GetGotExpErr("Test_groups_002_kw_igw:rows", len(tbl), 4); err != nil {
		t.Errorf("%v", err)
		return
	}

	ds := "Com1, * Com2 * grpa50 * <coM3> * grpb20 *"
	cntstr := getGroupCount(tbl, ds)
	if err := utils.GetGotExpErr("Test_groups_002_kw_igw:group count", cntstr, "20"); err != nil {
		t.Errorf("%v", err)
		return
	}
	ds = "Com1, grpa10 Com2 * grpa50 * <coM3> * * *"
	cntstr = getGroupCount(tbl, ds)
	if err := utils.GetGotExpErr("Test_groups_002_kw_igw:group count", cntstr, "10"); err != nil {
		t.Errorf("%v", err)
		return
	}

}

func Test_groups_003_regex(t *testing.T) {
	tbl := runGroups(t, "Test_groups_003_regex", "../../testdata/loganal/groups_003_regex.yml", nil)

	if err := utils.GetGotExpErr("Test_groups_003_regex:rows", len(tbl), 6); err != nil {
		t.Errorf("%v", err)
		return
	}

	ds := "Com1, grpb10 Com2 * * * <coM3> * grpa20 *"
	cntstr := getGroupCount(tbl, ds)
	if err := utils.GetGotExpErr("Test_groups_003_regex:group count", cntstr, "10"); err != nil {
		t.Errorf("%v", err)
		return
	}
	ds = "Com1, grpa10 Com2 (uniq)0010 * (uniq)0110 <coM3> (uniq)0210 grpa20 (uniq)0310"
	cntstr = getGroupCount(tbl, ds)
	if err := utils.GetGotExpErr("Test_groups_003_regex:group count", cntstr, "1"); err != nil {
		t.Errorf("%v", err)
		return
	}

}

func Test_groups_004_outfilter(t *testing.T) {
	extraArgs := []string{"-s", "grpa20"}
	tbl := runGroups(t, "Test_groups_004_outfilter", "../../testdata/loganal/groups_004_outfilter.yml", extraArgs)

	if err := utils.GetGotExpErr("Test_groups_004_outfilter:rows", len(tbl), 2); err != nil {
		t.Errorf("%v", err)
		return
	}

	extraArgs = []string{"-s", "grpa20", "-x", "grpa10"}
	tbl = runGroups(t, "Test_groups_004_outfilter", "../../testdata/loganal/groups_004_outfilter.yml", extraArgs)

	if err := utils.GetGotExpErr("Test_groups_004_outfilter:rows", len(tbl), 1); err != nil {
		t.Errorf("%v", err)
		return
	}
}
