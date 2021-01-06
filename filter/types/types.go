package types

import (
	"github.com/Grivn/libfalanx/logger"
	"github.com/Grivn/libfalanx/zcommon"
	"github.com/Grivn/libfalanx/zcommon/protos"
)

type Config struct {
	Replicas []int

	Order  chan *protos.OrderedLog
	Passed chan []string

	Logger logger.Logger
	Tools  zcommon.Tools
}
