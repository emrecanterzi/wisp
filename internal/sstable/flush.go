package sstable

import (
	"bytes"
	"encoding/binary"
	"os"

	"github.com/emrecanterzi/wisp/internal/types"
)

type flushState struct {
	file            *os.File
	indexFile       *os.File
	offset          int64
	bytesSinceIndex int
	index           []IndexEntry
}

func (fs *flushState) writeEntry(rec types.Record) error {
	entryOffset := fs.offset
	data := Encode(rec)
	entrySize := len(data)

	var buf bytes.Buffer
	binary.Write(&buf, binary.BigEndian, int32(entrySize))
	buf.Write(data)

	_, err := fs.file.Write(buf.Bytes())
	if err != nil {
		return err
	}

	fs.offset += 4 + int64(entrySize)
	fs.bytesSinceIndex += 4 + entrySize

	if fs.bytesSinceIndex >= 4096 {
		fs.index = append(fs.index, IndexEntry{Key: rec.Key, Offset: entryOffset})
		fs.bytesSinceIndex = 0
	}

	return nil
}

func (fs *flushState) persistIndex() error {
	var buf bytes.Buffer

	for _, index := range fs.index {
		keyLen := int32(len(index.Key))

		if err := binary.Write(&buf, binary.BigEndian, keyLen); err != nil {
			return err
		}
		if _, err := buf.Write([]byte(index.Key)); err != nil {
			return err
		}
		if err := binary.Write(&buf, binary.BigEndian, index.Offset); err != nil {
			return err
		}
	}

	_, err := fs.indexFile.Write(buf.Bytes())
	return err
}
