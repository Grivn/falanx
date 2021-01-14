package types

import (
	"github.com/Grivn/libfalanx/logger"
	"github.com/Grivn/libfalanx/zcommon/protos"
)

type Config struct {
	ID     uint64
	RecvC  chan *protos.OrderedReq
	OrderC chan string
	Logger logger.Logger
}
