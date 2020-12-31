package atxcontainer

import fCommonProto "github.com/ultramesh/flato-common/types/protos"

// Interface ================================================================
// TxsContainer is only used to receive the transactions and their status
// we don't need to maintain the order here and it is only a container for transactions
// no duplicated transactions
type TxsContainer interface {
	Add(tx *fCommonProto.Transaction)
	Get(txHash string) *fCommonProto.Transaction
	Remove(txHash string) error
}
