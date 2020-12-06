package analyzer

import (
	"fmt"
	"io"

	"github.com/pkg/errors"
)

type filePointer struct {
	files   []string
	epochs  []int64
	r       *reader
	lastRow int
	pos     int
	e       error
}

func newFilePointer(pathRegex string,
	lastEpoch int64, lastRow int) *filePointer {
	fp := new(filePointer)
	var targetFiles []string
	var targetEpochs []int64
	epochs, files := getSortedGlob(pathRegex)
	for i, f := range files {
		epoch := epochs[i]
		if (epoch == lastEpoch && lastRow != cEOF) || epoch > lastEpoch {
			targetFiles = append(targetFiles, f)
			targetEpochs = append(targetEpochs, epoch)
		}
	}

	fp.files = targetFiles
	fp.epochs = targetEpochs
	fp.lastRow = lastRow
	fp.pos = 0
	return fp
}

func (fp *filePointer) currFile() string {
	return fp.files[fp.pos]
}
func (fp *filePointer) currFileEpoch() int64 {
	return fp.epochs[fp.pos]
}

func (fp *filePointer) open() error {
	if fp.r != nil {
		fp.close()
	}
	if fp.files == nil || len(fp.files) == 0 {
		return errors.New("no files to open")
	}
	fp.pos = 0
	currRow := fp.lastRow
	r, err := newReader(fp.files[0])
	if err != nil {
		return errors.WithStack(err)
	}
	logInfo(fmt.Sprintf("Opened %s", fp.files[0]))
	row := 0
	if currRow > 0 {
		for r.next() {
			row++
			if row >= currRow {
				break
			}
		}
		if err := r.err(); err != nil {
			return err
		}
	}
	fp.r = r
	return nil
}

func (fp *filePointer) next() bool {
	if fp.r == nil {
		fp.e = errors.New("open() first")
		return false
	}

	ok := fp.r.next()
	err := fp.r.err()
	fp.e = err
	if !ok {
		if err != nil && err != io.EOF {
			return false
		}
	}
	if ok {
		fp.e = nil
		return true
	}

	if fp.pos+1 >= len(fp.files) {
		fp.e = io.EOF
		return false
	}
	if fp.r != nil {
		fp.r.close()
		fp.r = nil
	}

	fp.pos++
	r, err := newReader(fp.files[fp.pos])
	if err != nil {
		fp.e = errors.WithStack(err)
		return false
	}
	fp.r = r

	return fp.r.next()
}

func (fp *filePointer) err() error {
	return fp.e
}

func (fp *filePointer) text() string {
	return fp.r.text()
}

func (fp *filePointer) row() int {
	return fp.r.row()
}

func (fp *filePointer) close() {
	if fp.r != nil {
		fp.r.close()
		fp.r = nil
	}
	fp.pos = 0
}

func (fp *filePointer) isOpen() bool {
	if fp.r == nil || fp.r.isOpen() == false {
		return false
	}
	return true
}
