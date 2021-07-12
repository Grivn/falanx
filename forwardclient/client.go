package forwardclient

import (
	"github.com/Grivn/libfalanx/forwardclient/types"
	pb "github.com/Grivn/libfalanx/zcommon/protos"
)

func NewClient(c types.Config) *clientImpl {
	return newClientImpl(c)
}

func (c *clientImpl) ProposeTxs(txs []*pb.Transaction) {
	c.propose(txs)
}
