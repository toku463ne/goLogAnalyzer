package analyzer

import (
	"bufio"
	"compress/gzip"
	"io"
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
	e        error
}

func newReader(filename string) (*reader, error) {
	var fd *os.File
	var err error
	if filename == "" {
		fd = os.Stdin
	} else {
		fd, err = os.OpenFile(filename, os.O_RDONLY, 0644)
		if err != nil {
			return nil, errors.WithStack(err)
		}
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
		lr.zr = zr
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
	} else {
		err := lr.scanner.Err()
		if err == nil {
			lr.e = io.EOF
		} else if err != nil {
			lr.e = err
		}
	}
	return ok
}

func (lr *reader) err() error {
	//return lr.scanner.Err()
	return lr.e
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

func (lr *reader) isOpen() bool {
	if lr.mode == cRModeGZip {
		if lr.zr == nil {
			return false
		}
	}
	if lr.fd == nil {
		return false
	}
	return true
}
