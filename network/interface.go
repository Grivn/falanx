package network

import (
	"github.com/Grivn/libfalanx/zcommon/protos"

	fCommonProto "github.com/ultramesh/flato-common/types/protos"
)

type Network interface {
	BroadcastTransactions(txs []*fCommonProto.Transaction)
	BroadcastOrderedReq(req *protos.OrderedReq)
	BroadcastOrderedLog(log *protos.OrderedLog)
}
