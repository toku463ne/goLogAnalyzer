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

func Test_realtest(t *testing.T) {
	os.Args = []string{"test", "run", "-f", "/mnt/c/Users/kot/loganal/realtest/test.log*"}
	main()
}

func Test_pfsense(t *testing.T) {
	rootDir := "/home/ubuntu/logan/openvpn"
	logPathRegex := "/home/ubuntu/openvpn_logs/pfsense67051_openvpn.log*"
	os.Args = []string{"pfsense", "clean", "-d", rootDir}
	main()

	os.Args = []string{"pfsense", "run", "-d", rootDir, "-f", logPathRegex, "-save", "yes"}
	main()

	logPathRegex = "/home/ubuntu/openvpn_logs_new/new_logs.log"
	os.Args = []string{"pfsense", "monitor", "-d", rootDir, "-f", logPathRegex}
	main()
}
