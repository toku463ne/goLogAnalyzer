package analyzer

import (
	"encoding/csv"
	"io"
	"os"

	"github.com/pkg/errors"
)

type csvReader struct {
	fr       *os.File
	reader   *csv.Reader
	values   []string
	err      error
	filename string
	name     string
}

func newCsvReader(filename string) (*csvReader, error) {
	fr, err := os.Open(filename)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	r := csv.NewReader(fr)
	c := new(csvReader)
	c.fr = fr
	c.reader = r
	c.filename = filename
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
	if c.fr != nil {
		c.fr.Close()
	}
}
