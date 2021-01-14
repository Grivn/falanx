package types

import (
	"github.com/Grivn/libfalanx/logger"
	"github.com/Grivn/libfalanx/zcommon"
	"github.com/Grivn/libfalanx/zcommon/protos"
	fCommonProto "github.com/ultramesh/flato-common/types/protos"
)

type Config struct {
	ID     uint64
	N      int
	TxC    chan *fCommonProto.Transaction
	ReqC   chan *protos.OrderedReq
	LogC   chan *protos.OrderedLog
	Tools  zcommon.Tools
	Logger logger.Logger
}

type Peer struct {
	ID   uint64
	Hash string
}

type LocalBAEvent struct {
	TxHash          string
	MissingReplicas []uint64
}
