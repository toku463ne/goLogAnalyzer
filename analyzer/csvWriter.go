package analyzer

import (
	"compress/gzip"
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
)

type csvWriter struct {
	fw     *os.File
	zw     *gzip.Writer
	writer *csv.Writer
	path   string
	mode   string
}

func newCsvWriter(path string) (*csvWriter, error) {
	ext := filepath.Ext(path)
	var fw *os.File
	var zw *gzip.Writer
	var writer *csv.Writer
	mode := ""
	fw, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	if ext == ".gz" || ext == ".gzip" {
		zw = gzip.NewWriter(fw)
		writer = csv.NewWriter(zw)
		mode = cRModeGZip
	} else {
		writer = csv.NewWriter(fw)
		mode = cRModePlain
	}

	c := new(csvWriter)
	c.path = path
	c.writer = writer
	c.fw = fw
	c.zw = zw
	c.mode = mode

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
	if c.zw != nil {
		c.zw.Close()
	}

	if c.fw != nil {
		c.fw.Close()
	}
}
