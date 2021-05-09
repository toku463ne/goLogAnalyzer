package main

import (
	"os"
	"testing"
)

func Test_main(t *testing.T) {
	rootDir := "c:\\Users\\kot\\loganal\\realtest\\test"
	linesInBlock := "1000"
	maxBlocks := "100"
	logPathRegex := "c:\\Users\\kot\\loganal\\realtest\\test.log*"
	//rarityCountRate := "0.0001"
	os.Args = []string{"test", "clean", "-d", rootDir}
	debug()

	os.Args = []string{"test", "rar", "-f", logPathRegex,
		"-d", rootDir, "-linesInBlock", linesInBlock,
		"-maxBlock", maxBlocks, "-save"}

	debug()

	//os.Args = []string{"test", "rar", "-f", logPathRegex,
	//	"-d", rootDir, "-linesInBlock", linesInBlock,
	//	"-maxBlock", maxBlocks, "-n", "1000", "-a", "-g", "0.4"}

	//debug()
}

func Test_main2(t *testing.T) {
	rootDir := "c:\\Users\\kot\\loganal\\realtest5\\data"
	linesInBlock := "10000"
	maxBlocks := "1000"
	logPathRegex := "c:\\Users\\kot\\loganal\\realtest5\\test.log*"
	//rarityCountRate := "0.0001"
	os.Args = []string{"test", "clean", "-d", rootDir}
	debug()

	os.Args = []string{"test", "rar", "-f", logPathRegex,
		"-d", rootDir, "-linesInBlock", linesInBlock,
		"-maxBlock", maxBlocks, "-save", "-n", "20000"}

	debug()
}

func Test_main3(t *testing.T) {
	rootDir := "c:\\Users\\kot\\loganal\\realtest4\\data"
	os.Args = []string{"test", "topN",
		"-d", rootDir, "-start", "2020-12-06"}

	topN()
}

func Test_updSts(t *testing.T) {
	rootDir := "c:\\Users\\kot\\loganal\\realtest4\\data"
	linesInBlock := "1000"
	maxBlocks := "100"
	logPathRegex := "c:\\Users\\kot\\loganal\\realtest4\\test4.log*"
	//rarityCountRate := "0.0001"
	os.Args = []string{"test", "clean", "-d", rootDir}
	debug()

	os.Args = []string{"test", "rar", "-f", logPathRegex,
		"-d", rootDir, "-linesInBlock", linesInBlock,
		"-maxBlock", maxBlocks, "-save", "-n", "20000"}

	debug()

	os.Args = []string{"test", "stats",
		"-d", rootDir, "-u"}

	debug()
}
