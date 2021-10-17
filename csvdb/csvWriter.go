package csvdb

import (
	"compress/gzip"
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
)

func newCsvWriter(path, writeMode string) (*CsvWriter, error) {
	ext := filepath.Ext(path)
	var fw *os.File
	var zw *gzip.Writer
	var writer *csv.Writer
	mode := ""

	flags := 0
	switch writeMode {
	case CWriteModeWrite:
		flags = os.O_WRONLY | os.O_CREATE | os.O_TRUNC
	default:
		flags = os.O_WRONLY | os.O_CREATE | os.O_APPEND
	}

	fw, err := os.OpenFile(path, flags, 0644)
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

	c := new(CsvWriter)
	c.path = path
	c.writer = writer
	c.fw = fw
	c.zw = zw
	c.mode = mode

	return c, nil
}

func (c *CsvWriter) write(record []string) error {
	err := c.writer.Write(record)
	if err != nil {
		return extError(err, fmt.Sprintf("record=[%s]", strings.Join(record, ",")))
	}
	return nil
}

func (c *CsvWriter) flush() {
	c.writer.Flush()
}

func (c *CsvWriter) close() {
	if c.zw != nil {
		c.zw.Close()
	}

	if c.fw != nil {
		c.fw.Close()
	}
}
