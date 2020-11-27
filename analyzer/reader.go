package analyzer

import (
	"bufio"
	"compress/gzip"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
)

type reader struct {
	fd       *os.File
	zr       *gzip.Reader
	scanner  *bufio.Scanner
	rowNum   int
	mode     string
	filename string
}

func newReader(filename string) (*reader, error) {
	fd, err := os.Open(filename)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	lr := new(reader)
	lr.fd = fd

	ext := filepath.Ext(filename)
	if ext == ".gz" || ext == ".gzip" {
		lr.mode = cRModeGZip
	} else {
		lr.mode = cRModePlain
	}

	if lr.mode == cRModeGZip {
		zr, err := gzip.NewReader(fd)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		lr.scanner = bufio.NewScanner(zr)
	} else {
		lr.scanner = bufio.NewScanner(fd)
	}
	//lr.scanner.Split(bufio.ScanBytes)
	lr.filename = filename
	return lr, nil
}

func (lr *reader) next() bool {
	ok := lr.scanner.Scan()
	if ok {
		lr.rowNum++
	}
	return ok
}

func (lr *reader) err() error {
	return lr.scanner.Err()
}

func (lr *reader) text() string {
	return lr.scanner.Text()
}

func (lr *reader) row() int {
	return lr.rowNum
}

func (lr *reader) close() {
	if lr.mode == cRModeGZip {
		if lr.zr != nil {
			lr.zr.Close()
		}
	}
	if lr.fd != nil {
		lr.fd.Close()
	}
}
