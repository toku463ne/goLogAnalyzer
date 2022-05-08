package analyzer

import (
	"fmt"
	"io"
	"strings"

	"github.com/pkg/errors"
)

func newFilePointer(pathRegex string,
	lastEpoch int64, lastRow int) (*filePointer, error) {
	fp := new(filePointer)
	var targetFiles []string
	var targetEpochs []int64
	if pathRegex == "" {
		targetFiles = []string{""}
		targetEpochs = []int64{0}
	} else {
		epochs, files, err := getSortedGlob(pathRegex)
		if err != nil {
			fp.currErr = err
			return nil, err
		}
		for i, f := range files {
			epoch := epochs[i]
			if (epoch == lastEpoch && lastRow != cEOF) || epoch > lastEpoch {
				targetFiles = append(targetFiles, f)
				targetEpochs = append(targetEpochs, epoch)
			}
		}
	}
	if IsDebug {
		msg := "filePointer.newFilePointer(): "
		msg += fmt.Sprintf("targets=%s lastRow=%d",
			strings.Join(targetFiles, ";"), lastRow)
		ShowDebug(msg)
	}

	fp.files = targetFiles
	fp.epochs = targetEpochs
	fp.lastRow = lastRow
	fp.pos = 0
	fp.isEOF = false
	fp.currPos = 0
	return fp, nil
}

func (fp *filePointer) currFileEpoch() int64 {
	return fp.epochs[fp.pos]
}

func (fp *filePointer) err() error {
	return fp.currErr
}

func (fp *filePointer) open() error {
	if fp.r != nil {
		fp.close()
	}
	if fp.files == nil {
		return errors.New("no files to open")
	}
	fp.pos = 0
	currRow := fp.lastRow
	r, err := newReader(fp.files[0])
	if err != nil {
		return errors.WithStack(err)
	}

	if !r.next() {
		if err := r.err(); err != nil {
			return err
		}
	}

	row := 0
	if currRow > 0 {
		if IsDebug {
			msg := "filePointer.open(): "
			msg += fmt.Sprintf("Moving to row=%d", currRow)
			ShowDebug(msg)
		}
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
	if IsDebug {
		msg := "filePointer.open(): completed"
		ShowDebug(msg)
	}
	return nil
}

// this function is critical for performance
// need to keep as simple as possible
func (fp *filePointer) next() bool {
	// don't consider the case fp.r is nil
	// case it is nil, it means open() has not been done which is considered as a bug

	err := fp.e
	fp.currErr = err
	if err == io.EOF {
		return false
	}
	if err != nil {
		fp.currText = ""
		fp.currRow = -1
		return false
	}
	fp.currText = fp.r.text()
	fp.currRow = fp.r.rowNum
	fp.currPos = fp.pos

	ok := fp.r.next()
	if ok {
		fp.isEOF = false
		return true
	}

	err = fp.r.err()
	if err != nil && err != io.EOF {
		fp.e = err
	}
	fp.isEOF = true

	if fp.pos+1 >= len(fp.files) {
		fp.e = io.EOF
		return true
	}
	if fp.r != nil {
		fp.r.close()
		fp.r = nil
	}

	fp.pos++
	r, err := newReader(fp.files[fp.pos])
	if err != nil {
		fp.e = errors.WithStack(err)
		return true
	}
	fp.r = r
	fp.e = nil

	fp.r.next()
	return true
}

func (fp *filePointer) isLastFile() bool {
	return fp.currPos+1 >= len(fp.files)
}

func (fp *filePointer) text() string {
	//return fp.r.text()
	return fp.currText
}

func (fp *filePointer) row() int {
	//return fp.r.rowNum
	return fp.currRow
}

func (fp *filePointer) close() {
	if fp.r != nil {
		fp.r.close()
		fp.r = nil
	}
	fp.pos = 0
}

func (fp *filePointer) isOpen() bool {
	if fp.r == nil || !fp.r.isOpen() {
		return false
	}
	return true
}

func countNFiles(totalFileCount int, pathRegex string) (int, int, error) {
	fp, err := newFilePointer(pathRegex, 0, 0)
	if err != nil {
		return -1, -1, err
	}
	if fp == nil || len(fp.files) == 0 {
		return 0, 0, nil
	}
	cnt := 0
	fileCnt := 0
	for i, filename := range fp.files {
		if i >= totalFileCount {
			break
		}
		fileCnt++
		r, err := newReader(filename)
		if err != nil {
			return -1, fileCnt, err
		}
		for {
			_, _, err := r.reader.ReadLine()
			if err != nil {
				if err != io.EOF {
					return -1, fileCnt, err
				}
				break
			}
			cnt++
		}
	}
	fp.close()
	return cnt, fileCnt, nil
}
