package localorder

import (
	"github.com/golang/protobuf/proto"
	"time"

	"github.com/Grivn/libfalanx/localorder/types"
	"github.com/Grivn/libfalanx/logger"
	"github.com/Grivn/libfalanx/network"
	pb "github.com/Grivn/libfalanx/zcommon/protos"
)

type localOrderImpl struct {
	id uint64
	seqNo uint64
	recvC chan string
	selfC chan *pb.OrderedLog
	close chan bool
	network network.Network
	logger logger.Logger
}

func newLocalOrderImpl(c types.Config) *localOrderImpl {
	return &localOrderImpl{
		id:      c.ID,
		seqNo:   uint64(0),
		recvC:   c.RecvC,
		selfC:   c.SelfC,
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
	log := &pb.OrderedLog{
		ReplicaId: local.id,
		Sequence:  local.seqNo,
		TxHash:    txHash,
		Timestamp: time.Now().UnixNano(),
	}
	logPayload, err := proto.Marshal(log)
	if err != nil {
		return
	}
	logMsg := &pb.ConsensusMessage{
		Type:    pb.Type_ORDERED_LOG,
		Payload: logPayload,
	}
	local.network.Broadcast(logMsg)
	local.logger.Noticef("Replica %d broadcast local order: seq %d, hash %s", local.id, local.seqNo, txHash)
	local.inform(log)
}

func (local *localOrderImpl) inform(log *pb.OrderedLog) {
	local.selfC <- log
}
