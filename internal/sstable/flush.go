package sstable

import (
	"bytes"
	"encoding/binary"
	"os"
)

type flushState struct {
	file            *os.File
	indexFile       *os.File
	offset          int64
	bytesSinceIndex int
	index           []IndexEntry
}

func (fs *flushState) writeEntry(key, value string) error {
	entryOffset := fs.offset

	var buf bytes.Buffer

	keyLen := int32(len(key))
	valLen := int32(len(value))

	if err := binary.Write(&buf, binary.BigEndian, keyLen); err != nil {
		return err
	}
	if _, err := buf.Write([]byte(key)); err != nil {
		return err
	}
	if err := binary.Write(&buf, binary.BigEndian, valLen); err != nil {
		return err
	}
	if _, err := buf.Write([]byte(value)); err != nil {
		return err
	}

	_, err := fs.file.Write(buf.Bytes())
	if err != nil {
		return err
	}

	entrySize := 4 + len(key) + 4 + len(value)
	fs.offset += int64(entrySize)
	fs.bytesSinceIndex += entrySize

	if fs.bytesSinceIndex >= 4096 {
		fs.index = append(fs.index, IndexEntry{Key: key, Offset: entryOffset})
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
