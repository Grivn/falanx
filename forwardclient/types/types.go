package types

import (
	"github.com/Grivn/libfalanx/logger"
	"github.com/Grivn/libfalanx/network"
	"github.com/Grivn/libfalanx/zcommon"
	pb "github.com/Grivn/libfalanx/zcommon/protos"
)

type Config struct {
	ID     uint64
	Hash   string
	SelfC  chan *pb.OrderedReq
	Tools  zcommon.Tools
	Sender network.Network
	Logger logger.Logger
}
