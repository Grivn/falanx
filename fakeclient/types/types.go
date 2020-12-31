package types

import (
	"github.com/Grivn/libfalanx/zcommon"
	"github.com/Grivn/libfalanx/logger"
	"github.com/Grivn/libfalanx/network"
)

type Config struct {
	ID     uint64
	Hash   string
	Tools  zcommon.Tools
	Sender network.Network
	Logger logger.Logger
}
