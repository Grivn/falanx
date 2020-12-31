package creplicasorder

import (
	"github.com/Grivn/libfalanx/creplicasorder/utils"
	"github.com/Grivn/libfalanx/logger"
	"github.com/Grivn/libfalanx/zcommon"
	"github.com/Grivn/libfalanx/zcommon/protos"
)

func NewReplicaOrder(id uint64, tools zcommon.Tools, logger logger.Logger) *replicaOrderImpl {
	return newReplicaOrderImpl(id, tools, logger)
}

func (r *replicaOrderImpl) ReceiveOrderedReq(l *protos.OrderedLog){
	r.receiveOrderedLogs(l)
}

type replicaOrderImpl struct {
	id uint64

	// log order interface =======================================================
	// ===========================================================================
	cache    utils.CacheLog        // cache is used to store the logs which temporarily cannot be processed
	recorder utils.ReplicaRecorder // recorder si used to record the counter status of particular replica

	// essential tools ===========================================================
	tools  zcommon.Tools
	logger logger.Logger
}

func newReplicaOrderImpl(id uint64, tools zcommon.Tools, logger logger.Logger) *replicaOrderImpl {
	return &replicaOrderImpl{
		id:       id,
		cache:    utils.NewLogCache(),
		recorder: utils.NewReplicaRecorder(),
		tools:    tools,
		logger:   logger,
	}
}

func (r *replicaOrderImpl) receiveOrderedLogs(l *protos.OrderedLog) {
	if r == nil {
		r.logger.Warningf("Nil ordered request from client %d", r.id)
		return
	}
	if l.ReplicaId != r.id {
		r.logger.Warningf("Client %d received request from another client %d", r.id, l.ReplicaId)
		return
	}

	// store the request into the cache
	r.cacheRequest(l)

	// order the requests in the cache
	r.orderCachedRequests()
}

// cacheRequest is used to save the requests temporarily unable to process because of its sequence number
func (r *replicaOrderImpl) cacheRequest(l *protos.OrderedLog) {
	if r.cache.Has(l.Sequence) {
		r.logger.Warningf("Duplicated log-sequence %d from replica", l.Sequence)
		return
	}
	r.cache.Push(l)
}

func (r *replicaOrderImpl) orderCachedRequests() uint64 {
	if r.cache.Len() == 0 {
		return r.recorder.Counter()
	}

	for r.recorder.Check(r.cache.Top()) {
		l := r.cache.Pop()
		r.recorder.Update(l)
		r.logger.Debugf("Read log cache of replica %d, counter %d, tx %v",
			r.id, r.recorder.Counter(), l.TxHash)
	}
	return r.recorder.Counter()
}
