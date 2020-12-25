package types

type SequenceLog struct {
	Seq uint64
	Tx  []byte
}
