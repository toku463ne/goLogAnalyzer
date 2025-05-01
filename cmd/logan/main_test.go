package main

import (
	"goLogAnalyzer/pkg/utils"
	"os"
	"strings"
	"testing"
)

func Test_main_groups(t *testing.T) {
	testDir, err := utils.InitTestDir("Test_main_groups")
	if err != nil {
		t.Errorf("%v", err)
		return
	}
	dataDir := testDir + "/data"
	logPathRegex := "../../testdata/loganal/sample50_1.log"
	os.Args = []string{"logan", "groups", "-f", logPathRegex, "-d", dataDir, "-o", dataDir, "-N", "3"}
	main()

	header, records, err := utils.ReadCsv(dataDir+"/logGroups.csv", ',', false)
	if err != nil {
		t.Errorf("%v", err)
		return
	}

	if err := utils.GetGotExpErr("len(header)", len(header), 4); err != nil {
		t.Errorf("%v", err)
		return
	}
	if err := utils.GetGotExpErr("len(records)", len(records), 3); err != nil {
		t.Errorf("%v", err)
		return
	}

	os.Args = []string{"logan", "clean", "-d", dataDir, "-silent"}
	main()

}

func Test_main_config(t *testing.T) {
	testDir, err := utils.InitTestDir("Test_main_config")
	if err != nil {
		t.Errorf("%v", err)
		return
	}
	dataDir := testDir + "/data"
	config := "../../testdata/loganal/sample.yml.j2"
	os.Args = []string{"logan", "history", "-c", config, "-o", dataDir, "-N", "3"}
	main()

	_, records, err := utils.ReadCsv(dataDir+"/history.csv", ',', true)
	if err != nil {
		t.Errorf("%v", err)
		return
	}
	//if err := utils.GetGotExpErr("len(header)", len(header), 3); err != nil {
	//	t.Errorf("%v", err)
	//	return
	//}
	if err := utils.GetGotExpErr("len(records)", len(records), 5); err != nil {
		t.Errorf("%v", err)
		return
	}

	if _, records, err = utils.ReadCsv(dataDir+"/logGroups_last.csv", ',', false); err != nil {
		t.Errorf("%v", err)
		return
	}
	if err := utils.GetGotExpErr("len(records)", len(records), 3); err != nil {
		t.Errorf("%v", err)
		return
	}

	os.Args = []string{"logan", "history", "-c", config, "-o", dataDir, "-b", "20", "-m", "0.5"}
	main()
	_, records, err = utils.ReadCsv(dataDir+"/history.csv", ',', false)
	if err != nil {
		t.Errorf("%v", err)
		return
	}
	if err := utils.GetGotExpErr("len(records)", len(records), 4); err != nil {
		t.Errorf("%v", err)
		return
	}

	os.Args = []string{"logan", "clean", "-c", config, "-silent"}
	main()
}

func Test_testmode(t *testing.T) {
	baseDir, err := utils.InitTestDir("Test_testmode")
	if err != nil {
		t.Errorf("%v", err)
		return
	}
	errFile := baseDir + "/errors.log"
	dataDir := baseDir + "/data"

	config := "../../testdata/loganal/sample_test.yml.j2"
	os.Args = []string{"logan", "test", "-c", config,
		"-line", "2024-10-02T06:00:00] Com1, grpd10 Com2 (uniq)0031 grpa50 (uniq)0131 <coM3> (uniq)0231 grpb20 (uniq)0331",
		"-anallog", errFile, "-silent"}
	main()
	errNo, err := utils.CountFileLines(errFile)
	if err != nil {
		t.Errorf("%v", err)
		return
	}
	if err := utils.GetGotExpErr("errors", errNo, 0); err != nil {
		t.Errorf("%v", err)
		return
	}

	if err := utils.GetGotExpErr("dataDir existance", utils.PathExist(dataDir), false); err != nil {
		t.Errorf("%v", err)
		return
	}

	os.Args = []string{"logan", "feed", "-c", config}
	main()

	if err := utils.GetGotExpErr("dataDir existance", utils.PathExist(dataDir), true); err != nil {
		t.Errorf("%v", err)
		return
	}

	os.Args = []string{"logan", "test", "-c", config,
		"-line", "2024-10-02T06:00:00] Com1, grpd10 Com2 (uniq)0031 grpa50 (uniq)0131 <coM3> (uniq)0231 grpb20 (uniq)0331",
		"-anallog", errFile, "-silent"}
	main()

	errNo, err = utils.CountFileLines(errFile)
	if err != nil {
		t.Errorf("%v", err)
		return
	}
	if err := utils.GetGotExpErr("errors", errNo, 0); err != nil {
		t.Errorf("%v", err)
		return
	}
}

func Test_numbers(t *testing.T) {
	testDir, err := utils.InitTestDir("Test_numbers")
	if err != nil {
		t.Errorf("%v", err)
		return
	}
	dataDir := testDir + "/data"
	config := "../../testdata/loganal/sample_numbers.yml.j2"
	os.Args = []string{"logan", "groups", "-c", config, "-o", dataDir, "-N", "3"}
	main()

	_, records, err := utils.ReadCsv(dataDir+"/logGroups.csv", ',', false)
	if err != nil {
		t.Errorf("%v", err)
		return
	}
	line := "* Com1, grpf10 Com2 (uniq)* grpb50 (uniq)* <coM3> (uniq)* * grpc20 (uniq)*"
	if err := utils.GetGotExpErr("line", records[0][3], line); err != nil {
		t.Errorf("%v", err)
		return
	}
}

func Test_numbers2(t *testing.T) {
	testDir, err := utils.InitTestDir("Test_numbers2")
	if err != nil {
		t.Errorf("%v", err)
		return
	}
	dataDir := testDir + "/data"
	config := "../../testdata/loganal/sample_numbers2.yml.j2"
	os.Args = []string{"logan", "groups", "-c", config, "-o", dataDir, "-N", "3"}
	main()

	_, records, err := utils.ReadCsv(dataDir+"/logGroups.csv", ',', false)
	if err != nil {
		t.Errorf("%v", err)
		return
	}
	line := "246 Com1, grpf10 Com2 (uniq)* grpb50 (uniq)* <coM3> (uniq)* 1234 grpc20 (uniq)*"
	if err := utils.GetGotExpErr("line", records[0][3], line); err != nil {
		t.Errorf("%v", err)
		return
	}
}

func Test_netscreen(t *testing.T) {
	dataDir := os.Getenv("HOME") + "/logantests/Test_netscreen/data"
	utils.RemoveDirectory(dataDir)

	config := "../../testdata/loganal/netscreen.yml.j2"

	os.Args = []string{"logan", "groups", "-c", config, "-o", dataDir}
	main()

	_, records, err := utils.ReadCsv(dataDir+"/logGroups.csv", ',', false)
	if err != nil {
		t.Errorf("%v", err)
		return
	}
	if err := utils.GetGotExpErr("len(records)", len(records), 6); err != nil {
		t.Errorf("%v", err)
		return
	}
	if err := utils.GetGotExpErr("records[0][3]", records[0][3], "DNS has been refreshed."); err != nil {
		t.Errorf("%v", err)
		return
	}

	if err := utils.GetGotExpErr("records[0][1]", records[0][1], "5"); err != nil {
		t.Errorf("%v", err)
		return
	}
}

func Test_no_datadir(t *testing.T) {
	testDir, err := utils.InitTestDir("Test_no_datadir")
	if err != nil {
		t.Errorf("%v", err)
		return
	}

	// Prepare output redirection
	outputFile := testDir + "/output.log"
	file, err := os.Create(outputFile)
	if err != nil {
		t.Errorf("Failed to create output file: %v", err)
		return
	}
	defer file.Close()

	// Save original Stdout and Stderr
	origStdout := os.Stdout
	origStderr := os.Stderr

	// Redirect Stdout and Stderr
	os.Stdout = file
	os.Stderr = file

	// Run the main function
	logPath := "../../testdata/loganal/sample50_1.log"
	os.Args = []string{"logan", "groups", "-f", logPath}
	main()

	// Restore original Stdout and Stderr
	os.Stdout = origStdout
	os.Stderr = origStderr

	// Optionally check the output file content
	content, err := os.ReadFile(outputFile)
	if err != nil {
		t.Errorf("Failed to read output file: %v", err)
		return
	}
	// Convert content to a string for processing
	output := string(content)

	// Extract lines that look like group IDs
	lines := strings.Split(output, "\n")
	var groupIDCount int
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "173301083") {
			groupIDCount++
		}
	}

	// Assert that there are 5 group IDs
	if groupIDCount != 5 {
		t.Errorf("Expected 5 group IDs, but found %d", groupIDCount)
	}
}

func Test_others_001_parseline_test(t *testing.T) {
	_, err := initTestDir(t, "others_001_parseline_test")
	if err != nil {
		t.Errorf("%v", err)
		return
	}
	config := "../../testdata/loganal/others_001_parseline_test.yml"
	line := "10th, 02:14:49.143+0900 TBLV1 DAO : CTBCAFLogDao::Sync reseting GzipPendingEvent"

	// run test
	os.Args = []string{"logan", "test", "-c", config, "-line", line}
	main()
}

func Test_others_002_anomaly_test(t *testing.T) {
	_, err := initTestDir(t, "others_002_anomaly_test")
	if err != nil {
		t.Errorf("%v", err)
		return
	}
	config := "../../testdata/loganal/anomaly.yml.j2"
	// run test
	os.Args = []string{"logan", "stats", "-c", config}
	main()
}

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

func Test_real(t *testing.T) {
	config := "/home/ubuntu/tests/sophos/SOPHOS-01.yml"
	//os.Args = []string{"logan", "clean", "-c", config, "-silent"}
	//main()

	os.Args = []string{"logan", "history", "-c", config, "-o", "/tmp/out2", "-asc", "-N", "10"}
	main()
}
