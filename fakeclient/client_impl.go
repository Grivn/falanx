package fakeclient

import (
	"time"

	"github.com/Grivn/libfalanx/fakeclient/types"
	"github.com/Grivn/libfalanx/logger"
	"github.com/Grivn/libfalanx/network"
	"github.com/Grivn/libfalanx/zcommon"
	pb "github.com/Grivn/libfalanx/zcommon/protos"

	"github.com/golang/protobuf/proto"
	fCommonProto "github.com/ultramesh/flato-common/types/protos"
)

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

func newClientImpl(config types.Config) *clientImpl {
	return &clientImpl{
		id:     config.ID,
		txs:    make(map[string]*fCommonProto.Transaction),
		seq:    uint64(0),
		tools:  config.Tools,
		sender: config.Sender,
		logger: config.Logger,
	}
}

func (c *clientImpl) propose(txs []*fCommonProto.Transaction) {
	hashList := make([]string, len(txs))
	for index, tx := range txs {
		hash := c.tools.TransactionHash(tx)
		hashList[index] = hash
	}

	c.seq++
	//set := &pb.RequestSet{
	//	Requests: txs,
	//}
	req := &pb.OrderedReq{
		ClientId:   c.id,
		Sequence:   c.seq,
		TxHashList: hashList,
		Timestamp:  time.Now().UnixNano(),
	}
	c.logger.Debugf("Client %d broadcast ordered request: [seq]%d", c.id, req.Sequence)

	//txsPayload, err := proto.Marshal(set)
	//if err != nil {
	//	return
	//}
	//txsMsg := &pb.ConsensusMessage{
	//	Type:    pb.Type_REQUEST_SET,
	//	Payload: txsPayload,
	//}
	//c.sender.Broadcast(txsMsg)

	reqPayload, err := proto.Marshal(req)
	if err != nil {
		return
	}
	reqMsg := &pb.ConsensusMessage{
		Type:    pb.Type_ORDERED_REQ,
		Payload: reqPayload,
	}
	c.sender.Broadcast(reqMsg)
}
