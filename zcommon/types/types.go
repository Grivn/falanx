package types

import (
	"github.com/Grivn/libfalanx/logger"
	"github.com/Grivn/libfalanx/network"
	"github.com/Grivn/libfalanx/zcommon"
)

type Config struct {
	ID     uint64
	N      int
	Sender network.Network
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

const (
	DefaultChannelLen = 1000
	)
