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

	header, records, err := utils.ReadCsv(dataDir+"/history.csv", ',', false)
	if err != nil {
		t.Errorf("%v", err)
		return
	}
	if err := utils.GetGotExpErr("len(header)", len(header), 3); err != nil {
		t.Errorf("%v", err)
		return
	}
	if err := utils.GetGotExpErr("len(records)", len(records), 4); err != nil {
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
		if strings.HasPrefix(line, "1730671576") {
			groupIDCount++
		}
	}

	// Assert that there are 5 group IDs
	if groupIDCount != 5 {
		t.Errorf("Expected 5 group IDs, but found %d", groupIDCount)
	}
}

func Test_real(t *testing.T) {
	config := "/home/administrator/tests/sbc/gateway.yml"
	//os.Args = []string{"logan", "clean", "-c", config, "-silent"}
	//main()

	os.Args = []string{"logan", "history", "-c", config, "-o", "/tmp/out2", "-asc", "-N", "10"}
	main()
}
