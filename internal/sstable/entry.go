package sstable

import (
	"encoding/binary"
	"io"
	"os"
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

func (e *entry) scan(startOffset, endOffset int64, key string) (string, bool, error) {
	offset := startOffset
	file, err := os.Open(e.dataFile)
	if err != nil {
		return "", false, err
	}
	defer file.Close()

	_, err = file.Seek(startOffset, io.SeekStart)
	if err != nil {
		return "", false, err
	}

	keyLenBuf := make([]byte, 4)
	var keyBuf []byte
	valueLenBuf := make([]byte, 4)
	var valBuf []byte

	for {
		_, err = io.ReadFull(file, keyLenBuf)
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", false, err
		}

		keyBuf = make([]byte, binary.BigEndian.Uint32(keyLenBuf))
		_, err = io.ReadFull(file, keyBuf)
		if err != nil {
			return "", false, err
		}

		_, err = io.ReadFull(file, valueLenBuf)
		if err != nil {
			return "", false, err
		}

		valBuf = make([]byte, binary.BigEndian.Uint32(valueLenBuf))
		_, err = io.ReadFull(file, valBuf)
		if err != nil {
			return "", false, err
		}

		if key == string(keyBuf) {
			return string(valBuf), true, nil
		} else if string(keyBuf) > key {
			break
		}

		offset += int64(4 + len(keyBuf) + 4 + len(valBuf))

		if offset >= endOffset && endOffset != -1 {
			break
		}

	}

	return "", false, nil
}
