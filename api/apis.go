package api

import pb "github.com/Grivn/libfalanx/zcommon/protos"

type ModuleControl interface {
	Start()
	Stop()
}

// TxsContainer is only used to receive the transactions and their status
// we don't need to maintain the order here and it is only a container for transactions
// no duplicated transactions
type TxsContainer interface {
	Add(tx *pb.Transaction)
	Get(txHash string) *pb.Transaction
	Remove(txHash string) error
}

type ForwardClient interface {
	ProposeTxs(txs []*pb.Transaction)
}
