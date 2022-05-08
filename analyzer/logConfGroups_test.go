package analyzer

import "testing"

func Test_newLogConfGroups(t *testing.T) {
	lcr, err := newLogConfRoot("testdata/report/config_test.json")
	if err != nil {
		t.Errorf("%v", err)
		return
	}
	g := newLogConfGroups(lcr)

	if err := getGotExpErr("node1", g.g["test333"][0].Name, "a_test333"); err != nil {
		t.Errorf("%v", err)
		return
	}
	if err := getGotExpErr("node1.BlockSize", g.g["test333"][0].BlockSize, 5); err != nil {
		t.Errorf("%v", err)
		return
	}
	if err := getGotExpErr("node1.MaxItemBlocks", g.g["test333"][0].MaxItemBlocks, 99); err != nil {
		t.Errorf("%v", err)
		return
	}
	if err := getGotExpErr("node2", g.g["test333"][1].Name, "logB"); err != nil {
		t.Errorf("%v", err)
		return
	}
	if err := getGotExpErr("datadir", g.reportDir, "/tmp/reportest"); err != nil {
		t.Errorf("%v", err)
		return
	}
	if err := getGotExpErr("node2.BlockSize", g.g["test333"][1].BlockSize, 5); err != nil {
		t.Errorf("%v", err)
		return
	}
	if err := getGotExpErr("node2.MaxItemBlocks", g.g["test333"][1].MaxItemBlocks, 99); err != nil {
		t.Errorf("%v", err)
		return
	}
}
