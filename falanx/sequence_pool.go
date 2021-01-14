package falanx

import (
	"github.com/Grivn/libfalanx/txcontainer"
	"github.com/Grivn/libfalanx/clientsorder"
	"github.com/Grivn/libfalanx/replicasorder"
	"github.com/Grivn/libfalanx/logger"
	"github.com/Grivn/libfalanx/zcommon"
	"github.com/Grivn/libfalanx/zcommon/protos"

	fCommonProto "github.com/ultramesh/flato-common/types/protos"
)

type orderedPool struct {
	// transactions ================================================================
	// container: store the transactions send from every client
	//
	// transactions set will be send into container directly
	container txcontainer.TxsContainer

	// order =======================================================================
	// clients:  receive the ordered requests from every client
	// replicas: receive the ordered logs from every replica
	//
	// a ordered message only has a unique sequence number and hash value
	// the transactions ordered in clients will be transferred into the replicas as local log
	// the transactions matched target in replicas will be carried out to produce map
	clients  map[uint64]clientsorder.ClientOrder
	replicas map[uint64]replicasorder.ReplicaOrder

	// channel =====================================================================
	// transactionChan: transfer the transactions from client
	// clientsChan:     transfer the ordered requests from client
	//                  there is only one channel to order transactions from different client
	// orderedChan:     transfer between clients and replicas
	//                  the channel should be individual in order to process the ordered event
	// replicasChan:    transfer the ordered logs from replica
	//                  the channel could be individual to process
	// close:           stop the go-routine of current module
	transactionChan chan []*fCommonProto.Transaction
	clientsChan     chan *protos.OrderedReq
	orderedChan     map[uint64]chan string
	replicasChan    map[uint64]chan *protos.OrderedLog
	close           chan bool

	// general =====================================================================
	tools  zcommon.Tools
	logger logger.Logger
}

func newOrderedPool() *orderedPool {
	return nil
}

func (o *orderedPool) start() {
	go o.listenTransactions()
	go o.listenOrderedRequests()
}

// transactions' listener ===========================================================
func (o *orderedPool) listenTransactions() {
	for {
		select {
		case <-o.close:
			return
		case txs := <-o.transactionChan:
			o.processTransactions(txs)
		}
	}
}
func (o *orderedPool) processTransactions(txs []*fCommonProto.Transaction) {
	for _, tx := range txs {
		o.container.Add(tx)
	}
}
// transactions' listener ===========================================================

// clients' listener =================================================
func (o *orderedPool) listenOrderedRequests() {
	for {
		select {
		case <-o.close:
			return
		case req := <-o.clientsChan:
			o.processOrderedRequests(req)
		}
	}
}

func (o *orderedPool) listenOrderedChan(cid uint64) {
	for {
		select {
		case <-o.close:
			return
		case txHash := <-o.orderedChan[cid]:
			o.processOrderedHash(txHash)
		}
	}
}

func (o *orderedPool) processOrderedRequests(req *protos.OrderedReq) {
	cid := req.ClientId
	client, ok := o.clients[cid]
	if !ok {
		// initialization of client order for client with cid
		orderedChan := make(chan string)
		client = clientsorder.NewClientOrder(cid, orderedChan, o.tools, o.logger)
		o.orderedChan[cid] = orderedChan
		o.clients[cid] = client
		go o.listenOrderedChan(cid)
	}
	client.ReceiveOrderedReq(req)
}

func (o *orderedPool) processOrderedHash(txHash string) {
	// todo trigger the component of channel related to replica order
}
// clientsOrder' listener =================================================

func (o *orderedPool) listenOrderedLogs() {

}

func (o *orderedPool) processOrderedLog(l *protos.OrderedLog) {

}
