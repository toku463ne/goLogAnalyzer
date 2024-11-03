package utils

import "testing"

func Test_Replace(t *testing.T) {
	s := "aaaa 1725173655 1725173655 bbbb"
	target := "1725173655"
	replacement := "*"
	separators := " \t\r\n\"'\\,;[]<>{}=()|:&?/+.!-@"

	s = Replace(s, target, replacement, separators)
	if err := GetGotExpErr("replace1", s, "aaaa * * bbbb"); err != nil {
		t.Errorf("%v", err)
		return
	}

	s = "aaaa [1725173655] 17251736556 bbbb"
	s = Replace(s, target, replacement, separators)
	if err := GetGotExpErr("replace2", s, "aaaa [*] 17251736556 bbbb"); err != nil {
		t.Errorf("%v", err)
		return
	}

	s = "aaaa test@1725173655 test=1725173655;bbbb"
	s = Replace(s, target, replacement, separators)
	if err := GetGotExpErr("replace3", s, "aaaa test@* test=*;bbbb"); err != nil {
		t.Errorf("%v", err)
		return
	}
}
