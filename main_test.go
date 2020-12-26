package main

import (
	"os"
	"testing"
)

func Test_main(t *testing.T) {
	rootDir := "c:\\Users\\kot\\loganal\\realtest\\db"
	linesInBlock := "1000"
	maxBlocks := "10"
	logPathRegex := "c:\\Users\\kot\\loganal\\realtest\\test.log*"
	os.Args = []string{"test", "clean", "-d", rootDir}
	main()

	//loganal rar [-f LOGPATH] [-d DATADIR] [-g GAPVALUE] [-v] [-s SEARCH_KEYS] [-x EXCLUDE_KEYS]
	os.Args = []string{"test", "rar", "-f", logPathRegex,
		"-d", rootDir, "-linesInBlock", linesInBlock, "-maxBlock", maxBlocks}
	main()
}

func Test_main2(t *testing.T) {
	logPathRegex := "c:\\Users\\kot\\loganal\\realtest\\test.log"

	//loganal rar [-f LOGPATH] [-d DATADIR] [-g GAPVALUE] [-v] [-s SEARCH_KEYS] [-x EXCLUDE_KEYS]
	os.Args = []string{"test", "frq", "-f", logPathRegex, "-m", "100"}
	main()
}
