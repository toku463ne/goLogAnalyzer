package main

import (
	"os"
	"testing"
)

func Test_main1(t *testing.T) {
	//rootDir := "c:\\Users\\kot\\loganal\\fw03\\data"
	//linesInBlock := "100000"
	//maxBlocks := "100"
	//logPathRegex := "c:\\Users\\kot\\loganal\\fw03\\NFPFW003.log*"
	//rarityCountRate := "0.0001"
	//os.Args = []string{"test", "clean", "-d", rootDir}
	//main()

	//os.Args = []string{"test", "rar", "-f", logPathRegex,
	//	"-d", rootDir, "-linesInBlock", linesInBlock,
	//	"-maxBlock", maxBlocks, "-save"}

	//main()

	//os.Args = []string{"test", "topN", "-d", rootDir, "-n", "10"}

	conf := "c:\\Users\\kot\\loganal\\zimconfwin.json"
	os.Args = []string{"test", "report", "-c", conf}
	main()
}
