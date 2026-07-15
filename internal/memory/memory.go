package memory

import "github.com/emrecanterzi/wisp/internal/skiplist"

type Memory struct {
	skipList *skiplist.SkipList
}

func NewMemory(s *skiplist.SkipList) *Memory {
	return &Memory{
		skipList: s,
	}
}

func (m *Memory) Get(key string) (string, bool) {
	return m.skipList.Get(key)
}

func (m *Memory) Insert(key, value string) {
	m.skipList.Insert(key, value)
}

func (m *Memory) Delete(key string) bool {
	return m.skipList.Delete(key)
}
