package main

import (
	"flag"
	"fmt"
	"os"

	"./analyzer"
)

var (
	iniFlag  = flag.NewFlagSet("init", flag.ContinueOnError)
	runFlag  = flag.NewFlagSet("run", flag.ContinueOnError)
	testFlag = flag.NewFlagSet("test", flag.ContinueOnError)
)

func help() {

}

func destroy() error {
	iniFile := iniFlag.String("c", "", "Ini file")
	iniFlag.Parse(os.Args[2:])
	err := analyzer.Destroy(*iniFile)
	return err
}

/*
func initData() error {
	iniFile := iniFlag.String("c", "", "Ini file")
	iniFlag.Parse(os.Args[2:])
	err := analyzer.Init(*iniFile)
	return err
}
*/

func run() error {
	iniFile := runFlag.String("c", "", "Ini file")
	verbose := runFlag.Bool("v", false, "verbose")
	runFlag.Parse(os.Args[2:])
	fmt.Printf("run = %v %t\n", *iniFile, *verbose)
	err := analyzer.Run(*iniFile, *verbose)
	return err
}

func test() error {
	testName := runFlag.String("n", "", "Name test")
	testFlag.Parse(os.Args[2:])
	switch *testName {

	}
	return nil
}

func main() {
	if len(os.Args) < 2 {
		return
	}
	var err error
	opt := os.Args[1]
	switch opt {
	case "destroy":
		err = destroy()
	//case "init":
	//	err = initData()
	case "run":
		err = run()
	case "test":
		err = test()
	default:
		flag.Parse()
	}
	if err != nil {
		fmt.Printf("%+v", err)
	}

}
