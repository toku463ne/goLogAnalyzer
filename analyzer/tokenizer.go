package analyzer

import (
	"bytes"
	"fmt"
	"regexp"

	"golang.org/x/text/unicode/norm"
)

func getEnItems(content []byte, items1 *items, filterRe string) []int {
	content = bytes.ToLower(content)
	content = remTags.ReplaceAll(content, []byte(" "))
	content = norm.NFC.Bytes(content)
	words := regexp.MustCompile(fmt.Sprintf("%s|%s", cWordReStr, filterRe)).FindAll(content, -1)
	result := make([]int, 0)
	for _, w := range words {
		word := string(w)
		if len(word) <= 2 {
			continue
		}
		if _, ok := english[word]; ok == false {
			itemID := items1.regist(word, 1, true)
			result = append(result, itemID)
		}
	}
	return result
}

func tokenizeLine(line string,
	trans1 *trans, items1 *items, filterRe string,
	xFilterRe string,
	rowNum int) bool {
	isAdded := false

	if filterRe != "" && searchReg(line, filterRe) == false {
		return isAdded
	}
	if xFilterRe != "" && searchReg(line, xFilterRe) {
		return isAdded
	}

	trans1.mask.append(rowNum)
	bline := []byte(line)

	tran := getEnItems(bline, items1, filterRe)

	if len(tran) > 0 {
		trans1.add(tran, line, items1)
		if verbose {
			tranID := trans1.maxTranID
			trans1.lastMsg = fmt.Sprintf("%s",
				trans1.getSentenceAt(tranID, items1))
		}
		isAdded = true
	}
	return isAdded
}
