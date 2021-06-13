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
	clnFlag  = flag.NewFlagSet("clean", flag.ExitOnError)
	frqFlag  = flag.NewFlagSet("frq", flag.ExitOnError)
	rarFlag  = flag.NewFlagSet("rar", flag.ExitOnError)
	topNFlag = flag.NewFlagSet("topN", flag.ExitOnError)
	stsFlag  = flag.NewFlagSet("stats", flag.ExitOnError)

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
	silentDesc       = "Run without message"
	rarSilent        = rarFlag.Bool("silent", false, silentDesc)

	filePathDesc   = "Text file to analyze"
	frqPath        = frqFlag.String("f", "", filePathDesc)
	frqFilterRe    = frqFlag.String("s", "", filterReDesc)
	frqXFilterRe   = frqFlag.String("x", "", xfilterReDesc)
	minSupportDesc = "min support"
	frqMinSupport  = frqFlag.Int("m", 0, minSupportDesc)
	frqDebug       = frqFlag.Bool("v", false, showDebugDesc)

	stsUpdateDesc = "Update statistics in the saved data"
	stsUpdate     = stsFlag.Bool("u", false, stsUpdateDesc)

	clnRootDir = clnFlag.String("d", "", rootDirDesc)
	clnDebug   = clnFlag.Bool("v", false, showDebugDesc)

	stsRootDir = stsFlag.String("d", "", rootDirDesc)

	topnRootDir       = topNFlag.String("d", "", rootDirDesc)
	recordsToShowDesc = "Top N rare records to show"
	topnRecordsToShow = topNFlag.Int("n", 0, recordsToShowDesc)
	topnFilterRe      = topNFlag.String("s", "", filterReDesc)
	topnXFilterRe     = topNFlag.String("x", "", xfilterReDesc)
	startDateDesc     = "Start date to collect stats %Y-%m-%d format"
	topnStartDate     = topNFlag.String("start", "", startDateDesc)
	endDateDesc       = "End date to collect stats %Y-%m-%d format"
	topnEndDate       = topNFlag.String("end", "", endDateDesc)
	maxScoreDesc      = "Max score to show"
	topnMaxScore      = topNFlag.Float64("maxScore", 0.0, maxScoreDesc)

	usageTxt = `Usage of logan:  
logan [rar|clean|topN|stats|test|frq] OPTIONS  

logan -help:
	Shows this help

logan rar:
	Calculate rarity score of each log records and show the "rare" records.
	Run "logan rar -help" for details.

logan clean:
	Cleanups all statistics data.
	Run "logan clean -help" for details.  

logan stats:
	Shows the statistics according the saved data.
	Run "logan stats -help" for details.

logan topN:
	Shows the top N rare records
	Run "logan topN -help" for details.

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
	var err error
	if *stsUpdate {
		err = analyzer.UpdateStats(*stsRootDir)
	} else {
		err = analyzer.RarStats(*stsRootDir)
	}
	return err
}

func topN() error {
	topNFlag.Parse(os.Args[2:])
	if *topnRootDir == "" {
		return fmt.Errorf("rootDir is must")
	}
	if !analyzer.PathExist(*topnRootDir) {
		return fmt.Errorf("%s does not exist", *topnRootDir)
	}
	var startEpoch int64
	var endEpoch int64
	var err error
	if *topnStartDate != "" {
		startEpoch, err = analyzer.DateStringToEpoch(*topnStartDate)
		if err != nil {
			return err
		}
	}
	if *topnEndDate != "" {
		endEpoch, err = analyzer.DateStringToEpoch(*topnEndDate)
		if err != nil {
			return err
		}
	}

	err = analyzer.RarTopN(*topnRootDir, *topnRecordsToShow,
		*topnFilterRe, *topnXFilterRe, startEpoch, endEpoch, *topnMaxScore)
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
			*rarShowDebug, verbose, saveDb, *rarSilent); err != nil {
			return err
		}
	} else {
		if err := analyzer.RunRar(*rarPathRegex, *rarRootDir,
			*rarFilterRe, *rarXFilterRe,
			*rarGap,
			*rarMaxLines, *rarLinesInBlock, *rarMaxBlock,
			*rarShowDebug, verbose, saveDb, *rarSilent); err != nil {
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
	case "topN":
		err = topN()
	case "test":
		err = rar(true, false)
	case "frq":
		err = frq()
	default:
		flag.Usage()
	}
	if err != nil {
		fmt.Printf("%+v\n", err)
	}
}

func ensureTestDir(testname string) (string, error) {
	userDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	rootDir := fmt.Sprintf("%s/loganal/%s", userDir, testname)
	if _, err := os.Stat(rootDir); os.IsNotExist(err) {
		os.Mkdir(rootDir, 0755)
	}
	return rootDir, nil
}
