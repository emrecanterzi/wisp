package sstable

import (
	"fmt"
	"os"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/emrecanterzi/wisp/internal/skiplist"
	"github.com/emrecanterzi/wisp/internal/types"
)

type SSTableManager struct {
	dir     string
	entries []*entry
	mu      sync.RWMutex
}

func NewSSTable(dir string) (*SSTableManager, error) {
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		return nil, err
	}

	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var entries []*entry

	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".data") {
			filename := strings.Split(file.Name(), ".data")[0]

			entries = append(entries, &entry{
				dataFile:  dir + "/" + filename + ".data",
				indexFile: dir + "/" + filename + ".index",
			})
		}
	}

	slices.Reverse(entries)

	return &SSTableManager{dir: dir, entries: entries}, nil
}

func (sm *SSTableManager) Flush(sl *skiplist.SkipList) error {
	t := fmt.Sprintf("%d", time.Now().UnixNano())
	file, err := os.Create(sm.dir + "/sstable_" + t + ".data")
	if err != nil {
		return err
	}
	defer file.Close()

	indexFile, err := os.Create(sm.dir + "/sstable_" + t + ".index")
	if err != nil {
		return err
	}
	defer indexFile.Close()

	fs := &flushState{file: file, indexFile: indexFile}

	var callbackErr error
	sl.LoopAll(func(op types.Op, key, value string) {
		if callbackErr != nil {
			return
		}
		rec := types.Record{Op: op, Key: key, Value: value}

		callbackErr = fs.writeEntry(rec)
	})

	if callbackErr != nil {
		return callbackErr
	}

	if err := fs.persistIndex(); err != nil {
		return err
	}

	sm.mu.Lock()
	sm.entries = append([]*entry{{
		dataFile:  sm.dir + "/sstable_" + t + ".data",
		indexFile: sm.dir + "/sstable_" + t + ".index",
	}}, sm.entries...)
	sm.mu.Unlock()

	return nil
}

func (sm *SSTableManager) Search(key string) (*types.Record, error) {
	sm.mu.RLock()
	entries := sm.entries
	sm.mu.RUnlock()

	for _, entry := range entries {
		bestIdx := -1

		if !entry.loaded {
			err := entry.loadIndex()
			if err != nil {
				return nil, err
			}
		}

		left, right := 0, len(entry.index)-1

		for left <= right {
			mid := left + (right-left)/2

			if entry.index[mid].Key <= key {
				bestIdx = mid
				left = mid + 1
			} else if entry.index[mid].Key > key {
				right = mid - 1
			}
		}

		var startOffset, endOffset int64

		if bestIdx == -1 {
			startOffset = 0
			if len(entry.index) == 0 {
				endOffset = -1
			} else {
				endOffset = entry.index[0].Offset
			}
		} else {
			startOffset = entry.index[bestIdx].Offset
			if bestIdx == len(entry.index)-1 {
				endOffset = -1
			} else {
				endOffset = entry.index[bestIdx+1].Offset
			}
		}

		rec, err := entry.scan(startOffset, endOffset, key)
		if err != nil || rec != nil {
			return rec, err
		}
	}

	return nil, nil
}
