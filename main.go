package main

import (
	"bufio"
	"flag"
	"fmt"
	"goLogAnalyzer/analyzer"
	"log"
	"os"
	"strings"
)

const (
	cRootDirDesc   = "Directory to save the analyzation data"
	cPathRegexDesc = "Log file(regex) to analyze. Supports data from pipe"
	cFilterReDesc  = "key word to search"
	cXfilterReDesc = "key word to exclude"
	cGapDesc       = `Gap rate from average
		Log records with rarity score whose gap if higher that this value will be showed.`
	cForceSaveDbDesc       = "Update the data without asking"
	cBlockSizeDesc         = "Max number of lines in a block"
	cMaxBlockDesc          = "max blocks to save logs"
	cMaxItemBlockDesc      = "max blocks to save terms"
	cMaxLinesDesc          = "max lines to process"
	cNRecordsToShowDesc    = "Top N rare records to show"
	cStartDateDesc         = "Start date to collect stats %Y-%m-%d format"
	cEndDateDesc           = "End date to collect stats %Y-%m-%d format"
	cDatetimeStartDesc     = "Start position of datetime in the log starting from 0"
	cDatetimeLayoutDesc    = "Layout of datetime in the log"
	cScoreStyleDesc        = "How to calculate the score.\n 1:simple average\n 2:average of top scoreNSize terms in a record"
	cScoreNSizeDesc        = "How many terms to take into count"
	cTopNShowItemCountDesc = "Show score of items in the log record"
	cTopNMinScoreDesc      = "Minimum score to show"
	cTopNMaxScoreDesc      = "Maximum score to show"
	cRecordsToShowDesc     = "Number of history to show"
	cReportConfigDesc      = "Path of the config file (JSON)"
	cReportRecentNdaysDesc = "Recent N days to show the report"
	cModeblockPerFileDesc  = "If create blocks per files"
	cNItemTopDesc          = "Top N rare terms to display"
	cDebugDesc             = "Run with detailed messages"
)

var (
	clnFlag    = flag.NewFlagSet("clean", flag.ExitOnError)
	runFlag    = flag.NewFlagSet("run", flag.ExitOnError)
	topNFlag   = flag.NewFlagSet("topN", flag.ExitOnError)
	stsFlag    = flag.NewFlagSet("stats", flag.ExitOnError)
	reportFlag = flag.NewFlagSet("report", flag.ExitOnError)

	clnRootDir = clnFlag.String("d", "", cRootDirDesc)

	runRootDir           = runFlag.String("d", "", cRootDirDesc)
	runPathRegex         = runFlag.String("f", "", cPathRegexDesc)
	runGap               = runFlag.Float64("g", analyzer.CDefaultMinGap, cGapDesc)
	runForceSaveDb       = runFlag.Bool("save", false, cForceSaveDbDesc)
	runBlockSize         = runFlag.Int("blockSize", analyzer.CDefaultBlockSize, cBlockSizeDesc)
	runMaxBlock          = runFlag.Int("maxBlock", analyzer.CDefaultNBlocks, cMaxBlockDesc)
	runMaxItemBlock      = runFlag.Int("maxItemBlock", analyzer.CDefaultNItemBlocks, cMaxItemBlockDesc)
	runTopNRecordsToShow = runFlag.Int("n", analyzer.CDefaultTopNToShow, cNRecordsToShowDesc)
	runDatetimeStartPos  = runFlag.Int("dateStart", 0, cDatetimeStartDesc)
	runDatetimeLayout    = runFlag.String("dateLayout", "", cDatetimeLayoutDesc)
	runScoreStyle        = runFlag.Int("scoreStyle", analyzer.CDefaultScoreStyle, cScoreStyleDesc)
	runScoreNSize        = runFlag.Int("scoreNSize", analyzer.CDefaultScoreNSize, cScoreStyleDesc)
	runModeblockPerFile  = runFlag.Bool("blockPerFile", false, cModeblockPerFileDesc)
	runNRareTerms        = runFlag.Int("nRareTerms", analyzer.CDefaultNRareTerms, cNItemTopDesc)
	runDebug             = runFlag.Bool("debug", false, cDebugDesc)

	topNRootDir       = topNFlag.String("d", "", cRootDirDesc)
	topNRecordsToShow = topNFlag.Int("n", 10, cNRecordsToShowDesc)
	topNFilterRe      = topNFlag.String("s", "", cFilterReDesc)
	topNXFilterRe     = topNFlag.String("x", "", cXfilterReDesc)
	topNStartDate     = topNFlag.String("start", "", cStartDateDesc)
	topNEndDate       = topNFlag.String("end", "", cEndDateDesc)
	topNMinScore      = topNFlag.Float64("min", 0, cTopNMinScoreDesc)
	topNMaxScore      = topNFlag.Float64("max", 0, cTopNMaxScoreDesc)
	topNRareTerms     = topNFlag.Int("nRareTerms", analyzer.CDefaultNRareTerms, cNItemTopDesc)

	stsRootDir       = stsFlag.String("d", "", cRootDirDesc)
	stsRecordsToShow = stsFlag.Int("n", 5, cRecordsToShowDesc)

	reportConfig      = reportFlag.String("c", "", cReportConfigDesc)
	reportRecentNdays = reportFlag.Int("n", 0, "Recent N days to show the report")
	reportDebug       = reportFlag.Bool("debug", false, cDebugDesc)

	usageTxt = `Usage of logan:  
logan [rar|clean|topN|stats] OPTIONS  

logan -help:
	Shows this help

logan run:
	Calculate rarity score of each log records and show the "rare" records.
	Run "logan rar -help" for details.

logan report:
	Reads params from json config file.
	Run "logan report -help" for details.

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

func clean() error {
	log.Printf("cleaning up %s", *clnRootDir)
	clnFlag.Parse(os.Args[2:])
	return analyzer.Clean(*clnRootDir)
}

func run() error {
	runFlag.Parse(os.Args[2:])
	analyzer.InitLog(*runRootDir)
	analyzer.IsDebug = *runDebug
	forceSaveDb := *runForceSaveDb
	if !forceSaveDb {
		if *runRootDir != "" {
			if analyzer.PathExist(*runRootDir) {
				fmt.Printf("Update data on %s? (y/n) (default 'no') ", *runRootDir)
				stdin := bufio.NewScanner(os.Stdin)
				stdin.Scan()
				k := stdin.Text()
				if strings.ToLower(k) != "y" && strings.ToLower(k) != "yes" {
					fmt.Printf(`input='%s' will exit here.\n
You can also try to use -clean option to cleanup the database and try again\n`, k)
				} else {
					fmt.Printf("input='%s' will update %s\n", k, *runRootDir)
				}
			}
		}
	}

	log.Printf("start analyzing %s", *runPathRegex)

	c := analyzer.NewAnalConf(*runRootDir)
	c.LogPathRegex = *runPathRegex
	c.BlockSize = *runBlockSize
	c.MaxBlocks = *runMaxBlock
	c.MaxItemBlocks = *runMaxItemBlock
	c.DatetimeStartPos = *runDatetimeStartPos
	c.DatetimeLayout = *runDatetimeLayout
	c.ScoreStyle = *runScoreStyle
	c.ScoreNSize = *runScoreNSize
	c.MinGapToRecord = *runGap
	c.NTopRecordsCount = *runTopNRecordsToShow
	c.ModeblockPerFile = *runModeblockPerFile
	c.NRareTerms = *runNRareTerms

	err := analyzer.Run(c)
	if err != nil {
		log.Printf("%+v", err)
	}

	return nil
}

func topN() error {
	topNFlag.Parse(os.Args[2:])
	if *topNRootDir == "" {
		return fmt.Errorf("rootDir is must")
	}
	if !analyzer.PathExist(*topNRootDir) {
		return fmt.Errorf("%s does not exist", *topNRootDir)
	}
	var startEpoch, endEpoch int64
	var err error
	if *topNStartDate != "" {
		startEpoch, err = analyzer.DateStringToEpoch(*topNStartDate)
		if err != nil {
			return err
		}
	}
	if *topNEndDate != "" {
		endEpoch, err = analyzer.DateStringToEpoch(*topNEndDate)
		if err != nil {
			return err
		}
	}

	err = analyzer.PrintTopN(*topNRootDir,
		*topNRecordsToShow,
		*topNFilterRe, *topNXFilterRe,
		startEpoch, endEpoch,
		*topNMinScore, *topNMaxScore, *topNRareTerms)
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

	err := analyzer.RarStats(*stsRootDir, *stsRecordsToShow)
	return err
}

func report() error {
	reportFlag.Parse(os.Args[2:])
	analyzer.IsDebug = *reportDebug
	return analyzer.Report(*reportConfig, *reportRecentNdays)
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
	case "run":
		err = run()
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
