package csvdb

import (
	"compress/gzip"
	"encoding/csv"
	"io"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
)

func newCsvReader(filename string) (*CsvReader, error) {
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

	c := new(CsvReader)
	c.fr = fr
	c.zr = zr
	c.reader = r
	c.filename = filename
	c.mode = mode
	return c, nil
}

func (c *CsvReader) next() bool {
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

func (c *CsvReader) close() {
	if c.zr != nil {
		c.zr.Close()
		c.fr = nil
	}
	if c.fr != nil {
		c.fr.Close()
		c.fr = nil
	}
}
