package txcontainer

import (
	"github.com/Grivn/libfalanx/txcontainer/types"
	fCommonProto "github.com/ultramesh/flato-common/types/protos"
)

func NewTxContainer(config types.Config) *containerImpl {
	return &containerImpl{
		pendingTxs: make(map[string]*fCommonProto.Transaction),
		tools:      config.Tools,
		logger:     config.Logger,
	}
}
func (c *containerImpl) Add(tx *fCommonProto.Transaction) {
	c.add(tx)
}
func (c *containerImpl) Get(txHash string) *fCommonProto.Transaction {
	return c.get(txHash)
}
func (c *containerImpl) Remove(txHash string) error {
	return c.remove(txHash)
}
