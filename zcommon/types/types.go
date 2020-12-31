package types

type Peer struct {
	ID   uint64
	Hash string
}

type LocalBAEvent struct {
	TxHash          string
	MissingReplicas []uint64
}
