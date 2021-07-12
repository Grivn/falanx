package falanx

import (
	pb "github.com/Grivn/libfalanx/zcommon/protos"
	"github.com/Grivn/libfalanx/zcommon/types"
)

func NewFalanx(c types.Config) *falanxImpl {
	return newFalanxImpl(c)
}

func (falanx *falanxImpl) StartFalanx() {
	falanx.start()
}

func (falanx *falanxImpl) StopFalanx() {
	falanx.stop()
}

func (falanx *falanxImpl) StepMessage(msg *pb.ConsensusMessage) {
	falanx.step(msg)
}

func (falanx *falanxImpl) Propose(txs []*pb.Transaction) {
	falanx.forwardClient.ProposeTxs(txs)
	return
}
