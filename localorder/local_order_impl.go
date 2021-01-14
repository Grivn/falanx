package localorder

import (
	"time"

	"github.com/Grivn/libfalanx/localorder/types"
	"github.com/Grivn/libfalanx/logger"
	"github.com/Grivn/libfalanx/network"
	"github.com/Grivn/libfalanx/zcommon/protos"
)

type localOrderImpl struct {
	id uint64
	seqNo uint64
	recvC chan string
	close chan bool
	network network.Network
	logger logger.Logger
}

func newLocalOrderImpl(c types.Config) *localOrderImpl {
	return &localOrderImpl{
		id:      c.ID,
		seqNo:   uint64(0),
		recvC:   c.RecvC,
		network: c.Network,
		logger:  c.Logger,
	}
}

func (local *localOrderImpl) start() {
	go local.listenTxHash()
}

func (local *localOrderImpl) stop() {
	close(local.close)
}

func (local *localOrderImpl) listenTxHash() {
	for {
		select {
		case <-local.close:
			return

		case txHash := <-local.recvC:
			local.order(txHash)
		}
	}
}

func (local *localOrderImpl) order(txHash string) {
	local.seqNo++
	log := &protos.OrderedLog{
		ReplicaId: local.id,
		Sequence:  local.seqNo,
		TxHash:    txHash,
		Timestamp: time.Now().UnixNano(),
	}
	local.network.BroadcastOrderedLog(log)
}
