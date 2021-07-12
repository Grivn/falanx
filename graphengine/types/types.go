package types

type TxSet map[string]bool

type TxInfo struct {
	Hash   string
	SeqSet map[uint64]uint64
}

type Finality bool

type Graph map[string]map[string]bool

type IDMap map[uint64]string

type V struct {
	DFN    uint64
	Low    uint64
	Pushed bool
}

const (
	DefaultGraphSize = 50
)
