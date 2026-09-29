package wal

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"log"
	"math"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

const walMagic = "WSPW"
const walVersion byte = 1
const headerLen = len(walMagic) + 1

var (
	ErrInvalidMagic   = errors.New("invalid WAL magic header")
	ErrInvalidVersion = errors.New("unsupported WAL version")
)

type WAL struct {
	dir    string
	file   *os.File
	mu     sync.Mutex
	broken bool
}

func NewWAL(dir string) (*WAL, error) {
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		return nil, err
	}

	file, err := createWALFile(dir)
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
	if w.broken {
		return errors.New("wal: broken after failed cleanup, refusing writes")
	}
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

	stat, err := w.file.Stat()
	if err != nil {
		return err
	}
	sizeBef := stat.Size()

	_, err = w.file.Write(buf.Bytes())
	if err == nil {
		err = w.file.Sync()
	}
	if err != nil {
		if terr := w.file.Truncate(sizeBef); terr != nil {
			w.broken = true
		} else if serr := w.file.Sync(); serr != nil {
			w.broken = true
		}
		return err
	}
	return nil
}

func (w *WAL) ReplyWals(fn func(operation uint8, key, value []byte)) error {
	files, err := os.ReadDir(w.dir)
	if err != nil {
		return err
	}

	filteredFiles := []os.DirEntry{}

	for _, entry := range files {
		if entry.IsDir() {
			continue
		}
		if !strings.HasSuffix(entry.Name(), ".wal") {
			continue
		}
		if filepath.Base(w.file.Name()) == entry.Name() {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		if info.Size() <= int64(headerLen) {
			os.Remove(filepath.Join(w.dir, entry.Name()))
			log.Printf("wal: removing empty WAL file %s (%d bytes)", entry.Name(), info.Size())
			continue
		}

		filteredFiles = append(filteredFiles, entry)
	}

	for i, entry := range filteredFiles {
		path := filepath.Join(w.dir, entry.Name())

		file, err := os.Open(path)
		if err != nil {
			return err
		}
		finfo, err := file.Stat()
		if err != nil {
			file.Close()
			return err
		}
		fileSize := finfo.Size()

		validSize, err := w.replay(file, fileSize, fn)
		if err != nil {
			file.Close()
			return err
		}
		if err := file.Close(); err != nil {
			return err
		}

		if validSize < fileSize {
			if len(filteredFiles)-1 == i {

				if err := os.Truncate(path, validSize); err != nil {
					return err
				}
				log.Printf("wal: trimmed torn tail of %s: cut %d bytes (%d -> %d)", entry.Name(), fileSize-validSize, fileSize, validSize)
			} else {
				return fmt.Errorf("wal: %s is corrupted, but it is not the last file (valid %d bytes, file %d bytes)", entry.Name(), validSize, fileSize)
			}
		}
	}

	return nil
}

func (w *WAL) replay(file *os.File, fileSize int64, fn func(operation uint8, key, value []byte)) (int64, error) {
	_, err := readHeader(file)
	if err != nil {
		return 0, err
	}

	validEnd := int64(headerLen)

	opByte := make([]byte, 1)
	keyLenBuf := make([]byte, 4)
	var keyBuf []byte
	valueLenBuf := make([]byte, 4)
	var valueBuf []byte

	for {
		_, err := io.ReadFull(file, opByte)
		if err != nil {
			if err == io.EOF {
				break
			}

			return validEnd, err
		}

		operation := opByte[0]

		_, err = io.ReadFull(file, keyLenBuf)
		if err != nil {
			if isTorn(err) {
				break
			}
			return validEnd, err
		}
		keyLen := binary.BigEndian.Uint32(keyLenBuf)
		if validEnd+1+4+int64(keyLen) > fileSize {
			break
		}

		keyBuf = make([]byte, keyLen)

		_, err = io.ReadFull(file, keyBuf)
		if err != nil {
			if isTorn(err) {
				break
			}
			return validEnd, err
		}

		valueBuf = nil
		if operation == 1 {
			_, err = io.ReadFull(file, valueLenBuf)
			if err != nil {
				if isTorn(err) {
					break
				}
				return validEnd, err
			}
			valueLen := binary.BigEndian.Uint32(valueLenBuf)
			if validEnd+1+4+int64(keyLen)+4+int64(valueLen) > fileSize {
				break
			}

			valueBuf = make([]byte, valueLen)

			_, err = io.ReadFull(file, valueBuf)
			if err != nil {
				if isTorn(err) {
					break
				}
				return validEnd, err
			}

			validEnd += (1 + 4 + int64(len(keyBuf)) + 4 + int64(len(valueBuf)))
		} else {
			validEnd += (1 + 4 + int64(len(keyBuf)))
		}

		fn(operation, keyBuf, valueBuf)
	}

	return validEnd, nil
}

func (w *WAL) Rotate() error {
	w.mu.Lock()
	err := w.file.Close()
	if err != nil {
		w.mu.Unlock()
		return err
	}
	file, err := createWALFile(w.dir)
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

func readHeader(r io.Reader) (byte, error) {
	var header [headerLen]byte

	if _, err := io.ReadFull(r, header[:]); err != nil {
		return 0, fmt.Errorf("failed to read WAL header: %w", err)
	}

	if string(header[0:4]) != walMagic {
		return 0, ErrInvalidMagic
	}

	if header[4] != walVersion {
		return 0, fmt.Errorf("%w: got %d, want %d", ErrInvalidVersion, header[4], walVersion)
	}

	return header[4], nil
}

func createWALFile(dir string) (*os.File, error) {
	fileName := fmt.Sprintf("%s/%d.wal", dir, time.Now().UnixNano())
	file, err := os.OpenFile(fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}

	header := append([]byte(walMagic), walVersion)

	_, err = file.Write(header)
	if err != nil {
		_ = file.Close()
		_ = os.Remove(fileName)

		return nil, err
	}

	err = file.Sync()
	if err != nil {
		_ = file.Close()
		_ = os.Remove(fileName)

		return nil, err
	}

	return file, nil
}

func isTorn(err error) bool {
	return errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF)
}
