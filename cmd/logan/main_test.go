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
	if err := utils.GetGotExpErr("len(header)", len(header), 6); err != nil {
		t.Errorf("%v", err)
		return
	}
	if err := utils.GetGotExpErr("len(records)", len(records), 3); err != nil {
		t.Errorf("%v", err)
		return
	}

	_, records, err = utils.ReadCsv(dataDir+"/history_sum.csv", ',', false)
	if err != nil {
		t.Errorf("%v", err)
		return
	}
	if err := utils.GetGotExpErr("len(records)", len(records), 2); err != nil {
		t.Errorf("%v", err)
		return
	}
	if err := utils.GetGotExpErr("records[1][2]", records[1][2], "20"); err != nil {
		t.Errorf("%v", err)
		return
	}

	if _, records, err = utils.ReadCsv(dataDir+"/lastMessages.csv", ',', false); err != nil {
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
	if err := utils.GetGotExpErr("len(records)", len(records), 3); err != nil {
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
	if err := utils.GetGotExpErr("len(records)", len(records), 5); err != nil {
		t.Errorf("%v", err)
		return
	}
	if err := utils.GetGotExpErr("records[0][1]", records[0][1], "4"); err != nil {
		t.Errorf("%v", err)
		return
	}
}

func Test_real(t *testing.T) {
	config := "/home/administrator/tests/sbc/g.yml"
	//os.Args = []string{"logan", "clean", "-c", config, "-silent"}
	//main()

	os.Args = []string{"logan", "history", "-c", config, "-o", "/tmp/out2", "-asc", "-N", "10"}
	main()
}
