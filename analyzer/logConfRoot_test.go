package analyzer

import "testing"

func Test_newLogConfRoot(t *testing.T) {
	lcr, err := newLogConfRoot("testdata/report/config_test.json")
	if err != nil {
		t.Errorf("%v", err)
		return
	}
	if err := getGotExpErr("datadir", lcr.RootDir, "/home/administrator/loganal/reportest"); err != nil {
		t.Errorf("%v", err)
		return
	}
	if err := getGotExpErr("keyempasize", lcr.KeyEmphasize["333"][0],
		"font color='red'"); err != nil {
		t.Errorf("%v", err)
		return
	}
	if err := getGotExpErr("template", lcr.Templates["test333"].Search, "test334"); err != nil {
		t.Errorf("%v", err)
		return
	}
	if err := getGotExpErr("child", lcr.Children[0].LogPath,
		"testdata/report/samples/a/A.log*"); err != nil {
		t.Errorf("%v", err)
		return
	}
	if err := getGotExpErr("child.dataDir", lcr.Children[0].dataDir,
		"/home/administrator/loganal/reportest/logA"); err != nil {
		t.Errorf("%v", err)
		return
	}
	if err := getGotExpErr("child.Name", lcr.Children[0].Name, "logA"); err != nil {
		t.Errorf("%v", err)
		return
	}
	if err := getGotExpErr("category.dataDir", lcr.Children[0].Categories[0].dataDir,
		"/home/administrator/loganal/reportest/logA"); err != nil {
		t.Errorf("%v", err)
		return
	}
	if err := getGotExpErr("category.reportDir", lcr.Children[0].Categories[0].reportDir,
		"/tmp/reportest/logA"); err != nil {
		t.Errorf("%v", err)
		return
	}
	if err := getGotExpErr("category.Name", lcr.Children[0].Categories[0].Name, "a_test333"); err != nil {
		t.Errorf("%v", err)
		return
	}
	if err := getGotExpErr("category.Search", lcr.Children[0].Categories[0].Search, "test334"); err != nil {
		t.Errorf("%v", err)
		return
	}
	if err := getGotExpErr("child.GroupName", lcr.Children[0].Categories[0].GroupNames[0], "test333"); err != nil {
		t.Errorf("%v", err)
		return
	}
}
