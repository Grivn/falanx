package sequencepool

import (
	"github.com/Grivn/libfalanx/sequencepool/types"
)

type SequencePool interface {
	// ReceiveTransactions is used to receive the transactions sent from clients
	ReceiveTransactions(tx []byte)

	// ReceiveReplicaLogs is used to receive the log order of other replicas
	ReceiveReplicaLogs(sl *types.SequenceLog)

	// LocalLogOrder is used to sort the transactions at local and generate the log order to send
	LocalLogOrder() *types.SequenceLog
}