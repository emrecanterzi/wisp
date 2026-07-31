package memory

import (
	"log"
	"sync"
	"time"

	"github.com/emrecanterzi/wisp/internal/skiplist"
	"github.com/emrecanterzi/wisp/internal/sstable"
	"github.com/emrecanterzi/wisp/internal/types"
	"github.com/emrecanterzi/wisp/internal/wal"
)

type Memory struct {
	skipList        *skiplist.SkipList
	frozenSkipList  *skiplist.SkipList
	sm              *sstable.SSTableManager
	wal             *wal.WAL
	mu              sync.RWMutex
	insertCount     int
	flushInProgress bool
	retrySame       bool
}

func NewMemory(w *wal.WAL, sm *sstable.SSTableManager) *Memory {
	return &Memory{
		skipList: skiplist.NewSkipList(4),
		sm:       sm,
		wal:      w,
	}
}
func (m *Memory) Get(key string) (string, bool, error) {
	m.mu.RLock()
	sl := m.skipList
	frozen := m.frozenSkipList
	m.mu.RUnlock()

	rec := sl.Get(key)

	if rec == nil && frozen != nil {
		rec = frozen.Get(key)
	}

	if rec == nil {
		var err error
		rec, err = m.sm.Search(key)
		if err != nil {
			return "", false, err
		}
	}

	if rec == nil {
		return "", false, nil
	}

	if rec.Op == types.OpDelete {
		return "", false, nil
	}

	return rec.Value, true, nil
}

func (m *Memory) Insert(key, value string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	err := m.wal.Write(1, []byte(key), []byte(value))
	if err != nil {
		return err
	}
	log.Println(key, value)
	m.skipList.Insert(key, value)

	m.insertCount++

	return nil
}

func (m *Memory) Delete(key string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	err := m.wal.Write(0, []byte(key), nil)
	if err != nil {
		return err
	}

	m.skipList.Delete(key)
	return nil
}

func (m *Memory) Startup() error {
	count := 0
	err := m.wal.ReplyWals(func(operation uint8, key, value []byte) {
		m.applyRecord(operation, key, value)
		count++
	})
	if err != nil {
		return err
	}
	log.Printf("wal replay complete: %d records applied", count)

	go func() {
		ticker := time.NewTicker(10 * time.Second)

		for range ticker.C {
			m.mu.RLock()
			count := m.insertCount
			flushInProgress := m.flushInProgress
			m.mu.RUnlock()

			if count > 1000 && flushInProgress == false {
				log.Println("flush triggered")
				err := m.flushSSTable()
				if err != nil {
					log.Println(err)
				}
			}
		}
	}()

	return nil
}

func (m *Memory) applyRecord(operation uint8, key []byte, value []byte) {
	switch operation {
	case 0:
		m.skipList.Delete(string(key))
	case 1:
		m.skipList.Insert(string(key), string(value))
	}

}

func (m *Memory) flushSSTable() error {
	m.mu.Lock()
	frozen := m.frozenSkipList
	if !m.retrySame {
		frozen = m.skipList
		m.frozenSkipList = frozen
		m.skipList = skiplist.NewSkipList(4)
		err := m.wal.Rotate()
		if err != nil {
			m.mu.Unlock()
			return err
		}
	}
	m.insertCount = 0
	m.flushInProgress = true
	m.mu.Unlock()

	err := m.sm.Flush(frozen)
	if err != nil {
		m.mu.Lock()
		m.retrySame = true
		m.flushInProgress = false
		m.mu.Unlock()
		return err
	}

	m.mu.Lock()
	m.flushInProgress = false
	m.retrySame = false
	err = m.wal.Cleanup()
	if err != nil {
		m.mu.Unlock()
		return err
	}
	m.mu.Unlock()

	return nil
}
