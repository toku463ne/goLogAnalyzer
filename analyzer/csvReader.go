package analyzer

import (
	"compress/gzip"
	"encoding/csv"
	"io"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
)

type csvReader struct {
	fr       *os.File
	zr       *gzip.Reader
	reader   *csv.Reader
	values   []string
	err      error
	filename string
	mode     string
}

func newCsvReader(filename string) (*csvReader, error) {
	ext := filepath.Ext(filename)
	var fr *os.File
	var zr *gzip.Reader
	var r *csv.Reader
	var err error
	mode := ""

	fr, err = os.Open(filename)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	if ext == ".gz" || ext == ".gzip" {
		zr, err = gzip.NewReader(fr)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		r = csv.NewReader(zr)
		mode = cRModeGZip
	} else {
		r = csv.NewReader(fr)
		mode = cRModePlain
	}

	c := new(csvReader)
	c.fr = fr
	c.zr = zr
	c.reader = r
	c.filename = filename
	c.mode = mode
	return c, nil
}

func (c *csvReader) next() bool {
	values, err := c.reader.Read()
	c.err = err
	if err == io.EOF {
		return false
	}
	if err != nil {
		return false
	}
	c.values = values
	c.err = nil
	return true
}

func (c *csvReader) close() {
	if c.zr != nil {
		c.zr.Close()
	}
	if c.fr != nil {
		c.fr.Close()
	}
}
