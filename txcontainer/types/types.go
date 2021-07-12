package types

import (
	"github.com/Grivn/libfalanx/logger"
	"github.com/Grivn/libfalanx/zcommon"
)

// Config is used to initiate the instance in tx container
type Config struct {
	Tools  zcommon.Tools
	Logger logger.Logger
}
