package skiplist

import (
	"math/rand"
	"sync"
)

type node struct {
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

func (s *SkipList) Get(key string) (string, bool) {
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
		return current.value, true
	}
	return "", false
}

func (s *SkipList) Insert(key, value string) {
	l := 0
	for rand.Intn(2) == 0 && l < s.maxLevel-1 {
		l++
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	newNode := &node{key: key, value: value, next: make([]*node, l+1)}

	current := s.head

	for level := s.maxLevel - 1; level >= 0; level-- {
		for current.next[level] != nil && current.next[level].key < key {
			current = current.next[level]
		}

		if level <= l {
			newNode.next[level] = current.next[level]
			current.next[level] = newNode
		}
	}
}

func (s *SkipList) Delete(key string) bool {
	isDeleted := false
	current := s.head

	s.mu.Lock()
	defer s.mu.Unlock()

	for level := s.maxLevel - 1; level >= 0; level-- {
		for current.next[level] != nil && current.next[level].key < key {
			current = current.next[level]
		}

		if current.next[level] != nil && current.next[level].key == key {
			current.next[level] = current.next[level].next[level]
			isDeleted = true
		}
	}

	return isDeleted
}
