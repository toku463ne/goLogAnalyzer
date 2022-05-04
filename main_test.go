package main

import (
	"os"
	"testing"
)

// logan run -f '/mnt/c/Users/kot/loganal/fw03/NFPFW003*'
func Test_main1(t *testing.T) {
	rootDir := "/mnt/c/Users/kot/Documents/loganal/fw03/data"
	logPathRegex := "/mnt/c/Users/kot/loganal/fw03/NFPFW003.log*"
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
