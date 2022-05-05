package main

import (
	"os"
	"testing"
)

// logan run -f '/mnt/c/Users/kot/loganal/fw03/NFPFW003*'
func Test_main1(t *testing.T) {
	rootDir := "/mnt/c/Users/kot/Documents/loganal/ipx/data"
	logPathRegex := "/mnt/c/Users/kot/loganal/ipx/ipx/IPXFW001.log*"
	os.Args = []string{"test", "clean", "-d", rootDir}
	main()

	os.Args = []string{"test", "run", "-d", rootDir, "-f", logPathRegex, "-save", "yes"}
	main()
}
func Test_main2(t *testing.T) {
	rootDir := "/mnt/c/Users/kot/loganal/realtest3/data"
	os.Args = []string{"test", "topN", "-d", rootDir}
	main()
}

func Test_report(t *testing.T) {
	jconf := "configsample.json"
	os.Args = []string{"test", "report", "-c", jconf}
	main()
}

func Test_main4(t *testing.T) {
	datadir := "/mnt/c/Users/kot/Documents/loganal/test2root/test"
	os.Args = []string{"test", "topN", "-d", datadir,
		"-s", "(?i)(error|fatal|crit|fail|down|panic|timeout|warn)"}
	main()
}

func Test_reportipx(t *testing.T) {
	jconf := "ipx.json"
	os.Args = []string{"test", "report", "-c", jconf}
	main()
}
