package sequencepool

import (
	"github.com/Grivn/libfalanx/common"
	"github.com/Grivn/libfalanx/common/protos"
	"github.com/Grivn/libfalanx/logger"
	"github.com/Grivn/libfalanx/sequencepool/utils"
)

type ClientOrder interface {
	ReceiveOrderedReq(r *protos.OrderedReq)
}

func NewClientOrder(id uint64, tools common.Tools, logger logger.Logger) *clientOrderImpl {
	return newClientOrderImpl(id, tools, logger)
}

func (c *clientOrderImpl) ReceiveOrderedReq(r *protos.OrderedReq) {
	c.receiveOrderedRequest(r)
}

type clientOrderImpl struct {
	id uint64

	// request order interface ===================================================
	// there are two requests r1 and r2, the relation 'r1<-r2' is valid, iff:
	//   A1. r1.sequence+1 = r2.sequence
	//   A2. r1.timestamp < r2.timestamp
	// ===========================================================================
	cache    utils.CacheReq       // cache is used to store the requests which temporarily cannot be processed
	recorder utils.ClientRecorder // recorder si used to record the counter status of particular client

	// message channel ===========================================================
	orderC   chan string             // orderC is used to trigger local log sort

	// essential tools for ordered pool
	tools  common.Tools
	logger logger.Logger
}

func newClientOrderImpl(id uint64, tools common.Tools, logger logger.Logger) *clientOrderImpl {
	logger.Noticef("Initialize client order instance: [id]%d", id)
	return &clientOrderImpl{
		id:       id,
		cache:    utils.NewReqCache(),
		recorder: utils.NewClientRecorder(),
		orderC:   make(chan string),
		tools:    tools,
		logger:   logger,
	}
}

func (c *clientOrderImpl) receiveOrderedRequest(r *protos.OrderedReq) {
	if r == nil {
		c.logger.Warningf("Nil ordered request from client %d", c.id)
		return
	}
	if r.ClientId != c.id {
		c.logger.Warningf("Client %d received request from another client %d", c.id, r.ClientId)
		return
	}

	// store the request into the cache
	c.cacheRequest(r)

	// order the requests in the cache
	c.orderCachedRequests()
}

// cache is used to save the requests temporarily unable to process because of its sequence number
func (c *clientOrderImpl) cacheRequest(r *protos.OrderedReq) {
	if c.cache.Has(r.Sequence) {
		c.logger.Warningf("Duplicated req-sequence %d from client", r.Sequence)
		return
	}
	c.cache.Push(r)
}

func (c *clientOrderImpl) orderCachedRequests() uint64 {
	if c.cache.Len() == 0 {
		return c.recorder.Counter()
	}

	for c.recorder.Check(c.cache.Top()) {
		r := c.cache.Pop()
		c.recorder.Update(r)
		c.logger.Debugf("Read request cache of client %d, counter %d, tx %v",
			c.id, c.recorder.Counter(), r.TxHashList)
		c.postOrderedTxs(r.TxHashList)
	}
	return c.recorder.Counter()
}

func (c *clientOrderImpl) postOrderedTxs(list []string) {
	for _, txHash := range list {
		c.logger.Debugf("Post request %s from client %d", txHash, c.id)
		c.orderC <- txHash
	}
}