package main

import (
	"goLogAnalyzer/pkg/utils"
	"os"
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

	header, records, err := utils.ReadCsv(dataDir + "/logGroups.csv")
	if err != nil {
		t.Errorf("%v", err)
		return
	}

	if err := utils.GetGotExpErr("len(header)", len(header), 3); err != nil {
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

	header, records, err := utils.ReadCsv(dataDir + "/history.csv")
	if err != nil {
		t.Errorf("%v", err)
		return
	}
	if err := utils.GetGotExpErr("len(header)", len(header), 6); err != nil {
		t.Errorf("%v", err)
		return
	}
	if err := utils.GetGotExpErr("len(records)", len(records), 3); err != nil {
		t.Errorf("%v", err)
		return
	}

	os.Args = []string{"logan", "clean", "-c", config, "-silent"}
	main()
}

func Test_real(t *testing.T) {
	config := "/home/administrator/tests/sugcap2/sugcap.yml"
	//os.Args = []string{"logan", "clean", "-c", config, "-silent"}
	//main()

	dataDir := "/tmp/logantest/out"
	os.Args = []string{"logan", "history", "-c", config, "-o", dataDir, "-N", "100"}
	main()
}
