package types

import (
	"github.com/Grivn/libfalanx/logger"
	pb "github.com/Grivn/libfalanx/zcommon/protos"
)

type Config struct {
	ID     uint64
	RecvC  chan *pb.OrderedReq
	OrderC chan string
	Logger logger.Logger
}
