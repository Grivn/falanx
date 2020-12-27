package sequencepool

import (
	"github.com/Grivn/libfalanx/common"
	cTypes "github.com/Grivn/libfalanx/common/types"
	"github.com/Grivn/libfalanx/logger"
	"github.com/Grivn/libfalanx/sequencepool/types"
)

type sequencePoolImpl struct {
	author uint64

	timeoutC chan interface{}
	localBAC chan interface{}
	close    chan bool

	txs TxsContainer

	client  map[uint64]ClientOrder
	replica map[uint64]ReplicaOrder

	tools  common.Tools
	logger logger.Logger
}

func NewSequencePool(config types.Config) *sequencePoolImpl {
	txsContainer := NewTxContainer(config)
	return &sequencePoolImpl{
		author:   config.Author,
		timeoutC: config.TimeoutC,
		localBAC: config.LocalBAC,
		close:    make(chan bool),
		txs:      txsContainer,
		tools:    config.Tools,
		logger:   config.Logger,
	}
}

func (sp *sequencePoolImpl) start() {
	go sp.listenTimeoutEvent()
}

func (sp *sequencePoolImpl) stop() {
	close(sp.close)
}

func (sp *sequencePoolImpl) listenTimeoutEvent() {
	for {
		select {
		case <-sp.close:
			return
		case obj := <-sp.timeoutC:
			te, ok := obj.(types.TimeoutEvent)
			if !ok {
				sp.logger.Error("Parsing Error")
				return
			}
			sp.processTimeoutEvent(te)
		}
	}
}

func (sp *sequencePoolImpl) processTimeoutEvent(te types.TimeoutEvent) {
	le := cTypes.LocalBAEvent{
		TxHash: te.TxHash,
		MissingReplicas: nil,
	}
	sp.localBAC <- le
}
