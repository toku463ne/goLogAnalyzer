package main

import (
	"fmt"
	"os"

	"./analyzer"
)

func usage() {
	fmt.Printf(`usage: ./0 op_name args
Parse log file and show the result.
op_name: 
	%s
	%s`,
		analyzer.UsageDCI(), analyzer.UsageYu())
	fmt.Print("\n")

}

func main() {
	var err error
	if len(os.Args) < 2 {
		usage()
		return
	}
	logmsg("started")
	operation := os.Args[1]
	args := argParse(os.Args[2:])
	usagestr := ""
	os.Mkdir("report", 0755)
	args["workdir"] = "report"
	switch operation {
	case "dci":
		usagestr = analyzer.UsageDCI()
		err = analyzer.RunDCI(args)
	case "yu":
		usagestr = analyzer.UsageYu()
		err = analyzer.RunYu(args)
	default:
		usage()
		return
	}
	if err != nil {
		fmt.Printf("%+v\n", err)
		fmt.Printf("%s", usagestr)
		return
	}
	logmsg("finished")
}
