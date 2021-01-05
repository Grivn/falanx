package utils

import (
	"container/list"
	"errors"

	"github.com/Grivn/libfalanx/zcommon/protos"
)

// Recorder List Interfaces ===========================================
type TxList interface {
	// list controller
	Add(l *protos.OrderedLog)
	Has(key string) bool
	Front() *list.Element
	Remove(e *list.Element, key string) error

	// value controller
	FrontLog() *protos.OrderedLog
	GetLog(key string) *protos.OrderedLog
	GetSequence(key string) (uint64, error)
	GetTimestamp(key string) (int64, error)
}
// ====================================================================

func NewTxList() *txListImpl {
	return newTxListImpl()
}

func (rr *txListImpl) Add(l *protos.OrderedLog) {
	rr.add(l)
}

func (rr *txListImpl) GetLog(key string) *protos.OrderedLog {
	return rr.getLog(key)
}

func (rr *txListImpl) GetSequence(key string) (uint64, error) {
	return rr.getSequence(key)
}

func (rr *txListImpl) GetTimestamp(key string) (int64, error) {
	return rr.getTimestamp(key)
}

func (rr *txListImpl) Has(key string) bool {
	return rr.has(key)
}

func (rr *txListImpl) Remove(e *list.Element, key string) error {
	return rr.remove(e, key)
}

func (rr *txListImpl) Front() *list.Element {
	return rr.front()
}

func (rr *txListImpl) FrontLog() *protos.OrderedLog {
	return rr.frontLog()
}

type txListImpl struct {
	// list contains the ordered logs from a replica
	list *list.List
	// presence indicates the elements in log list
	presence map[string]*list.Element
}

func newTxListImpl() *txListImpl {
	return &txListImpl{}
}

func (rr *txListImpl) add(l *protos.OrderedLog) {
	if rr.has(l.TxHash) {
		return
	}
	e := rr.list.PushBack(l)
	rr.presence[l.TxHash] = e
}

func (rr *txListImpl) get(key string) *list.Element {
	e, ok := rr.presence[key]
	if !ok {
		return nil
	}
	return e
}

func (rr *txListImpl) getLog(key string) *protos.OrderedLog {
	e := rr.get(key)
	if e == nil {
		return nil
	}
	r, ok := e.Value.(*protos.OrderedLog)
	if !ok {
		return nil
	}
	return r
}

func (rr *txListImpl) getSequence(key string) (uint64, error) {
	e := rr.getLog(key)
	if e == nil {
		return 0, errors.New("nil element")
	}
	return e.Sequence, nil
}

func (rr *txListImpl) getTimestamp(key string) (int64, error) {
	e := rr.getLog(key)
	if e == nil {
		return 0, errors.New("nil element")
	}
	return e.Timestamp, nil
}

func (rr *txListImpl) has(key string) bool {
	_, ok := rr.presence[key]
	return ok
}

func (rr *txListImpl) remove(e *list.Element, key string) error {
	if e == nil {
		return errors.New("nil element")
	}
	if !rr.has(key) {
		return errors.New("non-exited element")
	}
	if rr.get(key) != e {
		return errors.New("unpaired element")
	}
	rr.list.Remove(e)
	delete(rr.presence, key)
	return nil
}

func (rr *txListImpl) frontLog() *protos.OrderedLog {
	log, ok := rr.front().Value.(*protos.OrderedLog)
	if !ok {
		return nil
	}
	return log
}

func (rr *txListImpl) front() *list.Element {
	return rr.list.Front()
}

func (rr *txListImpl) pushBack(key string, value interface{}) *list.Element {
	if rr.has(key) {
		return nil
	}
	e := rr.list.PushBack(value)
	rr.presence[key] = e
	return e
}

