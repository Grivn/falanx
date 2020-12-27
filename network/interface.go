package network

import (
	"github.com/Grivn/libfalanx/common/protos"

	fCommonProto "github.com/ultramesh/flato-common/types/protos"
)

type Network interface {
	BroadcastTransactions(txs []*fCommonProto.Transaction)
	BroadcastOrderedReq(req *protos.OrderedReq)
	BroadcastOrderedLog(log *protos.OrderedLog)
	BroadcastSuspect(sus *protos.Suspect)
}
