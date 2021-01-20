package analyzer

import (
	"bufio"
	"fmt"
	"os"

	"github.com/pkg/errors"
)

// NewFileAnalyzer ... new analyzer object
func tokenizeFile(maxRecs int, filepath, filterRe, xfilterRe string) (*trans, error) {
	trans := newTrans(true)

	var err error
	var file *os.File
	if filepath == "" {
		file = os.Stdin
	} else {
		file, err = os.Open(filepath)
	}
	defer file.Close()

	if err != nil {
		return nil, errors.Wrapf(err, fmt.Sprintf("file open error: %s", filepath))
	}
	logDebug(fmt.Sprintf("tokenizing file %s", filepath))
	reader := bufio.NewReader(file)
	var line string
	eof := false
	i := 0
	for {
		line, err = reader.ReadString('\n')
		if err != nil {
			eof = true
		}

		trans.tokenizeLine(line, filterRe, xfilterRe)
		if eof {
			break
		}
		i++
		if i >= maxRecs {
			break
		}
	}

	return trans, nil
}
