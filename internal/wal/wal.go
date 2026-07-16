package wal

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
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

func (w *WAL) Replay(fn func(operation uint8, key, value []byte)) error {
	file, err := os.Open("data/wisp.wal")
	if err != nil {
		return err
	}
	defer file.Close()

	opByte := make([]byte, 1)
	keyLenBuf := make([]byte, 4)
	var keyBuf []byte
	valueLenBuf := make([]byte, 4)
	var valueBuf []byte

	for {
		_, err := io.ReadFull(file, opByte)

		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		operation := opByte[0]

		_, err = io.ReadFull(file, keyLenBuf)
		if err != nil {
			return err
		}

		keyBuf = make([]byte, binary.BigEndian.Uint32(keyLenBuf))

		_, err = io.ReadFull(file, keyBuf)
		if err != nil {
			return err
		}

		valueBuf = nil
		if operation == 1 {
			_, err = io.ReadFull(file, valueLenBuf)
			if err != nil {
				return err
			}

			valueBuf = make([]byte, binary.BigEndian.Uint32(valueLenBuf))

			_, err = io.ReadFull(file, valueBuf)
			if err != nil {
				return err
			}
		}

		fn(operation, keyBuf, valueBuf)
	}

	return nil
}
