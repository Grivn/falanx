package utils

import pb "github.com/Grivn/libfalanx/zcommon/protos"

type ClientRecorder interface {
	Counter() uint64
	Check(r *pb.OrderedReq) bool
	Update(r *pb.OrderedReq)
}

func NewClientRecorder() *clientRecorderImpl {
	return newClientRecorderImpl()
}

func (cr *clientRecorderImpl) Counter() uint64 {
	return cr.counter
}

func (cr *clientRecorderImpl) Check(r *pb.OrderedReq) bool {
	return cr.check(r)
}

func (cr *clientRecorderImpl) Update(r *pb.OrderedReq) {
	cr.update(r)
}

type clientRecorderImpl struct {
	// counter indicates the order of requests from client
	counter uint64

	// timestamp indicates the latest request's timestamp from client
	timestamp int64
}

func newClientRecorderImpl() *clientRecorderImpl {
	return &clientRecorderImpl{}
}

func (cr *clientRecorderImpl) check(r *pb.OrderedReq) bool {
	if r == nil {
		return false
	}
	return r.Sequence == cr.counter+1 && r.Timestamp > cr.timestamp
}

func (cr *clientRecorderImpl) update(r *pb.OrderedReq) {
	cr.counter++
	cr.timestamp = r.Timestamp
}
