package forwardclient

import (
	"time"

	"github.com/Grivn/libfalanx/forwardclient/types"
	"github.com/Grivn/libfalanx/logger"
	"github.com/Grivn/libfalanx/network"
	"github.com/Grivn/libfalanx/zcommon"
	pb "github.com/Grivn/libfalanx/zcommon/protos"

	"github.com/golang/protobuf/proto"
)

type clientImpl struct {
	id uint64
	n  uint64
	f  uint64

	txs map[string]*pb.Transaction
	seq uint64

	res map[string]map[uint64]bool

	selfC chan *pb.OrderedReq

	tools  zcommon.Tools
	sender network.Network
	logger logger.Logger
}

func newClientImpl(config types.Config) *clientImpl {
	return &clientImpl{
		id:     config.ID,
		txs:    make(map[string]*pb.Transaction),
		seq:    uint64(0),
		selfC:  config.SelfC,
		tools:  config.Tools,
		sender: config.Sender,
		logger: config.Logger,
	}
}

func (c *clientImpl) propose(txs []*pb.Transaction) {
	hashList := make([]string, len(txs))
	for index, tx := range txs {
		hash := c.tools.TransactionHash(tx)
		hashList[index] = hash
	}

	c.seq++
	req := &pb.OrderedReq{
		ClientId:   c.id,
		Sequence:   c.seq,
		TxHashList: hashList,
		Timestamp:  time.Now().UnixNano(),
	}
	c.logger.Debugf("Client %d broadcast ordered request: [seq]%d", c.id, req.Sequence)

	reqPayload, err := proto.Marshal(req)
	if err != nil {
		return
	}
	reqMsg := &pb.ConsensusMessage{
		Type:    pb.Type_ORDERED_REQ,
		Payload: reqPayload,
	}
	c.sender.Broadcast(reqMsg)
	c.inform(req)
}

func (c *clientImpl) inform(req *pb.OrderedReq) {
	c.selfC <- req
}
