package analyzer

import (
	"bufio"
	"fmt"
	"os"

	"github.com/pkg/errors"
)

// FileAnalyzer ... File analyzer
type FileAnalyzer struct {
	filepath        string
	timeStampEndCol int
	matrix          *bitMatrix
	items           items
	trans           trans
	rowNum          int
}

// NewFileAnalyzer ... new analyzer object
func newFileAnalyzer(filepath string,
	timeStampEndCol int, regStr string) (*FileAnalyzer, error) {
	a := new(FileAnalyzer)
	a.filepath = filepath
	a.timeStampEndCol = timeStampEndCol
	a.items = *newItems()
	a.trans = *newTrans()
	rowNum, err := a.tokenizeFile(regStr)
	a.rowNum = rowNum
	if err != nil {
		return nil, err
	}
	return a, nil
}

func (a *FileAnalyzer) loadMatrix() {
	a.matrix = tran2BitMatrix(&a.trans, &a.items)
}

func (a *FileAnalyzer) tokenizeFile(regStr string) (int, error) {
	file, err := os.Open(a.filepath)
	defer file.Close()

	if err != nil {
		return -1, errors.Wrapf(err, fmt.Sprintf("file open error: %s", a.filepath))
	}

	reader := bufio.NewReader(file)
	var line string
	eof := false
	i := 0
	for {
		line, err = reader.ReadString('\n')
		if err != nil {
			eof = true
		}

		tokenizeLine(line, a.timeStampEndCol, &a.trans, &a.items, regStr, i)
		if eof {
			break
		}
		i++
	}
	if bolShowProgress {
		fmt.Print("\n")
	}

	return i, nil
}

func (a *FileAnalyzer) getTokens() [][]int {
	return a.trans.getSlice()
}

func (a *FileAnalyzer) getTimeStamps() []string {
	return a.trans.tranTimeStamps.getSlice()
}

func (a *FileAnalyzer) outTrans(filepath string) error {
	ou, err := os.Create(filepath)
	if err != nil {
		return err
	}
	w := bufio.NewWriter(ou)
	defer ou.Close()
	//ts := a.trans.tranTimeStamps.getSlice()
	for i := 0; i <= a.trans.maxTranID; i++ {
		line := fmt.Sprintf("%d %v\n", i, a.trans.get(i))
		if _, err := fmt.Fprint(w, line); err != nil {
			return err
		}
	}
	w.Flush()
	return nil
}
