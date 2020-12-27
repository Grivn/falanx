package types

type LocalBAEvent struct {
	TxHash          string
	MissingReplicas []uint64
}
