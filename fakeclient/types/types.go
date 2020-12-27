package types

import (
	"github.com/Grivn/libfalanx/common"
	"github.com/Grivn/libfalanx/logger"
	"github.com/Grivn/libfalanx/network"
)

type Config struct {
	ID uint64
	Hash string
	Tools common.Tools
	Sender network.Network
	Logger logger.Logger
}
