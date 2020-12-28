package analyzer

import (
	"bufio"
	"fmt"
	"os"

	"github.com/pkg/errors"
)

// FileAnalyzer ... File analyzer
type FileAnalyzer struct {
	filepath      string
	items         *items
	trans         *trans
	regStr        string
	excludeRegStr string
	rowNum        int
}

// NewFileAnalyzer ... new analyzer object
func newFileAnalyzer(filepath, regStr, excludeRegStr string) *FileAnalyzer {
	a := new(FileAnalyzer)
	a.filepath = filepath
	a.items = newItems()
	a.trans = newTrans()
	a.regStr = regStr
	a.excludeRegStr = excludeRegStr
	return a
}

func (a *FileAnalyzer) tokenizeFile(maxRecs int) error {
	var err error
	var file *os.File
	if a.filepath == "" {
		file = os.Stdin
	} else {
		file, err = os.Open(a.filepath)
	}
	defer file.Close()

	if err != nil {
		return errors.Wrapf(err, fmt.Sprintf("file open error: %s", a.filepath))
	}
	logDebug(fmt.Sprintf("tokenizing file %s", a.filepath))
	reader := bufio.NewReader(file)
	var line string
	eof := false
	i := 0
	for {
		line, err = reader.ReadString('\n')
		if err != nil {
			eof = true
		}

		tokenizeLine(line, a.trans, a.items, a.regStr, a.excludeRegStr, i)
		if eof {
			break
		}
		i++
		if i >= maxRecs {
			break
		}
	}
	a.rowNum = i

	return nil
}

func (a *FileAnalyzer) output(filepath string) error {
	for i := 0; i <= a.trans.maxTranID; i++ {
		fmt.Printf("%d | %v", i, a.trans.get(i))
	}
	return nil
}
