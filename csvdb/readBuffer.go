package csvdb

import (
	"fmt"

	"github.com/pkg/errors"
)

func newReadBuffer(path string, bufferSize int) *readBuff {
	b := new(readBuff)
	b.size = bufferSize
	b.path = path
	b.init()
	return b
}

func (b *readBuff) init() {
	b.rows = make([][]string, b.size)
	b.pos = -1
	b.readPos = 0
}

func (b *readBuff) append(row []string) error {
	b.pos++
	if b.pos >= b.size {
		return errors.WithStack(errors.New(fmt.Sprintf("read buffer over flow: max size=%d", b.size)))
	}
	b.rows[b.pos] = row
	b.values = row
	return nil
}

func (b *readBuff) load() error {
	reader, err := newCsvReader(b.path, 0)
	if err != nil {
		return err
	}
	defer reader.close()
	for reader.next() {
		if err := b.append(reader.values); err != nil {
			return err
		}
	}
	return nil
}

func (b *readBuff) initReadPos() {
	b.readPos = -1
}

func (b *readBuff) next() bool {
	b.readPos++
	if b.readPos > b.pos {
		return false
	}
	b.values = b.rows[b.readPos]
	return true
}
