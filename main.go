package main

import (
	"flag"
	"fmt"
	"os"

	"./analyzer"
)

var (
	iniFlag  = flag.NewFlagSet("clean", flag.ExitOnError)
	rarFlag  = flag.NewFlagSet("rar", flag.ExitOnError)
	frqFlag  = flag.NewFlagSet("frq", flag.ExitOnError)
	testFlag = flag.NewFlagSet("test", flag.ExitOnError)
	usageTxt = `Usage:
loganal rar [-f LOGPATH] [-d DATADIR] [-g GAPVALUE] [-v] [-s SEARCH_KEYS] [-x EXCLUDE_KEYS]
  Starts log analyzation.
	-f LOGPATH: 
		Path of the logfile (can use regex)  
	-v verbose  
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

loganal frq -f LOGPATH [-m MIN_SUPPORT] [-s SEARCH_KEYS] [-x EXCLUDE_KEYS]
  Shows Closed Frequent Itemset order by the support
	-f LOGPATH: Path of the logfile
	-m MIN_SUPPORT: minimum support of closed frequent item sets
	-s SEARCH_KEYS: key word to search (can use regex)
	-x EXCLUDE_KEYS: key word to exclude (can use regex)
  `
)

func clean() error {
	iniFile := iniFlag.String("c", "", "Ini file")
	iniFlag.Parse(os.Args[2:])
	err := analyzer.CleanupDb(*iniFile)
	return err
}

func rar() error {
	rootDir := rarFlag.String("d", "", "Directory to save the analyzation data")
	pathRegex := rarFlag.String("f", "", "Log file(regex) to analyze")
	filterRe := rarFlag.String("s", "", "key word to search")
	xFilterRe := rarFlag.String("x", "", "key word to exclude")
	gap := rarFlag.Float64("g", 0.0, "Gap rate from average")
	debug := rarFlag.Bool("v", false, "verbose")
	rarFlag.Parse(os.Args[2:])

	if err := analyzer.Rar(*pathRegex, *rootDir,
		*filterRe, *xFilterRe,
		*gap,
		-1, -1, *debug); err != nil {
		return err
	}

	return nil
}

func frq() error {
	path := frqFlag.String("f", "", "Log file to analyze")
	filterRe := frqFlag.String("s", "", "search key")
	xFilterRe := frqFlag.String("x", "", "exclude search key")
	minSupport := frqFlag.Int("m", 0, "min support")
	frqFlag.Parse(os.Args[2:])

	if err := analyzer.Frq(*path, *minSupport, *filterRe, *xFilterRe); err != nil {
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
		err = rar()
	case "frq":
		err = frq()
	default:
		flag.Usage()
	}
	if err != nil {
		fmt.Printf("%+v", err)
	}

}
