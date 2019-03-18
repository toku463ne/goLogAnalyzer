package analyzer

import (
	"errors"
	"fmt"
)

// UsageDCI ... Usage of DCI operation
func UsageDCI() string {
	return fmt.Sprintf(`dci:
	  ./0 dci filename [col=number] [sup=number] [search=regex]
	    col: column number of end of timestamp in the logfile
		sup: minimul support
		search: Search keywork. You can use regex.
`)
}

// UsageYu ... Usage of Yu operation
func UsageYu() string {
	return fmt.Sprintf(`yu:
	  yu filename op=stat: Get statistics of yu
	    ./0 yu filename op=stat [col="column number of end of timestamp in the logfile"]
	  yu filename op=list: List yu scores by time
	    ./0 yu filename op=list [col=column number] [sumlen=length of timestamp considered as one timestamp unit]
`)
}

// RunDCI ... Run DCI algorithm
func RunDCI(args map[string]string) error {
	workdir := args["workdir"]
	file, ok := args["file"]
	if ok == false {
		return errors.New(`"file" is must`)
	}
	col, err := argParseANum(args, "col")
	if err != nil {
		return err
	}

	minSup, err := argParseANum(args, "sup")
	if err != nil {
		return err
	}
	regStr, ok := args["search"]
	if ok == false {
		regStr = ""
	}
	excludeRegStr, ok := args["exclude"]
	if ok == false {
		excludeRegStr = ""
	}
	a, err := newFileAnalyzer(file, col, regStr, excludeRegStr)

	ldci := newLargeDCIClosed(minSup, &a.trans, &a.items, true)
	if err != nil {
		return err
	}
	err = ldci.run()
	if err != nil {
		return err
	}
	err = ldci.outLargeDCIClosed(fmt.Sprintf("%s/closedsets.txt", workdir),
		a.rowNum, regStr, a.trans.mask)
	if err != nil {
		return err
	}

	err = a.outTrans(fmt.Sprintf("%s/transactions.txt", workdir))
	if err != nil {
		return err
	}

	return nil
}

// RunYu ... Run and get Yu score
func RunYu(args map[string]string) error {
	op := args["op"]
	workdir := args["workdir"]
	file, ok := args["file"]
	if ok == false {
		return errors.New(`"file" is must`)
	}
	col, err := argParseANum(args, "col")
	if err != nil {
		return err
	}
	regStr, ok := args["search"]
	if ok == false {
		regStr = ""
	}
	excludeRegStr, ok := args["exclude"]
	if ok == false {
		excludeRegStr = ""
	}

	a, err := newFileAnalyzer(file, col, regStr, excludeRegStr)
	if err != nil {
		return err
	}

	ltrans := a.trans.tranList.len()
	for pos := 0; pos < ltrans; pos += maxBitMatrixXLen {
		xLen := 0
		if maxBitMatrixXLen > ltrans-pos {
			xLen = ltrans - pos
		} else {
			xLen = maxBitMatrixXLen
		}
		matrix, _, _ := tranPart2BitMatrix(&a.trans, &a.items,
			pos, xLen)

		idf := newIdfScores(matrix)

		switch op {
		case "list":
			sumlen, err := argParseANum(args, "sumlen")
			if err != nil {
				return err
			}
			err = idf.outYuScoreByTime(fmt.Sprintf("%s/yuscore.csv", workdir), &a.trans, sumlen)
			if err != nil {
				return err
			}
		case "stat":
			idf.yuStatistics(&a.trans)
		}
	}
	return nil
}
