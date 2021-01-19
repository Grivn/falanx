package filter

import (
	"github.com/Grivn/libfalanx/filter/types"
	tp "github.com/Grivn/libfalanx/zcommon/types"
	"math"
	"time"

	"github.com/Grivn/libfalanx/filter/utils"
	"github.com/Grivn/libfalanx/logger"
	pb "github.com/Grivn/libfalanx/zcommon/protos"
)

type transactionsFilterImpl struct {
	// basic ======================================================================
	// n:     indicates the amount of replicas
	// multi: is related to the amount of transactions we will deal with per time
	//        any time we will process n*multi transactions
	//
	// e.g. n=4, multi=2
	//      t1 --> t2 --> t3 --> t4
	//      t2 --> t1 --> t3
	//      t2 --> t1 --> t3 --> t4
	//      t3 --> t1 --> t4
	//
	// we would like to replicaOrder the set {{t1,t2,t2,t3},{t2,t1,t1,t1}}
	// in other words, we would like to replicaOrder the set T={t1,t2,t3}
	// here the transactions in T should meet some essential conditions
	n     int
	f     int
	multi int // default 1

	pavingMgr    *pavingMgr
	verifyingMgr *verifyingMgr
	graphingMgr  *graphingMgr

	// recorder ====================================================================
	// txsGraph
	// key: sequence number
	// val: replicas which have ordered the logs, the length of values should be fixed
	//      the length of each value should be equal to the amount of replicas
	// e.g. |___|s_1|s_2|s_3|...|s_n| <--key
	//      |r_1| 1 | 1 | 1 |...| 0 | <--val
	//      |r_2| 1 | 0 | 0 |...| 0 |   .
	//      |r_3| 1 | 1 | 0 |...| 0 |   .
	//      |...|...|...|...|...|...|   .
	//      |r_n| 1 | 1 | 1 |...| 0 | <--val
	//
	// txRecorder
	// indicates the essential messages for every transaction, including
	// 1) the replicas who have ordered current transaction
	// 2) the candidates for current transaction
	// 3) whether candidates have ordered the transaction or not
	//
	// vpRecorder
	// the log-replicaOrder from every replica, struct: replica id ==> transaction list
	//
	amountSeq  uint64
	txsGraph   map[uint64]map[uint64]string
	txRecorder map[string]utils.TxRecorder
	vpRecorder map[uint64]utils.TxList

	// timer =======================================================================
	// there are 3 types of timer for every sequence number:
	//
	// paving timer:
	// we need to trigger a timer for sequence number when there are n-f replicas have send
	// the ordered log. we would like to 'remove' the replicas which haven't send ordered
	// logs with specific sequence number if the timer expires.
	//
	// gathering timer:
	// we need to start a timer for a serial of transactions when we has collected them for
	// particular sequence number. we would like to 'remove' the transactions which haven't
	// been verified by efficient(n-f) replicas before the timer expires.
	//
	// appointing timer:
	// for every transaction we will appoint some replicas as candidates and we need to
	// receive the ordered logs from the candidates. if we cannot receive them before timer
	// expires, we will 'remove' the missed replica and re-select the candidates.
	//
	// notes:
	// only one instance will be maintained here for every type of timer, and such two types
	// of timer should be started one by one, which means we would like to track the only one
	// sequence number's legality at one moment.
	//
	whitelist    []int
	delay        time.Duration
	paving       bool
	gathering    bool
	appointing   bool
	pavedTxs     []string
	gatheredTxs  map[string]bool
	appointedTxs map[string]bool
	graphSize    int

	waiting []string

	// channel =====================================================================
	// replicaOrder:    channel used to deliver the ordered logs from replicas
	// graphEngine:     channel used to deliver the transactions to candidate_filter
	// pavingTimer:     channel used to process timeout events for paving check
	// pavingExit:      channel used to stop
	// gatheringTimer:  channel used to process timeout events for gathering check
	// gatheringExit:   channel used to stop
	// appointingTimer: channel used to process timeout events for appointing check
	// appointingExit:  channel used to stop
	// close:           channel used to stop
	//
	// replica_order ------------ replicaOrder -----> recorder
	// txsGraph ----------------- paving -----------> timeout or pavedTxs
	// pavedTxs, txRecorder ----- gathering --------> timeout or gatheredTxs
	// gatheredTxs, txRecorder -- appointing -------> timeout or appointedTxs
	// appointedTxs ------------- graphEngine ------> graph_engine
	//
	replicaOrder chan *pb.OrderedLog
	graphEngine  chan interface{}
	certStore map[types.RelationId]*types.RelationCert

	pavingTimer     chan bool
	pavingExit      chan bool
	gatheringTimer  chan bool
	gatheringExit   chan bool
	appointingTimer chan bool
	appointingExit  chan bool
	close           chan bool

	pavingRecvC    chan *pb.OrderedLog
	verifyingRecvC chan *pb.OrderedLog
	graphingRecvC  chan *pb.OrderedLog

	commC chan *pb.OrderedLog

	// logger
	logger    logger.Logger
}

func newTransactionsFilterImpl(c types.Config) *transactionsFilterImpl {
	n := len(c.Replicas)
	f := int(math.Floor((float64(n)-1)/4))
	if f == 0 {
		f = 1
	}
	multi := 1

	vpRecorderPaving := make(map[uint64]utils.TxList)
	for _, id := range c.Replicas {
		vpRecorderPaving[uint64(id)] = utils.NewTxList(c.Logger)
	}

	vpRecorderGraphing := make(map[uint64]utils.TxList)
	for _, id := range c.Replicas {
		vpRecorderGraphing[uint64(id)] = utils.NewTxList(c.Logger)
	}

	pavingRecvC := make(chan *pb.OrderedLog, tp.DefaultChannelLen)
	pavedC := make(chan types.PavedTxs)

	finishedC := make(chan []string)

	verifyingRecvC := make(chan *pb.OrderedLog, tp.DefaultChannelLen)
	verifyC := make(chan string, tp.DefaultChannelLen)

	graphingRecvC := make(chan *pb.OrderedLog, tp.DefaultChannelLen)

	closeC := make(chan bool)

	return &transactionsFilterImpl{
		n:     n,
		f:     f,
		multi: multi,

		pavingMgr:    newPavingMgr(n, f, vpRecorderPaving, pavingRecvC, pavedC, closeC, c.Logger, finishedC),
		verifyingMgr: newGatheringMgr(n, f, c.Replicas, verifyingRecvC, verifyC, closeC, c.Logger),
		graphingMgr:  newRelatingMgr(n, f, vpRecorderGraphing, graphingRecvC, verifyC, pavedC, closeC, c.Logger, finishedC),

		pavingRecvC:    pavingRecvC,
		verifyingRecvC: verifyingRecvC,
		graphingRecvC:  graphingRecvC,

		amountSeq:  uint64(0),
		txsGraph:   make(map[uint64]map[uint64]string),
		vpRecorder: nil,
		txRecorder: make(map[string]utils.TxRecorder),

		certStore: make(map[types.RelationId]*types.RelationCert),

		whitelist:    c.Replicas,
		delay:        6 * time.Second,
		paving:       false,
		gathering:    false,
		appointing:   false,
		pavedTxs:     nil,
		gatheredTxs:  make(map[string]bool),
		appointedTxs: make(map[string]bool),

		graphSize:       types.DefaultGraphSize,
		replicaOrder:    c.Order,
		graphEngine:     c.Graph,
		pavingTimer:     make(chan bool),
		pavingExit:      make(chan bool),
		gatheringTimer:  make(chan bool),
		gatheringExit:   make(chan bool),
		appointingTimer: make(chan bool),
		appointingExit:  make(chan bool),
		close:           make(chan bool),

		commC: make(chan *pb.OrderedLog),

		logger: c.Logger,
	}
}

func (tf *transactionsFilterImpl) start() {
	go tf.listenTimerEvent()
	go tf.pavingMgr.listener()
	go tf.verifyingMgr.listener()
	go tf.graphingMgr.listener()
}

func (tf *transactionsFilterImpl) stop() {
	close(tf.close)
	tf.stopPavingTimer()
	tf.stopGatheringTimer()
	tf.stopAppointingTimer()
}

func (tf *transactionsFilterImpl) listenTimerEvent() {
	for {
		select {
		case <-tf.close:
			return

		case log := <-tf.replicaOrder:
			tf.verifyingRecvC <- log
			tf.pavingRecvC <- log
			tf.graphingRecvC <- log

		case <-tf.graphEngine:


		case <-tf.pavingTimer:
			tf.stopPavingTimer()
			// TODO(wgr): trigger ba remove

		case <-tf.gatheringTimer:
			tf.stopGatheringTimer()
			// TODO(wgr): trigger ba remove

		case <-tf.appointingTimer:
			tf.stopAppointingTimer()
			// TODO(wgr): trigger ba remove
		}
	}
}
