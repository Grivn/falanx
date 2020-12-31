package fakeclient

import (
	"time"

	"github.com/Grivn/libfalanx/zcommon"
	"github.com/Grivn/libfalanx/zcommon/protos"
	"github.com/Grivn/libfalanx/fakeclient/types"
	"github.com/Grivn/libfalanx/logger"
	"github.com/Grivn/libfalanx/network"

	fCommonProto "github.com/ultramesh/flato-common/types/protos"
)

type Client interface {
	Propose(txs []*fCommonProto.Transaction)
	ReceiveReply(reply *protos.Reply)
}

type clientImpl struct {
	id uint64
	n  uint64
	f  uint64

	txs map[string]*fCommonProto.Transaction
	seq uint64

	res map[string]map[uint64]bool

	tools  zcommon.Tools
	sender network.Network
	logger logger.Logger
}

func NewClient(config types.Config) *clientImpl {
	return &clientImpl{
		id:     config.ID,
		txs:    make(map[string]*fCommonProto.Transaction),
		seq:    uint64(0),
		tools:  config.Tools,
		sender: config.Sender,
		logger: config.Logger,
	}
}

func (c *clientImpl) Propose(txs []*fCommonProto.Transaction) {
	hashList := make([]string, len(txs))
	for index, tx := range txs {
		hash := c.tools.TransactionHash(tx)
		hashList[index] = hash
	}

	c.seq++
	req := &protos.OrderedReq{
		ClientId:   c.id,
		Sequence:   c.seq,
		TxHashList: hashList,
		Timestamp:  time.Now().UnixNano(),
	}
	c.logger.Infof("Client %d broadcast ordered request: [seq]%d, [hash list]%v", req.Sequence, req.TxHashList)
	c.sender.BroadcastOrderedReq(req)
	c.sender.BroadcastTransactions(txs)
}

func (c *clientImpl) ReceiveReply(reply *protos.Reply) {

}