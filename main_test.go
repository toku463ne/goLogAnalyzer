package main

import (
	"os"
	"testing"
)

func Test_main(t *testing.T) {
	rootDir := "c:\\Users\\kot\\loganal\\realtest\\test"
	linesInBlock := "10000"
	maxBlocks := "100"
	logPathRegex := "c:\\Users\\kot\\loganal\\realtest\\test.log*"
	//rarityCountRate := "0.0001"
	os.Args = []string{"test", "clean", "-d", rootDir}
	main()

	os.Args = []string{"test", "rar", "-f", logPathRegex,
		"-d", rootDir, "-linesInBlock", linesInBlock,
		"-maxBlock", maxBlocks, "-g", "0.0", "-save"}

	main()
}

func Test_main2(t *testing.T) {
	logPathRegex := "c:\\Users\\kot\\loganal\\realtest\\test.log"

	//loganal rar [-f LOGPATH] [-d DATADIR] [-g GAPVALUE] [-v] [-s SEARCH_KEYS] [-x EXCLUDE_KEYS]
	os.Args = []string{"test", "frq", "-f", logPathRegex, "-m", "100", "-v"}
	main()
}

func Test_main3(t *testing.T) {
	rootDir := "c:\\Users\\kot\\loganal\\realtest\\test"
	os.Args = []string{"test", "stats", "-d", rootDir}
	main()
}
