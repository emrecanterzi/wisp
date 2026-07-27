package wal

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"os"
	"sync"
	"time"
)

type WAL struct {
	dir  string
	file *os.File
	mu   sync.Mutex
}

func NewWAL(dir string) (*WAL, error) {
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		return nil, err
	}

	file, err := os.OpenFile(fmt.Sprintf("%s/%d.wal", dir, time.Now().UnixNano()), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}

	return &WAL{
		dir:  dir,
		file: file,
	}, nil
}

// operation 0 is delete, operation 1 is insert
func (w *WAL) Write(operation uint8, key, value []byte) error {
	w.mu.Lock()
	defer w.mu.Unlock()
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

func (w *WAL) ReplyWals(fn func(operation uint8, key, value []byte)) error {
	files, err := os.ReadDir(w.dir)
	if err != nil {
		return err
	}

	for _, entry := range files {
		if entry.IsDir() {
			continue
		}

		file, err := os.Open(w.dir + "/" + entry.Name())
		if err != nil {
			return err
		}

		if err := w.replay(file, fn); err != nil {
			file.Close()
			return err
		}
		if err := file.Close(); err != nil {
			return err
		}
	}

	return nil
}

func (w *WAL) replay(file *os.File, fn func(operation uint8, key, value []byte)) error {
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

func (w *WAL) Rotate() error {
	w.mu.Lock()
	err := w.file.Close()
	if err != nil {
		w.mu.Unlock()
		return err
	}
	file, err := os.OpenFile(fmt.Sprintf("%s/%d.wal", w.dir, time.Now().UnixNano()), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		w.mu.Unlock()
		return err
	}
	w.file = file
	w.mu.Unlock()

	return nil
}

func (w *WAL) Cleanup() error {
	w.mu.Lock()
	fileName := w.file.Name()
	w.mu.Unlock()

	files, err := os.ReadDir(w.dir)
	if err != nil {
		return err
	}

	for _, file := range files {
		if file.IsDir() || w.dir+"/"+file.Name() == fileName {
			continue
		}

		os.Remove(fmt.Sprintf("%s/%s", w.dir, file.Name()))
	}

	return nil
}
