package memory

import (
	"github.com/emrecanterzi/wisp/internal/skiplist"
	"github.com/emrecanterzi/wisp/internal/wal"
)

type Memory struct {
	skipList *skiplist.SkipList
	wal      *wal.WAL
}

func NewMemory(s *skiplist.SkipList, w *wal.WAL) *Memory {
	return &Memory{
		skipList: s,
		wal:      w,
	}
}

func (m *Memory) Get(key string) (string, bool) {
	return m.skipList.Get(key)
}

func (m *Memory) Insert(key, value string) error {
	err := m.wal.Write(1, []byte(key), []byte(value))
	if err != nil {
		return err
	}
	m.skipList.Insert(key, value)
	return nil
}

func (m *Memory) Delete(key string) (bool, error) {
	err := m.wal.Write(0, []byte(key), nil)
	if err != nil {
		return false, err
	}
	return m.skipList.Delete(key), nil
}
