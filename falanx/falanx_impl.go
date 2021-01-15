package falanx

import (
	"github.com/Grivn/libfalanx/clientsorder"
	clientOrderType "github.com/Grivn/libfalanx/clientsorder/types"
	"github.com/Grivn/libfalanx/fakeclient"
	fakeClientType "github.com/Grivn/libfalanx/fakeclient/types"
	"github.com/Grivn/libfalanx/falanx/external"
	"github.com/Grivn/libfalanx/filter"
	filterType "github.com/Grivn/libfalanx/filter/types"
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
	"github.com/golang/protobuf/proto"
	fCommonProto "github.com/ultramesh/flato-common/types/protos"
)

type falanxImpl struct {
	// properties ====================================================================================
	// id: identifier of current replica
	id uint64

	// modules =======================================================================================
	// txContainer:   used to contain the transactions
	// clientsOrder:  used to process the ordered requests from clients
	// replicasOrder: used to process the ordered logs from replicas
	fakeClient    external.FakeClient
	txContainer   external.TxsContainer
	localOrder    external.ModuleControl
	clientsOrder  map[uint64]external.ModuleControl
	replicasOrder map[uint64]external.ModuleControl
	txFilter      external.ModuleControl
	graphEngine   external.ModuleControl

	// channel =======================================================================================
	// the channels which will be used to deliver messages between different modules
	//
	// external channel
	// txC:  receive the transactions from network
	// reqC: receive the ordered requests from network
	// logC: receive the ordered logs from network
	//
	// internal channel
	// reqRecvC:  dispatch the ordered requests to specific client order module
	// reqOrderC: collect the ordered txHash from different client and deliver them to local order module
	// reqRecvC:  dispatch the ordered logs to specific replica order module
	// reqOrderC: collect the ordered logs from different replica and deliver them to filter module
	//
	// network ------------------> reqC ------------------> ordered_req
	// ordered_req --------------> reqRecvC -------------> clients_order
	// clients_order will generate the txHash by order
	//
	// txHash (by order) --------> reqOrderC ------------> local_order
	// local_order will order the txHash from client and broadcast ordered_logs
	//
	// network ------------------> logC ------------------> ordered_log
	// ordered_log --------------> logRecvC ------------> replicas_order
	// replicas_order will generate the logs by order
	//
	// ordered_log (by order) ---> logOrderC -----------> filter
	// in filter, we will collect the logs from different replicasOrder and generate the relation graph)
	txC         chan *fCommonProto.Transaction
	reqC        chan *pb.OrderedReq
	logC        chan *pb.OrderedLog
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

	// client
	clientConfig := fakeClientType.Config{
		ID:     c.ID,
		Tools:  c.Tools,
		Sender: c.Sender,
		Logger: c.Logger,
	}
	fakeClient := fakeclient.NewClient(clientConfig)

	// initialize the tx container
	containerConfig := containerType.Config{
		Logger: c.Logger,
		Tools:  c.Tools,
	}
	txContainer := txcontainer.NewTxContainer(containerConfig)

	// local order
	localConfig := localOrderType.Config{
		ID:      c.ID,
		RecvC:   reqOrderC,
		Network: c.Sender,
		Logger:  c.Logger,
	}
	localOrder := localorder.NewLocalOrder(localConfig)

	// initialize the replica order
	var replicas []int
	replicasOrder := make(map[uint64]external.ModuleControl)
	for i:=0; i<c.N; i++ {
		id := uint64(i+1)
		recvC := make(chan *pb.OrderedLog)
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
		fakeClient:    fakeClient,
		txContainer:   txContainer,
		localOrder:    localOrder,
		clientsOrder:  make(map[uint64]external.ModuleControl),
		replicasOrder: replicasOrder,
		txFilter:      txFilter,
		graphEngine:   graphEngine,
		txC:           c.TxC,
		reqC:          c.ReqC,
		reqRecvC:      reqRecvC,
		reqOrderC:     reqOrderC,
		logC:          c.LogC,
		logRecvC:      logRecvC,
		logOrderC:     logOrderC,
		close:         make(chan bool),
		logger:        c.Logger,
	}

	return falanx
}

func (falanx *falanxImpl) start() {
	go falanx.listenTransactions()
	go falanx.listenOrderedReqs()
	go falanx.listenOrderedLogs()

	falanx.localOrder.Start()

	for _, replica := range falanx.replicasOrder {
		replica.Start()
	}

	falanx.txFilter.Start()

	falanx.graphEngine.Start()

	falanx.logger.Notice(`

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
		falanx.fakeClient.ProposeTxs(request.Requests)
	case pb.Type_ORDERED_REQ:
		req := &pb.OrderedReq{}
		err := proto.Unmarshal(msg.Payload, req)
		if err != nil {
			return
		}
		falanx.reqC <- req
	case pb.Type_ORDERED_LOG:
		log := &pb.OrderedLog{}
		err := proto.Unmarshal(msg.Payload, log)
		if err != nil {
			return
		}
		falanx.logC <- log
	}
}

func (falanx *falanxImpl) listenTransactions() {
	for {
		select {
		case <-falanx.close:
			return

		case tx := <-falanx.txC:
			falanx.txContainer.Add(tx)
		}
	}
}

func (falanx *falanxImpl) listenOrderedReqs() {
	for {
		select {
		case <-falanx.close:
			return

		case req := <-falanx.reqC:
			falanx.processOrderedReqs(req)
		}
	}
}

func (falanx *falanxImpl) processOrderedReqs(req *pb.OrderedReq) {
	falanx.logger.Debugf("Replica %d receive an ordered request from client %d", falanx.id, req.ClientId)
	recvC, ok := falanx.reqRecvC[req.ClientId]
	if ok {
		falanx.logger.Debugf("Already initialize channel for client %d", req.ClientId)
		go func() {
			recvC <- req
		}()
		return
	}

	// initialize a new client order for particular client
	initC := make(chan *pb.OrderedReq)
	falanx.reqRecvC[req.ClientId] = initC
	config := clientOrderType.Config{
		ID:     req.ClientId,
		RecvC:  initC,
		OrderC: falanx.reqOrderC,
		Logger: falanx.logger,
	}
	client := clientsorder.NewClientOrder(config)
	client.Start()
	initC <- req
}

func (falanx *falanxImpl) listenOrderedLogs() {
	for {
		select {
		case <-falanx.close:
			return

		case log := <-falanx.logC:
			falanx.processOrderedLogs(log)
		}
	}
}

func (falanx *falanxImpl) processOrderedLogs(log *pb.OrderedLog) {
	falanx.logger.Debugf("Replica %d receive an ordered log from replica %d", falanx.id, log.ReplicaId)
	recvC, ok := falanx.logRecvC[log.ReplicaId]
	if ok {
		go func() {
			recvC <- log
		}()
		return
	}

	falanx.logger.Errorf("invalid replica %d", log.ReplicaId)
}
