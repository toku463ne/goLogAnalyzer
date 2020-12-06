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
	testFlag = flag.NewFlagSet("test", flag.ContinueOnError)
	usageTxt = `Usage:
loganal run [-c CONFIGFILE] [-v]
loganal run [-p LOGPATH] [-v]
  Starts log analyzation.
  Cannot use -c and -p together.
    -c CONFIGFILE: The path of config file in ini format. 
    -v verbose
    -p LOGPATH: Path of the logfile.
loganal clanup [-c CONFIGFILE]
  Cleans up the analyzation data in previous analysis
    -c CONFIGFILE: The path of config file in ini format. 
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
	pathRegex := runFlag.String("p", "", "Log file to analyze")
	verbose := runFlag.Bool("v", false, "verbose")
	runFlag.Parse(os.Args[2:])

	if *iniFile != "" && *pathRegex != "" {
		return errors.New("Cannot use -c and -p together")
	}

	if *iniFile != "" {
		fmt.Printf("run conf=%v verbose=%t\n", *iniFile, *verbose)
	}
	err := analyzer.Run(*iniFile, *verbose, *pathRegex)
	return err
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
	default:
		flag.Usage()
	}
	if err != nil {
		fmt.Printf("%+v", err)
	}

}
