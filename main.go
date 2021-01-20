package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"

	"./analyzer"
)

var (
	clnFlag = flag.NewFlagSet("clean", flag.ExitOnError)
	frqFlag = flag.NewFlagSet("frq", flag.ExitOnError)
	rarFlag = flag.NewFlagSet("rar", flag.ExitOnError)
	stsFlag = flag.NewFlagSet("stats", flag.ExitOnError)

	rootDirDesc   = "Directory to save the analyzation data"
	rarRootDir    = rarFlag.String("d", "", rootDirDesc)
	pathRegexDesc = "Log file(regex) to analyze. Supports data from pipe"
	rarPathRegex  = rarFlag.String("f", "", pathRegexDesc)
	filterReDesc  = "key word to search"
	rarFilterRe   = rarFlag.String("s", "", filterReDesc)
	xfilterReDesc = "key word to exclude"
	rarXFilterRe  = rarFlag.String("x", "", xfilterReDesc)
	gapDesc       = `Gap rate from average
		Log records with rarity score whose gap if higher that this value will be showed.`
	rarGap           = rarFlag.Float64("g", 0.0, gapDesc)
	showDebugDesc    = "show debug logs"
	rarShowDebug     = rarFlag.Bool("v", false, showDebugDesc)
	forceSaveDb      = "Update the data without asking"
	rarForceSaveDb   = rarFlag.Bool("save", false, forceSaveDb)
	linesInBlockDesc = "lines in block"
	rarLinesInBlock  = rarFlag.Int("linesInBlock", -1, linesInBlockDesc)
	maxBlockDesc     = "max blocks"
	rarMaxBlock      = rarFlag.Int("maxBlock", -1, maxBlockDesc)
	maxLinesDesc     = "max lines to process"
	rarMaxLines      = rarFlag.Int("n", 0, maxLinesDesc)

	filePathDesc   = "Text file to analyze"
	frqPath        = frqFlag.String("f", "", filePathDesc)
	frqFilterRe    = frqFlag.String("s", "", filterReDesc)
	frqXFilterRe   = frqFlag.String("x", "", xfilterReDesc)
	minSupportDesc = "min support"
	frqMinSupport  = frqFlag.Int("m", 0, minSupportDesc)
	frqDebug       = frqFlag.Bool("v", false, showDebugDesc)

	clnRootDir = clnFlag.String("d", "", rootDirDesc)
	clnDebug   = clnFlag.Bool("v", false, showDebugDesc)

	stsRootDir = stsFlag.String("d", "", rootDirDesc)

	usageTxt = `Usage of logan:  
logan [rar|clean|stats|test|frq] OPTIONS  

logan -help:
	Shows this help

logan rar:
	Calculate rarity score of each log records and show the "rare" records.
	Run "logan rar -help" for details.

logan clean:
	Cleanups all statistics data.
	Run "logan clean -help" for details.  

logan stats:
	Shows the statistics according the data in the last execution.
	Run "logan stats -help" for details.

logan test:
	Shows all log records with the score gap.
	Run "logan test -help" for details.

logan frq:
	Shows the closed frequent itemsets order by the supports.
	Only calculate at most 10000 records.
`
)

func clean(isDebug bool) error {
	clnFlag.Parse(os.Args[2:])
	var err error
	if isDebug {
		err = analyzer.CleanupDBProc(*clnRootDir, *clnDebug)
	} else {
		err = analyzer.CleanupDB(*clnRootDir, *clnDebug)
	}
	return err
}

func stats() error {
	stsFlag.Parse(os.Args[2:])
	if *stsRootDir == "" {
		return fmt.Errorf("rootDir is must")
	}
	if !analyzer.PathExist(*stsRootDir) {
		return fmt.Errorf("%s does not exist", *stsRootDir)
	}

	err := analyzer.RarStats(*stsRootDir)
	return err
}

func rar(verbose, isDebug bool) error {
	rarFlag.Parse(os.Args[2:])

	forceSaveDb1 := *rarForceSaveDb
	saveDb := false
	if forceSaveDb1 == false {
		if *rarRootDir != "" {
			if analyzer.PathExist(*rarRootDir) {
				fmt.Printf("Update data on %s? (y/n) (default 'no') ", *rarRootDir)
				stdin := bufio.NewScanner(os.Stdin)
				stdin.Scan()
				k := stdin.Text()
				if strings.ToLower(k) != "y" && strings.ToLower(k) != "yes" {
					fmt.Printf("input='%s' will not update %s\n", k, *rarRootDir)
					saveDb = false
				} else {
					fmt.Printf("input='%s' will update %s\n", k, *rarRootDir)
					saveDb = true
				}
			} else {
				saveDb = true
			}
		}
	} else {
		if *rarRootDir != "" {
			saveDb = true
		}
	}

	if isDebug {
		if err := analyzer.RunRarProc(*rarPathRegex, *rarRootDir,
			*rarFilterRe, *rarXFilterRe,
			*rarGap,
			*rarMaxLines, *rarLinesInBlock, *rarMaxBlock,
			*rarShowDebug, verbose, saveDb); err != nil {
			return err
		}
	} else {
		if err := analyzer.RunRar(*rarPathRegex, *rarRootDir,
			*rarFilterRe, *rarXFilterRe,
			*rarGap,
			*rarMaxLines, *rarLinesInBlock, *rarMaxBlock,
			*rarShowDebug, verbose, saveDb); err != nil {
			return err
		}
	}

	return nil
}

func frq() error {
	frqFlag.Parse(os.Args[2:])

	if err := analyzer.RunFrq(*frqPath, *frqMinSupport,
		*frqFilterRe, *frqXFilterRe, *frqDebug); err != nil {
		return err
	}
	return nil
}

func testGap() error {
	testGapFlg := flag.NewFlagSet("testGap", flag.ExitOnError)
	pathRegex := testGapFlg.String("f", "", "log file path")
	testGapFlg.Parse(os.Args[2:])
	if err := analyzer.TestGap(*pathRegex); err != nil {
		return err
	}
	return nil
}

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "%s\n", usageTxt)
	}
	if len(os.Args) < 2 {
		return
	}
	var err error
	err = nil
	opt := os.Args[1]
	switch opt {
	case "clean":
		err = clean(false)
	case "rar":
		err = rar(false, true)
	case "stats":
		err = stats()
	case "test":
		err = rar(true, false)
	case "frq":
		err = frq()
	//case "testGap":
	//	err = testGap()
	default:
		flag.Usage()
	}
	if err != nil {
		fmt.Printf("%+v\n", err)
	}
}

func debug() {
	if len(os.Args) < 2 {
		flag.Usage()
		return
	}
	var err error
	err = nil
	opt := os.Args[1]
	switch opt {
	case "clean":
		err = clean(true)
	case "rar":
		err = rar(false, true)
	case "test":
		err = rar(true, true)
	default:
		flag.Usage()
	}
	if err != nil {
		fmt.Printf("%+v\n", err)
	}
}
