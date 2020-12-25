package types

import (
	cTypes "github.com/Grivn/libfalanx/common/types"
)

type Transaction struct {
	Hash    string
	Content cTypes.Transaction
}

type OrderedRequest struct {
	Sequence   uint64
	TxHashList []string
}

type OrderedLog struct {
	Sequence uint64
	TxHash   string
}

type ReplicaRecorder struct {
	Author cTypes.Author

	OrderMap map[uint64]string

	MaxID uint64
}

type TxRecorder struct {
	Transaction *cTypes.Transaction
	OrderedReplicas []cTypes.Author
}
