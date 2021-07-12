package txcontainer

import (
	"errors"
	pb "github.com/Grivn/libfalanx/zcommon/protos"

	"github.com/Grivn/libfalanx/logger"
	"github.com/Grivn/libfalanx/zcommon"
)

// Implementation ==================================================
type containerImpl struct {
	// pendingTxs means the transactions which have not been executed
	pendingTxs map[string]*pb.Transaction

	// essential external tools for txPool
	tools  zcommon.Tools
	logger logger.Logger
}

func (c *containerImpl) add(tx *pb.Transaction) {
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

func (c *containerImpl) get(txHash string) *pb.Transaction {
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
