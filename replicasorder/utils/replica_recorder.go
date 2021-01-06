package utils

import (
	"github.com/Grivn/libfalanx/zcommon/protos"
)

type ReplicaRecorder interface {
	Counter() uint64
	Check(r *protos.OrderedLog) bool
	Update(r *protos.OrderedLog)
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
}
