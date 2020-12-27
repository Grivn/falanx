package utils

import (
	"container/heap"

	"github.com/Grivn/libfalanx/common/protos"
)

// ================ CacheReq Interfaces ==================
type CacheReq interface {
	Len() int
	Has(seq uint64) bool
	Top() *protos.OrderedReq
	Push(r *protos.OrderedReq)
	Pop() *protos.OrderedReq
}

type cacheReqImpl struct {
	// heap is the container to order requests
	heap *ReqHeap

	// presence indicates whether a request exists
	presence map[uint64]*protos.OrderedReq
}

func NewReqCache() *cacheReqImpl {
	cache := &cacheReqImpl{
		heap:     &ReqHeap{},
		presence: make(map[uint64]*protos.OrderedReq),
	}
	heap.Init(cache.heap)
	return cache
}

func (c *cacheReqImpl) Len() int {
	return c.heap.Len()
}

func (c *cacheReqImpl) Has(seq uint64) bool {
	_, ok := c.presence[seq]
	return ok
}

func (c *cacheReqImpl) Top() *protos.OrderedReq {
	r := c.heap.top()
	if r == nil {
		return nil
	}

	ret, ok := r.(*protos.OrderedReq)
	if !ok {
		return nil
	}
	return ret
}

func (c *cacheReqImpl) Push(r *protos.OrderedReq) {
	seq := r.Sequence
	if c.Has(seq) {
		return
	}

	c.heap.Push(r)
	c.presence[seq] = r
}

func (c *cacheReqImpl) Pop() *protos.OrderedReq {
	if c.heap.Len() == 0 {
		return nil
	}
	r, ok := c.heap.Pop().(*protos.OrderedReq)
	if !ok {
		return nil
	}
	delete(c.presence, r.Sequence)
	return r
}

// ======================= heap Interfaces ==============================
type ReqHeap []*protos.OrderedReq

// Len is the number of elements in the collection.
func (h ReqHeap) Len() int {
	return len(h)
}
// Less reports whether the element with index i should sort before the element with index j.
// Less here has initialized a minheap
func (h ReqHeap) Less(i, j int) bool {
	return h[i].Sequence < h[j].Sequence
}
// Swap swaps the elements with indexes i and j.
func (h ReqHeap) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
}
// Pop removes and returns the minimum element (according to Less) from the heap.
// The complexity is O(log n) where n = h.Len().
// Pop is equivalent to Remove(h, 0).
func (h *ReqHeap) Pop() interface{} {
	old := *h
	n := h.Len()
	x := old[n-1]
	*h = old[0:n-1]
	return x
}
// Push pushes the element x onto the heap.
// The complexity is O(log n) where n = h.Len().
func (h *ReqHeap) Push(x interface{}) {
	*h = append(*h, x.(*protos.OrderedReq))
}

// ======================= Essential Functions ==============================
func (h *ReqHeap) top() interface{} {
	if h.Len() == 0 {
		return nil
	}
	old := *h
	n := h.Len()
	x := old[n-1]
	return x
}
