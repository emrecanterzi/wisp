package skiplist

import (
	"math/rand"
	"sync"

	"github.com/emrecanterzi/wisp/internal/types"
)

type node struct {
	op    types.Op
	key   string
	value string
	next  []*node
}

type SkipList struct {
	head     *node
	maxLevel int
	mu       sync.RWMutex
}

func NewSkipList(maxLevel int) *SkipList {
	return &SkipList{
		head:     &node{next: make([]*node, maxLevel)},
		maxLevel: maxLevel,
	}
}

func (s *SkipList) Get(key string) *types.Record {
	current := s.head

	s.mu.RLock()
	defer s.mu.RUnlock()
	for level := s.maxLevel - 1; level >= 0; level-- {
		for current.next[level] != nil && current.next[level].key < key {
			current = current.next[level]
		}
	}

	current = current.next[0]
	if current != nil && current.key == key {
		return &types.Record{
			Op:    current.op,
			Key:   current.key,
			Value: current.value,
		}
	}
	return nil
}

func (s *SkipList) set(key, value string, op types.Op) {
	s.mu.Lock()
	defer s.mu.Unlock()

	update := make([]*node, s.maxLevel)
	current := s.head

	for level := s.maxLevel - 1; level >= 0; level-- {
		for current.next[level] != nil && current.next[level].key < key {
			current = current.next[level]
		}
		update[level] = current
	}

	if current.next[0] != nil && current.next[0].key == key {
		current.next[0].value = value
		current.next[0].op = op
		return
	}

	l := 0
	for rand.Intn(2) == 0 && l < s.maxLevel-1 {
		l++
	}

	newNode := &node{key: key, value: value, op: op, next: make([]*node, l+1)}

	for level := 0; level <= l; level++ {
		newNode.next[level] = update[level].next[level]
		update[level].next[level] = newNode
	}
}

func (s *SkipList) Insert(key, value string) {
	s.set(key, value, types.OpSet)
}

func (s *SkipList) Delete(key string) {
	s.set(key, "", types.OpDelete)
}
func (s *SkipList) LoopAll(fn func(op types.Op, key, value string)) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	current := s.head.next[0]
	for current != nil {
		fn(current.op, current.key, current.value)
		current = current.next[0]
	}
}
