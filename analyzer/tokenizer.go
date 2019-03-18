package analyzer

import (
	"bytes"
	"fmt"
	"regexp"

	"golang.org/x/text/unicode/norm"
)

func getEnItems(content []byte, items1 *items, regStr string) []int {
	content = bytes.ToLower(content)
	content = remTags.ReplaceAll(content, []byte(" "))
	content = norm.NFC.Bytes(content)
	words := regexp.MustCompile(fmt.Sprintf("%s|%s", wordReStr, regStr)).FindAll(content, -1)
	result := make([]int, 0)
	for _, w := range words {
		word := string(w)
		if len(word) <= 2 {
			continue
		}
		if _, ok := english[word]; ok == false {
			itemID := items1.regist(word)
			result = append(result, itemID)
		}
	}
	return result
}

func tokenizeLine(line string, timeStampEndCol int,
	trans1 *trans, items1 *items, regStr string, excludeRegStr string, rowNum int) {
	var timeStamp string
	if len(line) > timeStampEndCol {
		timeStamp = line[:timeStampEndCol]
		line = line[timeStampEndCol:]
	} else {
		timeStamp = ""
	}

	if searchReg(line, regStr) == false {
		return
	}
	if searchReg(line, excludeRegStr) {
		return
	}

	trans1.mask.append(rowNum)
	bline := []byte(line)

	tran := getEnItems(bline, items1, regStr)

	//if bolShowProgress {
	//	fmt.Printf("\rOn %d maxid=%d", maxTranID, maxItemID)
	//}
	if len(tran) > 0 {
		trans1.add(timeStamp, tran, line)
	}
}
