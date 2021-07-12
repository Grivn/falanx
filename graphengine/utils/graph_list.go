package utils

import (
	"sync"
)

type TarjanStack interface {
	Len() int
	Peek() uint64
	Pop() uint64
	Push(id uint64)
}
func NewTarjanStack() *tarjanStack {
	return newTarjanStack()
}
func (s *tarjanStack) Len() int {
	return s.len()
}
func (s *tarjanStack) Peek() uint64 {
	return s.peek()
}
func (s *tarjanStack) Pop() uint64 {
	return s.pop()
}
func (s *tarjanStack) Push(hash uint64) {
	s.push(hash)
}

type (
	tarjanStack struct {
	top      *node
	length   int
	lock     *sync.RWMutex
}
	node struct {
		id   uint64
		prev *node
	}
)

func newTarjanStack() *tarjanStack {
	return &tarjanStack{
		top:      nil,
		length:   0,
		lock:     &sync.RWMutex{},
	}
}
func (s *tarjanStack) len() int {
	return s.length
}
func (s *tarjanStack) peek() uint64 {
	if s.length == 0 {
		return 0
	}
	return s.top.id
}
func (s *tarjanStack) pop() uint64 {
	s.lock.Lock()
	defer s.lock.Unlock()
	if s.length == 0 {
		return 0
	}
	n := s.top
	s.top = n.prev
	s.length--
	return n.id
}
func (s *tarjanStack) push(id uint64) {
	s.lock.Lock()
	defer s.lock.Unlock()
	n := &node{
		id: id,
	}
	s.top = n
	s.length++
}
