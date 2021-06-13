package main

import (
	"os"
	"testing"
)

func Test_rar(t *testing.T) {
	rootDir, err := ensureTestDir("main")
	if err != nil {
		t.Errorf("%+v", err)
	}
	linesInblock := "5"
	maxBlocks := "3"
	logPathRegex := "analyzer/inputs/maintest.log"

	os.Args = []string{"logan", "clean", "-d", rootDir}
	main()

	os.Args = []string{"logan", "rar", "-f", logPathRegex,
		"-d", rootDir, "-linesInBlock", linesInblock,
		"-maxBlock", maxBlocks, "-save"}
	main()

	os.Args = []string{"logan", "stats", "-d", rootDir}
	main()

	os.Args = []string{"logan", "stats", "-d", rootDir, "-u"}
	main()

	os.Args = []string{"logan", "rar", "-f", logPathRegex,
		"-d", rootDir, "-linesInBlock", linesInblock,
		"-maxBlock", maxBlocks, "-save"}
	main()

}

func Test_frq(t *testing.T) {
	logPathRegex := "analyzer/inputs/maintest.log"

	os.Args = []string{"logan", "frq", "-f", logPathRegex}
	main()
}
