package sstable

import (
	"encoding/binary"
	"io"
	"os"

	"github.com/emrecanterzi/wisp/internal/types"
)

type entry struct {
	dataFile  string
	indexFile string
	index     []IndexEntry
	loaded    bool
}

func (e *entry) loadIndex() error {
	file, err := os.Open(e.indexFile)
	if err != nil {
		return err
	}
	defer file.Close()

	keyLenBuf := make([]byte, 4)
	var keyBuf []byte
	offsetBuf := make([]byte, 8)

	for {
		_, err := io.ReadFull(file, keyLenBuf)
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		keyBuf = make([]byte, binary.BigEndian.Uint32(keyLenBuf))

		_, err = io.ReadFull(file, keyBuf)
		if err != nil {
			return err
		}

		_, err = io.ReadFull(file, offsetBuf)
		if err != nil {
			return err
		}

		e.index = append(e.index, IndexEntry{
			Key:    string(keyBuf),
			Offset: int64(binary.BigEndian.Uint64(offsetBuf)),
		})
	}

	e.loaded = true
	return nil
}

func (e *entry) scan(startOffset, endOffset int64, key string) (*types.Record, error) {
	offset := startOffset
	file, err := os.Open(e.dataFile)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	_, err = file.Seek(startOffset, io.SeekStart)
	if err != nil {
		return nil, err
	}

	recordLenBuf := make([]byte, 4)
	var recordBuf []byte

	for {
		_, err = io.ReadFull(file, recordLenBuf)
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}

		recordBuf = make([]byte, binary.BigEndian.Uint32(recordLenBuf))
		_, err = io.ReadFull(file, recordBuf)
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}

		rec, err := Decode(recordBuf)
		if err != nil {
			return nil, err
		}

		if key == rec.Key {
			return &rec, nil
		} else if rec.Key > key {
			break
		}

		offset += 4 + int64(binary.BigEndian.Uint32(recordLenBuf))

		if offset >= endOffset && endOffset != -1 {
			break
		}

	}

	return nil, nil
}
