package txcontainer

import (
	"github.com/Grivn/libfalanx/txcontainer/types"
	pb "github.com/Grivn/libfalanx/zcommon/protos"
)

func NewTxContainer(config types.Config) *containerImpl {
	return &containerImpl{
		pendingTxs: make(map[string]*pb.Transaction),
		tools:      config.Tools,
		logger:     config.Logger,
	}
}
func (c *containerImpl) Add(tx *pb.Transaction) {
	c.add(tx)
}
func (c *containerImpl) Get(txHash string) *pb.Transaction {
	return c.get(txHash)
}
func (c *containerImpl) Remove(txHash string) error {
	return c.remove(txHash)
}
