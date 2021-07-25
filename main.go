package main

import (
	"analyzer/analyzer"
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"
)

const (
	cDefaultBlockSize     = 10000
	cDefaultMaxBlocks     = 10
	cDefaultMaxItemBlocks = 100
	cMinGapToRecord       = 0.5
)

var (
	clnFlag  = flag.NewFlagSet("clean", flag.ExitOnError)
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
	rarGap           = rarFlag.Float64("g", cMinGapToRecord, gapDesc)
	forceSaveDb      = "Update the data without asking"
	rarForceSaveDb   = rarFlag.Bool("save", false, forceSaveDb)
	linesInBlockDesc = "lines in block"
	rarLinesInBlock  = rarFlag.Int("linesInBlock", cDefaultBlockSize, linesInBlockDesc)
	maxBlockDesc     = "max blocks"
	rarMaxBlock      = rarFlag.Int("maxBlock", cDefaultMaxBlocks, maxBlockDesc)
	maxItemBlockDesc = "max blocks for items"
	rarMaxItemBlock  = rarFlag.Int("maxItemBlock", cDefaultMaxItemBlocks, maxItemBlockDesc)
	maxLinesDesc     = "max lines to process"
	rarMaxLines      = rarFlag.Int("n", 0, maxLinesDesc)
	silentDesc       = "Run without message"
	rarSilent        = rarFlag.Bool("silent", false, silentDesc)

	clnRootDir = clnFlag.String("d", "", rootDirDesc)

	topnRootDir       = topNFlag.String("d", "", rootDirDesc)
	recordsToShowDesc = "Top N rare records to show"
	topnRecordsToShow = topNFlag.Int("n", 0, recordsToShowDesc)
	topnFilterRe      = topNFlag.String("s", "", filterReDesc)
	topnXFilterRe     = topNFlag.String("x", "", xfilterReDesc)
	startDateDesc     = "Start date to collect stats %Y-%m-%d format"
	topnStartDate     = topNFlag.String("start", "", startDateDesc)

	usageTxt = `Usage of logan:  
logan [rar|clean|topN|stats] OPTIONS  

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
`
)

func pathExist(path string) bool {
	if _, err := os.Stat(path); err != nil {
		return false
	}
	return true
}

func dateStringToEpoch(date string) (int64, error) {
	t, err := time.Parse(time.RFC3339, date+"T00:00:00Z")
	if err != nil {
		return -1, err
	}
	return t.Unix(), nil
}

func clean() error {
	clnFlag.Parse(os.Args[2:])
	return analyzer.Clean(*clnRootDir)
}

func topN() error {
	topNFlag.Parse(os.Args[2:])
	if *topnRootDir == "" {
		return fmt.Errorf("rootDir is must")
	}
	if !pathExist(*topnRootDir) {
		return fmt.Errorf("%s does not exist", *topnRootDir)
	}
	var startEpoch int64
	var err error
	if *topnStartDate != "" {
		startEpoch, err = dateStringToEpoch(*topnStartDate)
		if err != nil {
			return err
		}
	}
	/*
		rootDir, msg string,
		recordsToShow int, startEpoch int64,
		filterReStr, xFilterReStr string
	*/

	msg := fmt.Sprintf("%d top rare records", *topnRecordsToShow)

	err = analyzer.PrintRarTopN(*topnRootDir, msg,
		*topnRecordsToShow, startEpoch,
		*topnFilterRe, *topnXFilterRe)
	return err
}

func rar() error {
	rarFlag.Parse(os.Args[2:])

	forceSaveDb := *rarForceSaveDb
	if forceSaveDb == false {
		if *rarRootDir != "" {
			if pathExist(*rarRootDir) {
				fmt.Printf("Update data on %s? (y/n) (default 'no') ", *rarRootDir)
				stdin := bufio.NewScanner(os.Stdin)
				stdin.Scan()
				k := stdin.Text()
				if strings.ToLower(k) != "y" && strings.ToLower(k) != "yes" {
					fmt.Printf(`input='%s' will exit here.\n
You can also try to use -clean option to cleanup the database and try again\n`, k)
				} else {
					fmt.Printf("input='%s' will update %s\n", k, *rarRootDir)
				}
			}
		}
	}

	/*
		rootDir, logPathRegex,
		filterStr, xFilterStr string,
		minGapToRecord float64,
		maxBlocks, maxItemBlocks, linesInBlock int,
		linesToProcess int
	*/

	linesProcessed, err := analyzer.AnalyzeRarity(*rarRootDir, *rarPathRegex,
		*rarFilterRe, *rarXFilterRe,
		*rarGap,
		*rarMaxBlock, *rarMaxItemBlock, *rarLinesInBlock,
		*rarMaxLines)

	fmt.Printf("%d lines processed\n", linesProcessed)

	return err
}

func stats() error {
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
		err = clean()
	case "rar":
		err = rar()
	case "stats":
		err = stats()
	case "topN":
		err = topN()
	default:
		flag.Usage()
	}
	if err != nil {
		fmt.Printf("%+v\n", err)
	}
}
