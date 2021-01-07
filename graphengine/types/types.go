package types

import "github.com/Grivn/libfalanx/zcommon/protos"

type Graph [][]V

type Type string

const (
	TypeGraphPassedTxs   = "type_graph_passed_txs"
	TypeGraphTxsRelation = "type_graph_txs_relation"
)

type GraphEvent struct {
	Type Type
	Event interface{}
}

type V struct {
	ID     uint64
	Hash   string
	DFN    uint64
	Low    uint64
	Pushed bool
	Group  map[uint64]bool
}

type TxLog struct {
	Hash string
	Logs map[uint64]*protos.OrderedLog
}

const (
	DefaultGraphSize = 50
)
