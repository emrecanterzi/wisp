package sstable

import (
	"bytes"
	"encoding/binary"
	"io"

	"github.com/emrecanterzi/wisp/internal/types"
)

func Encode(rec types.Record) []byte {
	var buf bytes.Buffer

	_ = binary.Write(&buf, binary.BigEndian, rec.Op)
	_ = binary.Write(&buf, binary.BigEndian, int32(len(rec.Key)))
	_, _ = buf.Write([]byte(rec.Key))

	if rec.Op != types.OpDelete {
		_ = binary.Write(&buf, binary.BigEndian, int32(len(rec.Value)))
		_, _ = buf.Write([]byte(rec.Value))
	}

	return buf.Bytes()
}

func Decode(data []byte) (types.Record, error) {
	record := types.Record{}
	reader := bytes.NewReader(data)

	opBuf := make([]byte, 1)
	_, err := io.ReadFull(reader, opBuf)
	if err != nil {
		return record, err
	}
	op := types.Op(opBuf[0])
	record.Op = op

	keyLenBuf := make([]byte, 4)
	_, err = io.ReadFull(reader, keyLenBuf)
	if err != nil {
		return record, err
	}

	keyBuf := make([]byte, binary.BigEndian.Uint32(keyLenBuf))
	_, err = io.ReadFull(reader, keyBuf)
	if err != nil {
		return record, err
	}
	record.Key = string(keyBuf)

	if op != types.OpDelete {
		valLenBuf := make([]byte, 4)
		_, err = io.ReadFull(reader, valLenBuf)
		if err != nil {
			return record, err
		}

		valBuf := make([]byte, binary.BigEndian.Uint32(valLenBuf))
		_, err = io.ReadFull(reader, valBuf)
		if err != nil {
			return record, err
		}

		record.Value = string(valBuf)
	}

	return record, nil
}
