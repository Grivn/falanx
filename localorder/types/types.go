package types

import (
	"github.com/Grivn/libfalanx/logger"
	"github.com/Grivn/libfalanx/network"
	pb "github.com/Grivn/libfalanx/zcommon/protos"
)

type Config struct {
	ID      uint64
	RecvC   chan string
	SelfC   chan *pb.OrderedLog
	Network network.Network
	Logger  logger.Logger
}
