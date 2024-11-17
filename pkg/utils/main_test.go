package utils

import (
	"testing"
	"time"
)

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

func Test_Str2date(t *testing.T) {
	tests := []struct {
		title      string
		format     string
		input      string
		expected   time.Time
		shouldFail bool
	}{
		{
			title:    "Syslog format without year",
			format:   "Jan  2 15:04:05",
			input:    "Nov  1 03:13:26",
			expected: time.Date(time.Now().Year(), 11, 1, 3, 13, 26, 0, time.Local),
		},
		{
			title:    "Time only",
			format:   "15:04:05",
			input:    "12:30:45",
			expected: time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), 12, 30, 45, 0, time.Local),
		},
		{
			title:    "Month-Day",
			format:   "01-02",
			input:    "12-25",
			expected: time.Date(time.Now().Year()-1, 12, 25, 0, 0, 0, 0, time.Local),
		},
		{
			title:    "Month-Day Time",
			format:   "01-02 15:04:05",
			input:    "12-25 08:15:30",
			expected: time.Date(time.Now().Year()-1, 12, 25, 8, 15, 30, 0, time.Local),
		},
		{
			title:    "Day only",
			format:   "02",
			input:    "15",
			expected: time.Date(time.Now().Year(), time.Now().Month(), 15, 0, 0, 0, 0, time.Local),
		},
		{
			title:    "Day and Time",
			format:   "02 15:04:05",
			input:    "15 14:20:10",
			expected: time.Date(time.Now().Year(), time.Now().Month(), 15, 14, 20, 10, 0, time.Local),
		},
		{
			title:      "Invalid format",
			format:     "2006-01-02",
			input:      "invalid-date",
			shouldFail: true,
		},
	}

	for _, test := range tests {
		t.Run(test.title, func(t *testing.T) {
			got, err := Str2date(test.format, test.input)

			if test.shouldFail {
				if err == nil {
					t.Errorf("Expected an error for %s but got none", test.title)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error for %s: %v", test.title, err)
					return
				}

				if err := GetGotExpErr(test.title, got, test.expected); err != nil {
					t.Error(err)
				}
			}
		})
	}
}
