package sequencepool

import (
	"container/list"
)

// LogList wraps a doubly-linked list and a map to achieve an
// ordered and retrievable data structure. flag presence is
// used to ensure no duplicate element can be put into pool
// at the same time.
// This data structure is mainly used for non-batched transactions.
type LogList struct {
	pool     *list.List
	presence map[string]*list.Element
}

// NewLogList initializes a new LogList with parameter size
// and returns its pointer.
func NewLogList() *LogList {
	return &LogList{
		pool:     list.New(),
		presence: make(map[string]*list.Element),
	}
}

// Len returns the length of the receiver list.
func (list *LogList) Len() int {
	return list.pool.Len()
}

// Has returns true if there is an element with key key
// in the receiver list.
func (list *LogList) Has(key string) bool {
	_, ok := list.presence[key]
	return ok
}

// Get returns the element corresponding to the key if such a key exists
// and nil otherwise.
func (list *LogList) Get(key string) *list.Element {
	e, ok := list.presence[key]
	if ok {
		return e
	}
	return nil
}

// PushBack inserts a new element with value value and key key at the back of list
// and returns the element, returns nil if there is a duplicate element..
func (list *LogList) PushBack(key string, value interface{}) *list.Element {
	if list.Has(key) {
		return nil
	}
	e := list.pool.PushBack(value)
	list.presence[key] = e
	return e
}

// PushFront inserts a new element with value value and key key at the front of list
// and returns the element, returns nil if there is a duplicate element.
func (list *LogList) PushFront(key string, value interface{}) *list.Element {
	if list.Has(key) {
		return nil
	}
	e := list.pool.PushFront(value)
	list.presence[key] = e
	return e
}

// InsertAfter inserts a new element with value value and key key
// immediately after mark and returns the element. If mark is not an element of list,
// the list is not modified and nil is returned. The mark must not be nil.
// Note: potential nil pointer dereference when mark is nil. Method is currently not used.
func (list *LogList) InsertAfter(key string, value interface{}, mark *list.Element) *list.Element {
	if list.Has(key) {
		return nil
	}
	e := list.pool.InsertAfter(value, mark)
	if e != nil {
		list.presence[key] = e
	}
	return e
}

// InsertBefore inserts a new element with value value and key key
// immediately before mark and returns the element. If mark is not an element of list,
// the list is not modified and nil is returned. The mark must not be nil.
// Note: potential nil pointer dereference when mark is nil. Method is currently not used.
func (list *LogList) InsertBefore(key string, value interface{}, mark *list.Element) *list.Element {
	if list.Has(key) {
		return nil
	}
	e := list.pool.InsertBefore(value, mark)
	if e != nil {
		list.presence[key] = e
	}
	return e
}

// Front returns the first element of list or nil if the LogList is empty.
func (list *LogList) Front() *list.Element {
	return list.pool.Front()
}

// Remove removes element e with key key from list if e is an element of list
// and Key,e is a valid key-value pair. It returns the element value e.Value.
// The element must not be nil.
func (list *LogList) Remove(e *list.Element, key string) error {
	if e == nil {
		return ErrNil
	}
	if !list.Has(key) {
		return ErrNotFoundElement
	}
	if list.Get(key) != e {
		return ErrMismatchElement
	}
	list.pool.Remove(e)
	delete(list.presence, key)
	return nil
}

