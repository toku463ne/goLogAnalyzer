package analyzer

import (
	"encoding/csv"
	"fmt"
	"os"
	"strings"

	"github.com/pkg/errors"
)

type csvWriter struct {
	fw     *os.File
	writer *csv.Writer
	path   string
}

func newCsvWriter(path string) (*csvWriter, error) {
	fw, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	writer := csv.NewWriter(fw)
	c := new(csvWriter)
	c.path = path
	c.writer = writer
	c.fw = fw
	return c, nil
}

func (c *csvWriter) write(record []string) error {
	err := c.writer.Write(record)
	if err != nil {
		return extError(err, fmt.Sprintf("record=[%s]", strings.Join(record, ",")))
	}
	return nil
}

func (c *csvWriter) flush() {
	c.writer.Flush()
}

func (c *csvWriter) close() {
	if c.fw != nil {
		c.fw.Close()
	}
}
