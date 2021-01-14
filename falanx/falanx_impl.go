package falanx

import (
	"github.com/Grivn/libfalanx/clientsorder"
	clientOrderType "github.com/Grivn/libfalanx/clientsorder/types"
	"github.com/Grivn/libfalanx/falanx/external"
	"github.com/Grivn/libfalanx/logger"
	"github.com/Grivn/libfalanx/replicasorder"
	replicaOrderType "github.com/Grivn/libfalanx/replicasorder/types"
	"github.com/Grivn/libfalanx/txcontainer"
	containerType "github.com/Grivn/libfalanx/txcontainer/types"
	pb "github.com/Grivn/libfalanx/zcommon/protos"
	"github.com/Grivn/libfalanx/zcommon/types"
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
	txContainer   external.TxsContainer
	clientsOrder  map[uint64]external.ModuleControl
	replicasOrder map[uint64]external.ModuleControl

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

	// initialize the tx container
	containerConfig := containerType.Config{
		Logger: c.Logger,
		Tools:  c.Tools,
	}
	txContainer := txcontainer.NewTxContainer(containerConfig)

	// initialize the replica order
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
	}

	falanx := &falanxImpl{
		id:            c.ID,
		txContainer:   txContainer,
		clientsOrder:  make(map[uint64]external.ModuleControl),
		replicasOrder: replicasOrder,
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
}

func (falanx *falanxImpl) stop() {
	close(falanx.close)
}

func (falanx *falanxImpl) step(payload []byte) {

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
	recvC, ok := falanx.reqRecvC[req.ClientId]
	if ok {
		recvC <- req
		return
	}

	// initialize a new client order for particular client
	initC := make(chan *pb.OrderedReq)
	config := clientOrderType.Config{
		ID:     req.ClientId,
		RecvC:  initC,
		OrderC: falanx.reqOrderC,
		Logger: falanx.logger,
	}
	client := clientsorder.NewClientOrder(config)
	client.Start()
	falanx.reqRecvC[req.ClientId] = initC
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
	recvC, ok := falanx.logRecvC[log.ReplicaId]
	if ok {
		recvC <- log
		return
	}

	falanx.logger.Errorf("invalid replica %d", log.ReplicaId)
}
