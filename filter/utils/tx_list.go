package utils

import (
	"container/list"
	"errors"

	pb "github.com/Grivn/libfalanx/zcommon/protos"
)

// Recorder List Interfaces ===========================================
type TxList interface {
	// list controller
	Add(l *pb.OrderedLog)
	Has(key string) bool
	Front() *list.Element
	Remove(e *list.Element, key string) error
	Len() int

	// value controller
	FrontLog() *pb.OrderedLog
	GetLog(key string) *pb.OrderedLog
	GetSequence(key string) (uint64, error)
	GetTimestamp(key string) (int64, error)
	RemoveLog(txHash string)
}
// ====================================================================

func NewTxList() *txListImpl {
	return newTxListImpl()
}

func (list *txListImpl) Add(l *pb.OrderedLog) {
	list.add(l)
}

func (list *txListImpl) GetLog(key string) *pb.OrderedLog {
	return list.getLog(key)
}

func (list *txListImpl) GetSequence(key string) (uint64, error) {
	return list.getSequence(key)
}

func (list *txListImpl) GetTimestamp(key string) (int64, error) {
	return list.getTimestamp(key)
}

func (list *txListImpl) Has(key string) bool {
	return list.has(key)
}

func (list *txListImpl) Remove(e *list.Element, key string) error {
	return list.remove(e, key)
}

func (list *txListImpl) Front() *list.Element {
	return list.front()
}

func (list *txListImpl) FrontLog() *pb.OrderedLog {
	return list.frontLog()
}

func (list *txListImpl) Len() int {
	return list.len()
}

func (list *txListImpl) RemoveLog(txHash string) {
	list.removeLog(txHash)
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

func (list *txListImpl) add(l *pb.OrderedLog) {
	if list.has(l.TxHash) {
		return
	}
	e := list.list.PushBack(l)
	list.presence[l.TxHash] = e
}

func (list *txListImpl) get(key string) *list.Element {
	e, ok := list.presence[key]
	if !ok {
		return nil
	}
	return e
}

func (list *txListImpl) getLog(key string) *pb.OrderedLog {
	e := list.get(key)
	if e == nil {
		return nil
	}
	r, ok := e.Value.(*pb.OrderedLog)
	if !ok {
		return nil
	}
	return r
}

func (list *txListImpl) getSequence(key string) (uint64, error) {
	e := list.getLog(key)
	if e == nil {
		return 0, errors.New("nil element")
	}
	return e.Sequence, nil
}

func (list *txListImpl) getTimestamp(key string) (int64, error) {
	e := list.getLog(key)
	if e == nil {
		return 0, errors.New("nil element")
	}
	return e.Timestamp, nil
}

func (list *txListImpl) has(key string) bool {
	_, ok := list.presence[key]
	return ok
}

func (list *txListImpl) remove(e *list.Element, key string) error {
	if e == nil {
		return errors.New("nil element")
	}
	if !list.has(key) {
		return errors.New("non-exited element")
	}
	if list.get(key) != e {
		return errors.New("unpaired element")
	}
	list.list.Remove(e)
	delete(list.presence, key)
	return nil
}

func (list *txListImpl) frontLog() *pb.OrderedLog {
	log, ok := list.front().Value.(*pb.OrderedLog)
	if !ok {
		return nil
	}
	return log
}

func (list *txListImpl) front() *list.Element {
	return list.list.Front()
}

func (list *txListImpl) pushBack(key string, value interface{}) *list.Element {
	if list.has(key) {
		return nil
	}
	e := list.list.PushBack(value)
	list.presence[key] = e
	return e
}

func (list *txListImpl) len() int {
	if list.presence == nil {
		return 0
	}
	return len(list.presence)
}

func (list *txListImpl) removeLog(txHash string) {
	if !list.has(txHash) {
		return
	}
	e := list.get(txHash)
	err := list.remove(e, txHash)
	if err != nil {
		panic(err)
		return
	}
}
