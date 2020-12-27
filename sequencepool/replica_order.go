package sequencepool

import (
	"github.com/Grivn/libfalanx/common"
	"github.com/Grivn/libfalanx/logger"
	"github.com/Grivn/libfalanx/sequencepool/utils"
)

type ReplicaOrder interface {

}

type replicaOrderImpl struct {
	id uint64

	cached  *utils.LogHeap
	ordered *utils.ReplicaRecorder

	// essential tools for ordered pool
	tools  common.Tools
	logger logger.Logger
}

func NewReplicaOrder(id uint64, tools common.Tools, logger logger.Logger) *replicaOrderImpl {
	logger.Noticef("Initialize replica order instance: [id]%d", id)
	return &replicaOrderImpl{
		id: id,
		tools: tools,
		logger: logger,
	}
}
