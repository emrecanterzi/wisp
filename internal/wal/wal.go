package wal

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math"
	"os"
)

type WAL struct {
	file *os.File
}

func NewWAL() (*WAL, error) {
	err := os.MkdirAll("data", 0755)
	if err != nil {
		return nil, err
	}

	file, err := os.OpenFile("data/wisp.wal", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}

	return &WAL{
		file: file,
	}, nil
}

// operation 0 is delete, operation 1 is insert
func (w *WAL) Write(operation uint8, key, value []byte) error {
	var buf bytes.Buffer

	if operation == 0 || operation == 1 {
		err := buf.WriteByte(operation)
		if err != nil {
			return err
		}
	} else {
		return fmt.Errorf("operation not supported")
	}

	if len(key) > math.MaxUint32 || len(value) > math.MaxUint32 {
		return fmt.Errorf("key or value too large")
	}

	if err := binary.Write(&buf, binary.BigEndian, uint32(len(key))); err != nil {
		return err
	}
	if _, err := buf.Write(key); err != nil {
		return err
	}

	if operation == 1 {
		if err := binary.Write(&buf, binary.BigEndian, uint32(len(value))); err != nil {
			return err
		}

		if _, err := buf.Write(value); err != nil {
			return err
		}
	}

	_, err := w.file.Write(buf.Bytes())
	return err
}
