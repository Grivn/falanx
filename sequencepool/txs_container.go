package sequencepool

import (
	"github.com/Grivn/libfalanx/common"
	"github.com/Grivn/libfalanx/logger"
	"github.com/Grivn/libfalanx/sequencepool/types"

	fCommonProto "github.com/ultramesh/flato-common/types/protos"
)

// TxsContainer is only used to receive the transactions send from clients or replicas
// we don't need to maintain the order here and it is only a container for transactions
type TxsContainer interface {
	AddTransaction(tx *fCommonProto.Transaction)
	GetTransaction(txHash string) *fCommonProto.Transaction
}

type containerImpl struct {
	// author is the identifier
	author uint64

	// pendingTxs means the transactions which have not been executed
	pendingTxs map[string]*fCommonProto.Transaction

	// essential external tools for txPool
	tools  common.Tools
	logger logger.Logger
}

func NewTxContainer(config types.Config) *containerImpl {
	return &containerImpl{
		author:     config.Author,
		pendingTxs: make(map[string]*fCommonProto.Transaction),
		tools:      config.Tools,
		logger:     config.Logger,
	}
}

func (c *containerImpl) AddTransaction(tx *fCommonProto.Transaction) {
	if tx == nil {
		c.logger.Warningf("Replica %d received a nil transaction", c.author)
		return
	}
	txHash := c.tools.TransactionHash(tx)
	c.pendingTxs[txHash] = tx
}

func (c *containerImpl) GetTransaction(txHash string) *fCommonProto.Transaction {
	tx, ok := c.pendingTxs[txHash]
	if !ok {
		c.logger.Debugf("Replica %d cannot find such a transaction %d", txHash)
		return nil
	}
	return tx
}
