package falanx

import (
	pb "github.com/Grivn/libfalanx/zcommon/protos"
	"github.com/Grivn/libfalanx/zcommon/types"
	fCommonProto "github.com/ultramesh/flato-common/types/protos"
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
	go falanx.step(msg)
}

func (falanx *falanxImpl) Propose(txs []*fCommonProto.Transaction) {
	falanx.fakeClient.ProposeTxs(txs)
	return
}
