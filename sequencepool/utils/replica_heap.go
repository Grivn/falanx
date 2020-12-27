package utils

import (
	"container/heap"

	"github.com/Grivn/libfalanx/common/protos"
)

// ================ CacheLog Interfaces ==================
type CacheLog interface {
	Len() int
	Has(seq uint64) bool
	Top() *protos.OrderedLog
	Push(r *protos.OrderedLog)
	Pop() *protos.OrderedLog
}

type cacheLogImpl struct {
	// heap is the container to order log
	heap *LogHeap

	// presence indicates whether a log exists
	presence map[uint64]*protos.OrderedLog
}

func NewLogCache() *cacheLogImpl {
	cache := &cacheLogImpl{
		heap:     &LogHeap{},
		presence: make(map[uint64]*protos.OrderedLog),
	}
	heap.Init(cache.heap)
	return cache
}

func (c *cacheLogImpl) Len() int {
	return c.heap.Len()
}

func (c *cacheLogImpl) Has(seq uint64) bool {
	_, ok := c.presence[seq]
	return ok
}

func (c *cacheLogImpl) Top() *protos.OrderedLog {
	r := c.heap.top()
	if r == nil {
		return nil
	}

	ret, ok := r.(*protos.OrderedLog)
	if !ok {
		return nil
	}
	return ret
}

func (c *cacheLogImpl) Push(r *protos.OrderedLog) {
	seq := r.Sequence
	if c.Has(seq) {
		return
	}

	c.heap.Push(r)
	c.presence[seq] = r
}

func (c *cacheLogImpl) Pop() *protos.OrderedLog {
	if c.heap.Len() == 0 {
		return nil
	}
	r, ok := c.heap.Pop().(*protos.OrderedLog)
	if !ok {
		return nil
	}
	delete(c.presence, r.Sequence)
	return r
}

// ======================= heap Interfaces ==============================
type LogHeap []*protos.OrderedLog

// Len is the number of elements in the collection.
func (h LogHeap) Len() int {
	return len(h)
}
// Less reports whether the element with index i should sort before the element with index j.
// Less here has initialized a minheap
func (h LogHeap) Less(i, j int) bool {
	return h[i].Sequence < h[j].Sequence
}
// Swap swaps the elements with indexes i and j.
func (h LogHeap) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
}
// Pop removes and returns the minimum element (according to Less) from the heap.
// The complexity is O(log n) where n = h.Len().
// Pop is equivalent to Remove(h, 0).
func (h *LogHeap) Pop() interface{} {
	old := *h
	n := h.Len()
	x := old[n-1]
	*h = old[0:n-1]
	return x
}
// Push pushes the element x onto the heap.
// The complexity is O(log n) where n = h.Len().
func (h *LogHeap) Push(x interface{}) {
	*h = append(*h, x.(*protos.OrderedLog))
}

// ======================= Essential Functions ==============================
func (h *LogHeap) top() interface{} {
	if h.Len() == 0 {
		return nil
	}
	old := *h
	n := h.Len()
	x := old[n-1]
	return x
}
