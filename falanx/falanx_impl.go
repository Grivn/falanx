package falanx

import (
	"github.com/Grivn/libfalanx/api"
	"github.com/Grivn/libfalanx/clientsorder"
	clientOrderType "github.com/Grivn/libfalanx/clientsorder/types"
	"github.com/Grivn/libfalanx/filter"
	filterType "github.com/Grivn/libfalanx/filter/types"
	"github.com/Grivn/libfalanx/forwardclient"
	fakeClientType "github.com/Grivn/libfalanx/forwardclient/types"
	"github.com/Grivn/libfalanx/graphengine"
	"github.com/Grivn/libfalanx/localorder"
	localOrderType "github.com/Grivn/libfalanx/localorder/types"
	"github.com/Grivn/libfalanx/logger"
	"github.com/Grivn/libfalanx/replicasorder"
	replicaOrderType "github.com/Grivn/libfalanx/replicasorder/types"
	"github.com/Grivn/libfalanx/txcontainer"
	containerType "github.com/Grivn/libfalanx/txcontainer/types"
	pb "github.com/Grivn/libfalanx/zcommon/protos"
	"github.com/Grivn/libfalanx/zcommon/types"
	"github.com/gogo/protobuf/proto"
)

type falanxImpl struct {
	// properties ====================================================================================
	// id: identifier of current replica
	id uint64

	// modules =======================================================================================
	// forwardClient: used to forward the txs send from the clients which trust current replica
	// txContainer:   used to contain the transactions
	// localOrder:    used to generate current node's ordered logs
	// clientsOrder:  used to process the ordered requests from clients
	// replicasOrder: used to process the ordered logs from replicas
	// txFilter:      used to generate graph
	// graphEngine:   used to deal with the raw graph
	forwardClient api.ForwardClient
	txContainer   api.TxsContainer
	localOrder    api.ModuleControl
	clientsOrder  map[uint64]api.ModuleControl
	replicasOrder map[uint64]api.ModuleControl
	txFilter      api.ModuleControl
	graphEngine   api.ModuleControl

	// channel =======================================================================================
	// the channels which will be used to deliver messages between different modules
	//
	// internal channel
	// reqRecvC:  dispatch the ordered requests to specific client order module
	// reqOrderC: collect the ordered txHash from different client and deliver them to local order module
	// reqRecvC:  dispatch the ordered logs to specific replica order module
	// reqOrderC: collect the ordered logs from different replica and deliver them to filter module
	//
	// ordered_req ---> reqRecvC ---> clientsOrder
	// the ordered reqs will be picked from clientsOrder one by one
	//
	// txHash --------> reqOrderC --> localOrder
	// localOrder will order the txHash from client and broadcast ordered logs
	//
	// ordered_log ---> logRecvC ---> replicasOrder
	// the ordered logs will be picked from replicasOrder one by one
	//
	// ordered_log ---> logOrderC --> txFilter
	// txFilter will collect the logs from different replicasOrder and generate a graph
	reqRecvC    map[uint64]chan *pb.OrderedReq
	reqOrderC   chan string
	logRecvC    map[uint64]chan *pb.OrderedLog
	logOrderC   chan *pb.OrderedLog
	close       chan bool

	// essential =====================================================================================
	logger logger.Logger
}

func newFalanxImpl(c types.Config) *falanxImpl {
	reqRecvC := make(map[uint64]chan *pb.OrderedReq)
	logRecvC := make(map[uint64]chan *pb.OrderedLog)
	reqOrderC := make(chan string)
	logOrderC := make(chan *pb.OrderedLog)
	graphC := make(chan interface{})

	// initialize the tx container
	containerConfig := containerType.Config{
		Logger: c.Logger,
		Tools:  c.Tools,
	}
	txContainer := txcontainer.NewTxContainer(containerConfig)

	// initialize the client order
	clientsOrder := make(map[uint64]api.ModuleControl)
	for i:=0; i<c.N; i++ {
		id := uint64(i+1)
		recvC := make(chan *pb.OrderedReq, types.DefaultChannelLen)
		clientConfig := clientOrderType.Config{
			ID:     id,
			RecvC:  recvC,
			OrderC: reqOrderC,
			Logger: c.Logger,
		}
		reqRecvC[id] = recvC
		clientsOrder[id] = clientsorder.NewClientOrder(clientConfig)
	}

	// initialize the replica order
	var replicas []int
	replicasOrder := make(map[uint64]api.ModuleControl)
	for i:=0; i<c.N; i++ {
		id := uint64(i+1)
		recvC := make(chan *pb.OrderedLog, types.DefaultChannelLen)
		replicaConfig := replicaOrderType.Config{
			ID:     id,
			RecvC:  recvC,
			OrderC: logOrderC,
			Logger: c.Logger,
		}
		logRecvC[id] = recvC
		replicasOrder[id] = replicasorder.NewReplicaOrder(replicaConfig)
		replicas = append(replicas, int(id))
	}

	// client
	clientConfig := fakeClientType.Config{
		ID:     c.ID,
		SelfC:  reqRecvC[c.ID],
		Tools:  c.Tools,
		Sender: c.Sender,
		Logger: c.Logger,
	}
	fakeClient := forwardclient.NewClient(clientConfig)

	// local order
	localConfig := localOrderType.Config{
		ID:      c.ID,
		RecvC:   reqOrderC,
		SelfC:   logRecvC[c.ID],
		Network: c.Sender,
		Logger:  c.Logger,
	}
	localOrder := localorder.NewLocalOrder(localConfig)

	// filter
	filterConfig := filterType.Config{
		Replicas: replicas,
		Order:    logOrderC,
		Graph:    graphC,
		Logger:   c.Logger,
		Tools:    c.Tools,
	}
	txFilter := filter.NewTransactionFilter(filterConfig)

	graphEngine := graphengine.NewGraphEngine(graphC)

	falanx := &falanxImpl{
		id:            c.ID,
		forwardClient: fakeClient,
		txContainer:   txContainer,
		localOrder:    localOrder,
		clientsOrder:  clientsOrder,
		replicasOrder: replicasOrder,
		txFilter:      txFilter,
		graphEngine:   graphEngine,
		reqRecvC:      reqRecvC,
		reqOrderC:     reqOrderC,
		logRecvC:      logRecvC,
		logOrderC:     logOrderC,
		close:         make(chan bool),
		logger:        c.Logger,
	}

	return falanx
}

func (falanx *falanxImpl) start() {

	falanx.localOrder.Start()

	for _, replica := range falanx.replicasOrder {
		replica.Start()
	}

	for _, client := range falanx.clientsOrder {
		client.Start()
	}

	falanx.txFilter.Start()

	falanx.graphEngine.Start()

	falanx.logger.Info(`

+=============================================================================+
｜                                                                           ｜
｜        _/_/_/_/    _/_/    _/          _/_/    _/      _/  _/      _/     ｜
｜       _/        _/    _/  _/        _/    _/  _/_/    _/    _/  _/        ｜
｜      _/_/_/    _/_/_/_/  _/        _/_/_/_/  _/  _/  _/      _/           ｜
｜     _/        _/    _/  _/        _/    _/  _/    _/_/    _/  _/          ｜
｜    _/        _/    _/  _/_/_/_/  _/    _/  _/      _/  _/      _/         ｜
｜                                                                           ｜
+=============================================================================+

`)
}

func (falanx *falanxImpl) stop() {
	close(falanx.close)
}

func (falanx *falanxImpl) step(msg *pb.ConsensusMessage) {
	switch msg.Type {
	case pb.Type_REQUEST_SET:
		request := &pb.RequestSet{}
		err := proto.Unmarshal(msg.Payload, request)
		if err != nil {
			return
		}
		falanx.forwardClient.ProposeTxs(request.Requests)
	case pb.Type_ORDERED_REQ:
		falanx.logger.Info("[REQ] Receive an ordered request")
		req := &pb.OrderedReq{}
		err := proto.Unmarshal(msg.Payload, req)
		if err != nil {
			return
		}
		falanx.processOrderedReq(req)
	case pb.Type_ORDERED_LOG:
		falanx.logger.Info("[LOG] Receive an ordered log")
		log := &pb.OrderedLog{}
		err := proto.Unmarshal(msg.Payload, log)
		if err != nil {
			return
		}
		falanx.processOrderedLog(log)
	}
}

func (falanx *falanxImpl) processOrderedReq(req *pb.OrderedReq) {
	falanx.logger.Debugf("Replica %d receive an ordered request from client %d", falanx.id, req.ClientId)
	recvC, ok := falanx.reqRecvC[req.ClientId]
	if ok {
		recvC <- req
		return
	}
}

func (falanx *falanxImpl) processOrderedLog(log *pb.OrderedLog) {
	falanx.logger.Debugf("Replica %d receive an ordered log from replica %d, seq %d", falanx.id, log.ReplicaId, log.Sequence)
	recvC, ok := falanx.logRecvC[log.ReplicaId]
	if ok {
		recvC <- log
		return
	}

	falanx.logger.Errorf("invalid replica %d", log.ReplicaId)
}
