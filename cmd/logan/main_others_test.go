package main

import (
	"os"
	"testing"
)

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
