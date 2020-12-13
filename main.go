package main

import (
	"errors"
	"flag"
	"fmt"
	"os"

	"./analyzer"
)

var (
	iniFlag  = flag.NewFlagSet("init", flag.ContinueOnError)
	runFlag  = flag.NewFlagSet("run", flag.ContinueOnError)
	frqFlag  = flag.NewFlagSet("frq", flag.ContinueOnError)
	testFlag = flag.NewFlagSet("test", flag.ContinueOnError)
	usageTxt = `Usage:
loganal run [-c CONFIGFILE] [-v]
loganal run [-f LOGPATH] [-v]
  Starts log analyzation.
  Cannot use -c and -f together.
    -c CONFIGFILE: The path of config file in ini format. 
    -v verbose
    -f LOGPATH: Path of the logfile.
loganal clanup [-c CONFIGFILE]
  Cleans up the analyzation data in previous analysis
	-c CONFIGFILE: The path of config file in ini format. 
loganal frq -f LOGPATH -m MIN_SUPPORT [-s SEARCH_KEYS] [-x EXCLUDE_KEYS]
  `
)

func destroy() error {
	iniFile := iniFlag.String("c", "", "Ini file")
	iniFlag.Parse(os.Args[2:])
	err := analyzer.Destroy(*iniFile)
	return err
}

func run() error {
	iniFile := runFlag.String("c", "", "Ini file")
	pathRegex := runFlag.String("f", "", "Log file to analyze")
	debug := runFlag.Bool("d", false, "debug")
	runFlag.Parse(os.Args[2:])

	if *iniFile != "" && *pathRegex != "" {
		return errors.New("Cannot use -c and -f together")
	}

	if *iniFile != "" {
		fmt.Printf("run conf=%v verbose=%t\n", *iniFile, *debug)
	}
	err := analyzer.Run(*iniFile, *debug, *pathRegex)
	return err
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
	case "cleanup":
		err = destroy()
	case "run":
		err = run()
	case "frq":
		err = frq()
	default:
		flag.Usage()
	}
	if err != nil {
		fmt.Printf("%+v", err)
	}

}
