package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/toku463ne/goLogAnalyzer/analyzer"
)

const (
	cDefaultBlockSize     = 10000
	cDefaultMaxBlocks     = 100
	cDefaultMaxItemBlocks = 1000
	cMinGapToRecord       = 1.5
	cDefaultHistSize      = 5
)

var (
	clnFlag    = flag.NewFlagSet("clean", flag.ExitOnError)
	rarFlag    = flag.NewFlagSet("rar", flag.ExitOnError)
	topNFlag   = flag.NewFlagSet("topN", flag.ExitOnError)
	stsFlag    = flag.NewFlagSet("stats", flag.ExitOnError)
	reportFlag = flag.NewFlagSet("report", flag.ExitOnError)

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
	rarGap               = rarFlag.Float64("g", cMinGapToRecord, gapDesc)
	forceSaveDb          = "Update the data without asking"
	rarForceSaveDb       = rarFlag.Bool("save", false, forceSaveDb)
	linesInBlockDesc     = "lines in block"
	rarLinesInBlock      = rarFlag.Int("linesInBlock", cDefaultBlockSize, linesInBlockDesc)
	maxBlockDesc         = "max blocks"
	rarMaxBlock          = rarFlag.Int("maxBlock", cDefaultMaxBlocks, maxBlockDesc)
	maxItemBlockDesc     = "max blocks for items"
	rarMaxItemBlock      = rarFlag.Int("maxItemBlock", cDefaultMaxItemBlocks, maxItemBlockDesc)
	maxLinesDesc         = "max lines to process"
	rarMaxLines          = rarFlag.Int("n", 0, maxLinesDesc)
	recordsToShowDesc    = "Top N rare records to show"
	rarTopnRecordsToShow = rarFlag.Int("silent", 10, recordsToShowDesc)

	clnRootDir = clnFlag.String("d", "", rootDirDesc)

	topnRootDir           = topNFlag.String("d", "", rootDirDesc)
	topnRecordsToShow     = topNFlag.Int("n", 10, recordsToShowDesc)
	topnFilterRe          = topNFlag.String("s", "", filterReDesc)
	topnXFilterRe         = topNFlag.String("x", "", xfilterReDesc)
	startDateDesc         = "Start date to collect stats %Y-%m-%d format"
	topnStartDate         = topNFlag.String("start", "", startDateDesc)
	endDateDesc           = "End date to collect stats %Y-%m-%d format"
	topnEndDate           = topNFlag.String("end", "", endDateDesc)
	topnShowItemCountDesc = "Show score of items in the log record"
	topnShowItemCount     = topNFlag.Bool("v", false, topnShowItemCountDesc)
	topnMinScore          = topNFlag.Float64("min", 0, "Minimum score to show")
	topnMaxScore          = topNFlag.Float64("max", 0, "Maximum score to shoe")

	stsRootDir       = stsFlag.String("d", "", rootDirDesc)
	stsRecordsToShow = stsFlag.Int("n", 5, "Number of history to show")

	reportConfig      = reportFlag.String("c", "", "Path of the config file (JSON)")
	reportRecentNdays = reportFlag.Int("n", 0, "Recent N days to show the report")

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
	log.Printf("cleaning up %s", *clnRootDir)
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
	var startEpoch, endEpoch int64
	var err error
	if *topnStartDate != "" {
		startEpoch, err = dateStringToEpoch(*topnStartDate)
		if err != nil {
			return err
		}
	}
	if *topnEndDate != "" {
		endEpoch, err = dateStringToEpoch(*topnEndDate)
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
		*topnRecordsToShow, startEpoch, endEpoch,
		*topnFilterRe, *topnXFilterRe, *topnShowItemCount,
		*topnMinScore, *topnMaxScore)
	return err
}

func rar() error {
	rarFlag.Parse(os.Args[2:])

	forceSaveDb := *rarForceSaveDb
	if !forceSaveDb {
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

	log.Printf("start analyzing %s", *rarPathRegex)

	linesProcessed, err := analyzer.AnalyzeRarity(*rarRootDir, *rarPathRegex,
		*rarFilterRe, *rarXFilterRe,
		*rarGap,
		*rarMaxBlock, *rarMaxItemBlock, *rarLinesInBlock,
		*rarMaxLines, *rarTopnRecordsToShow)

	log.Printf("%d lines processed\n", linesProcessed)

	return err
}

func stats() error {
	stsFlag.Parse(os.Args[2:])
	if *stsRootDir == "" {
		return fmt.Errorf("rootDir is must")
	}
	if !pathExist(*stsRootDir) {
		return fmt.Errorf("%s does not exist", *stsRootDir)
	}

	err := analyzer.RarStats(*stsRootDir, *stsRecordsToShow)
	return err
}

func report() error {
	reportFlag.Parse(os.Args[2:])
	return analyzer.Report(*reportConfig, *reportRecentNdays,
		cMinGapToRecord,
		cDefaultMaxBlocks, cDefaultMaxItemBlocks,
		cDefaultBlockSize, 10, cDefaultHistSize)
}

func main() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

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
	case "report":
		err = report()
	default:
		flag.Usage()
	}
	if err != nil {
		log.Printf("error occurred!\n%+v", err)
	}
}
