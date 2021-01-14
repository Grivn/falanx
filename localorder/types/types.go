package types

import (
	"github.com/Grivn/libfalanx/logger"
	"github.com/Grivn/libfalanx/network"
)

type Config struct {
	ID      uint64
	RecvC   chan string
	Network network.Network
	Logger  logger.Logger
}
