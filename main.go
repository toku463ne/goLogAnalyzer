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
	iniFlag  = flag.NewFlagSet("clean", flag.ExitOnError)
	frqFlag  = flag.NewFlagSet("frq", flag.ExitOnError)
	rarFlag  = flag.NewFlagSet("rar", flag.ExitOnError)
	usageTxt = `Usage:
loganal rar [-f LOGPATH] [-d DATADIR] [-g GAPVALUE] [-v] [-s SEARCH_KEYS] [-x EXCLUDE_KEYS]
  Starts log analyzation.
	-f LOGPATH: 
		Path of the logfile (can use regex)  
	-v verbose  
	-r RARITY_RATE:
		Top RARITY_RATE log records will be showed.
		Default is 0.0001 (1 rare record out of 10000 records will be showed) 
	-g GAPVALUE: 
		Gap from average score. Default is 0.8
		0 is the average. 
		1 is 1 deviation width from the average. 
		The score is calculated as below and indicates how rare the log record is.
		term score: log10((count of all terms)/(count of the term)) + 1
		log record score: average of term scores in the log record
		* Count is calculated at the point the log record appeared. 
	-d DATADIR: 
		Directory to save the analyzation data.
		This data will be also used in the next time execution
		Only onmemory if not specified.
	-s SEARCH_KEYS: 
		key word to search (can use regex)
	-x EXCLUDE_KEYS: 
		key word to exclude (can use regex)
	

loganal clean -d DATADIR
  Cleans up the analyzation data in previous analysis

loganal stats -d DATADIR
  Shows the statistics from saved data

loganal frq -f LOGPATH [-m MIN_SUPPORT] [-s SEARCH_KEYS] [-x EXCLUDE_KEYS]
  Shows Closed Frequent Itemset order by the support
	-f LOGPATH: Path of the logfile
	-m MIN_SUPPORT: minimum support of closed frequent item sets
	-s SEARCH_KEYS: key word to search (can use regex)
	-x EXCLUDE_KEYS: key word to exclude (can use regex)
  `
)

func clean() error {
	rootDir := iniFlag.String("d", "", "data directory")
	debug := iniFlag.Bool("v", false, "verbose")

	iniFlag.Parse(os.Args[2:])
	err := analyzer.CleanupDb(*rootDir, *debug)
	return err
}

func stats() error {
	rootDir := iniFlag.String("d", "", "data directory")

	iniFlag.Parse(os.Args[2:])
	err := analyzer.Stats(*rootDir)
	return err
}

func rar(verbose bool) error {
	rootDir := rarFlag.String("d", "", "Directory to save the analyzation data")
	pathRegex := rarFlag.String("f", "", "Log file(regex) to analyze")
	filterRe := rarFlag.String("s", "", "key word to search")
	xFilterRe := rarFlag.String("x", "", "key word to exclude")
	gap := rarFlag.Float64("g", 0.0, "Gap rate from average")
	rarityRate := rarFlag.Float64("r", 0.0, "Top RARITY_RATE log records will be showed.")
	debug := rarFlag.Bool("v", false, "show debug logs")
	forceSaveDb := rarFlag.Bool("save", false, "Update the data without asking")
	linesInBlock := rarFlag.Int("linesInBlock", -1, "lines in block")
	maxBlock := rarFlag.Int("maxBlock", -1, "max blocks")
	maxLines := rarFlag.Int("n", 0, "max lines to process")

	rarFlag.Parse(os.Args[2:])

	forceSaveDb1 := *forceSaveDb
	saveDb := false
	if forceSaveDb1 == false {
		if *rootDir != "" {
			if analyzer.PathExist(*rootDir) {
				fmt.Printf("Update data on %s? (y/n) (default 'no')", *rootDir)
				stdin := bufio.NewScanner(os.Stdin)
				stdin.Scan()
				k := stdin.Text()
				if strings.ToLower(k) != "y" && strings.ToLower(k) != "yes" {
					fmt.Printf("input='%s' will not update %s", k, *rootDir)
					saveDb = false
				} else {
					fmt.Printf("input='%s' will update %s", k, *rootDir)
					saveDb = true
				}
			} else {
				saveDb = true
			}
		}
	} else {
		if *rootDir != "" {
			saveDb = true
		}
	}
	rarityCountRate := 1 - *rarityRate

	if err := analyzer.Rar(*pathRegex, *rootDir,
		*filterRe, *xFilterRe,
		*gap, rarityCountRate,
		*maxLines, *linesInBlock, *maxBlock,
		*debug, verbose, saveDb); err != nil {
		return err
	}

	return nil
}

func frq() error {
	path := frqFlag.String("f", "", "Log file to analyze")
	filterRe := frqFlag.String("s", "", "search key")
	xFilterRe := frqFlag.String("x", "", "exclude search key")
	minSupport := frqFlag.Int("m", 0, "min support")
	debug := frqFlag.Bool("v", false, "debug logs")

	frqFlag.Parse(os.Args[2:])

	if err := analyzer.Frq(*path, *minSupport, *filterRe, *xFilterRe, *debug); err != nil {
		return err
	}
	return nil
}

func main() {
	flag.Usage = func() {
		fmt.Printf("%s\n", usageTxt)
	}

	if len(os.Args) < 2 {
		flag.Usage()
		return
	}
	var err error
	err = nil
	opt := os.Args[1]
	switch opt {
	case "clean":
		err = clean()
	case "rar":
		err = rar(false)
	case "stats":
		err = stats()
	case "test":
		err = rar(true)
	case "frq":
		err = frq()
	default:
		flag.Usage()
	}
	if err != nil {
		fmt.Printf("%+v\n", err)
	}

}
