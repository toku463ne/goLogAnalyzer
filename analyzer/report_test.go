package analyzer

import (
	"fmt"
	"testing"
)

func Test_report(t *testing.T) {
	if _, err := initTestDir("reportest"); err != nil {
		t.Errorf("%v", err)
		return
	}

	r, err := newReport("testdata/report/config_test.json", 0)
	if err != nil {
		t.Errorf("%v", err)
		return
	}
	if err := r.run(); err != nil {
		t.Errorf("%v", err)
		return
	}

	m := map[string][]string{
		"test": {"b"},
	}
	gotstr := r.insertHtmlTag("This is tESt", m)
	if err := getGotExpErr("replace test",
		gotstr,
		"This is <b>tESt</b>"); err != nil {
		t.Errorf("%+v", err)
		return
	}
	gotstr = r.insertHtmlTag("This is tESt. Not tesT2. But teste this.", m)
	if err := getGotExpErr("replace test",
		gotstr,
		"This is <b>tESt</b>. Not <b>tesT</b>2. But <b>test</b>e this."); err != nil {
		t.Errorf("%+v", err)
		return
	}
	path := fmt.Sprintf("%s/%s.html",
		r.conf.Children[0].Categories[0].reportDir,
		r.conf.Children[0].Categories[0].Name)
	if !PathExist(path) {
		t.Errorf("%s does not exists", path)
		return
	}
	path = fmt.Sprintf("%s/%s.html",
		r.conf.Children[0].Categories[1].reportDir,
		r.conf.Children[0].Categories[1].Name)
	if PathExist(path) {
		t.Errorf("%s does not exists", path)
		return
	}
	path = fmt.Sprintf("%s/%s.html",
		r.conf.Children[1].reportDir,
		r.conf.Children[1].Name)
	if !PathExist(path) {
		t.Errorf("%s does not exists", path)
		return
	}
	path = fmt.Sprintf("%s/%s.html",
		r.confGroups.g["test333"][0].reportDir,
		r.confGroups.g["test333"][0].Name,
	)
	if !PathExist(path) {
		t.Errorf("%s does not exists", path)
		return
	}
}
