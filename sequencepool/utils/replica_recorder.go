package utils

import (
	"container/list"
	"errors"

	"github.com/Grivn/libfalanx/common/protos"
)

type ReplicaRecorder interface {
	Counter() uint64
	Check(r *protos.OrderedLog) bool
	Update(r *protos.OrderedLog)
	RecorderList
}

func NewReplicaRecorder() *replicaRecorderImpl {
	return newReplicaRecorderImpl()
}

func (rr *replicaRecorderImpl) Counter() uint64 {
	return rr.counter
}

func (rr *replicaRecorderImpl) Check(r *protos.OrderedLog) bool {
	return rr.check(r)
}

func (rr *replicaRecorderImpl) Update(r *protos.OrderedLog) {
	rr.update(r)
}

type replicaRecorderImpl struct {
	// counter indicates the order of logs from replica
	counter uint64

	// timestamp indicates the latest log's timestamp from replica
	timestamp int64

	// list recorder =============================================
	// record the order of a replica's logs
	// ===========================================================
	// list contains the ordered logs from a replica
	list *list.List
	// presence indicates the elements in log list
	presence map[string]*list.Element
}

func newReplicaRecorderImpl() *replicaRecorderImpl {
	return &replicaRecorderImpl{}
}

func (rr *replicaRecorderImpl) check(r *protos.OrderedLog) bool {
	return r.Sequence == rr.counter+1 && r.Timestamp > rr.timestamp
}

func (rr *replicaRecorderImpl) update(r *protos.OrderedLog) {
	rr.counter++
	rr.timestamp = r.Timestamp
	e := rr.list.PushBack(r)
	rr.presence[r.TxHash] = e
}

// Recorder List Interfaces ===========================================
type RecorderList interface {
	Get(key string) *list.Element
	GetLog(key string) *protos.OrderedLog
	GetSequence(key string) (uint64, error)
	GetTimestamp(key string) (int64, error)
	Has(key string) bool
	Remove(e *list.Element, key string) error
	Front() *list.Element
	PushBack(key string, value interface{}) *list.Element
}
// ====================================================================

func (rr *replicaRecorderImpl) Get(key string) *list.Element {
	return rr.get(key)
}

func (rr *replicaRecorderImpl) GetLog(key string) *protos.OrderedLog {
	return rr.getLog(key)
}

func (rr *replicaRecorderImpl) GetSequence(key string) (uint64, error) {
	return rr.getSequence(key)
}

func (rr *replicaRecorderImpl) GetTimestamp(key string) (int64, error) {
	return rr.getTimestamp(key)
}

func (rr *replicaRecorderImpl) Has(key string) bool {
	return rr.has(key)
}

func (rr *replicaRecorderImpl) Remove(e *list.Element, key string) error {
	return rr.remove(e, key)
}

func (rr *replicaRecorderImpl) Front() *list.Element {
	return rr.front()
}

func (rr *replicaRecorderImpl) PushBack(key string, value interface{}) *list.Element {
	return rr.pushBack(key, value)
}

func (rr *replicaRecorderImpl) get(key string) *list.Element {
	e, ok := rr.presence[key]
	if !ok {
		return nil
	}
	return e
}

func (rr *replicaRecorderImpl) getLog(key string) *protos.OrderedLog {
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

func (rr *replicaRecorderImpl) getSequence(key string) (uint64, error) {
	e := rr.getLog(key)
	if e == nil {
		return 0, errors.New("nil element")
	}
	return e.Sequence, nil
}

func (rr *replicaRecorderImpl) getTimestamp(key string) (int64, error) {
	e := rr.getLog(key)
	if e == nil {
		return 0, errors.New("nil element")
	}
	return e.Timestamp, nil
}

func (rr *replicaRecorderImpl) has(key string) bool {
	_, ok := rr.presence[key]
	return ok
}

func (rr *replicaRecorderImpl) remove(e *list.Element, key string) error {
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

func (rr *replicaRecorderImpl) front() *list.Element {
	return rr.list.Front()
}

func (rr *replicaRecorderImpl) pushBack(key string, value interface{}) *list.Element {
	if rr.has(key) {
		return nil
	}
	e := rr.list.PushBack(value)
	rr.presence[key] = e
	return e
}
