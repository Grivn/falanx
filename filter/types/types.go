package types

import (
	"github.com/Grivn/libfalanx/logger"
	"github.com/Grivn/libfalanx/zcommon"
	pb "github.com/Grivn/libfalanx/zcommon/protos"
)

type Config struct {
	Replicas []int

	Order  chan *pb.OrderedLog
	Graph  chan interface{}

	Logger logger.Logger
	Tools  zcommon.Tools
}

type BeforeCheck uint64

const (
	NotEfficient   = 0x0
	FormerPriority = 0x01
	LatterPriority = 0x02
)

type RelationId struct {
	From string
	To   string
}

type RelationCert struct {
	Finished        bool
	Status          BeforeCheck
	Scanned         map[uint64]bool
	FormerPreferred int
	LatterPreferred int
}

const (
	DefaultGraphSize = 5
)

type PavedTxs struct {
	Seq uint64
	Txs map[string]bool
}
