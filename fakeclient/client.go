package fakeclient

import (
	"github.com/Grivn/libfalanx/fakeclient/types"

	fCommonProto "github.com/ultramesh/flato-common/types/protos"
)

func NewClient(c types.Config) *clientImpl {
	return newClientImpl(c)
}

func (c *clientImpl) ProposeTxs(txs []*fCommonProto.Transaction) {
	c.propose(txs)
}
