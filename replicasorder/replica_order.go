package replicasorder

import "github.com/Grivn/libfalanx/zcommon/protos"

type ReplicaOrder interface {
	ReceiveOrderedReq(l *protos.OrderedLog)
}
