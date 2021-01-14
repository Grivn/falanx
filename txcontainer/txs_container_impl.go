package txcontainer

import (
	"errors"

	"github.com/Grivn/libfalanx/logger"
	"github.com/Grivn/libfalanx/zcommon"

	fCommonProto "github.com/ultramesh/flato-common/types/protos"
)

// Implementation ==================================================
type containerImpl struct {
	// pendingTxs means the transactions which have not been executed
	pendingTxs map[string]*fCommonProto.Transaction

	// essential external tools for txPool
	tools  zcommon.Tools
	logger logger.Logger
}

func (c *containerImpl) add(tx *fCommonProto.Transaction) {
	if tx == nil {
		c.logger.Warning("container received a nil transaction")
		return
	}
	txHash := c.tools.TransactionHash(tx)
	c.pendingTxs[txHash] = tx
}

func (c *containerImpl) has(txHash string) bool {
	_, ok := c.pendingTxs[txHash]
	return ok
}

func (c *containerImpl) get(txHash string) *fCommonProto.Transaction {
	if !c.has(txHash) {
		c.logger.Debugf("Replica %d cannot find such a transaction %d", txHash)
		return nil
	}
	return c.pendingTxs[txHash]
}

func (c *containerImpl) remove(txHash string) error {
	if !c.has(txHash) {
		c.logger.Debugf("Replica %d cannot find such a transaction %d", txHash)
		return errors.New("non-existed transaction")
	}
	delete(c.pendingTxs, txHash)
	return nil
}
