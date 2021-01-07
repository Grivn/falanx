package types

import (
	graphTypes "github.com/Grivn/libfalanx/graphengine/types"
	"github.com/Grivn/libfalanx/logger"
	"github.com/Grivn/libfalanx/zcommon"
	"github.com/Grivn/libfalanx/zcommon/protos"
)

type Config struct {
	Replicas []int

	Order  chan *protos.OrderedLog
	Graph  chan graphTypes.GraphEvent

	Logger logger.Logger
	Tools  zcommon.Tools
}
