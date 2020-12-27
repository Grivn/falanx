package types

import (
	"github.com/Grivn/libfalanx/common"
	"github.com/Grivn/libfalanx/logger"
)

// Config is used to initiate the instance in sequence pool
type Config struct {
	Author uint64

	TimeoutC chan interface{}
	LocalBAC chan interface{}

	Tools  common.Tools
	Logger logger.Logger
}

// TxRecorder is used to describe the major information to make order for certain transaction
type TxRecorder struct {
	TxHash          string
	OrderedReplicas []uint64
	Candidates      []uint64
}

// TimeoutEvent is used to process local ba progress
type TimeoutEvent struct {
	// TxHash is used to track the particular transaction waiting for quorum candidates
	TxHash string
}
