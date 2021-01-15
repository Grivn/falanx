package filter

import (
	"github.com/Grivn/libfalanx/filter/types"
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

	vpRecorder := make(map[uint64]utils.TxList)
	for _, id := range c.Replicas {
		vpRecorder[uint64(id)] = utils.NewTxList()
	}
	return &transactionsFilterImpl{
		n:     n,
		f:     f,
		multi: multi,

		amountSeq:  uint64(0),
		txsGraph:   make(map[uint64]map[uint64]string),
		vpRecorder: vpRecorder,
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

		logger: c.Logger,
	}
}

func (tf *transactionsFilterImpl) start() {
	go tf.listenTimerEvent()
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
			tf.add(log)
			tf.scanner()

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

func (tf *transactionsFilterImpl) add(l *pb.OrderedLog) {
	// update txsGraph
	if tf.txsGraph[l.Sequence] == nil {
		tf.txsGraph[l.Sequence] = make(map[uint64]string)
	}
	tf.txsGraph[l.Sequence][l.ReplicaId] = l.TxHash

	// update vpRecorder
	tf.vpRecorder[l.ReplicaId].Add(l)

	// update txRecorder
	if tf.txRecorder[l.TxHash] == nil {
		tf.txRecorder[l.TxHash] = utils.NewTxRecorder(tf.whitelist, l.TxHash, tf.n, tf.f)
	}
	tf.txRecorder[l.TxHash].Add(l.ReplicaId)
}

func (tf *transactionsFilterImpl) scanner() bool {

	// we should finish the paving stage at first to find a stable set
	tf.pavingScanner()
	if tf.paving {
		// we are scanning amount now
		// we should return to receive more ordered logs from replicas
		return true
	}

	// we should make sure that every tx in stable set has been ordered by efficient replicas
	// and we would like to gather the txs one by one
	tf.gatheringScanner()
	if tf.gathering {
		// we are scanning verification now
		// we should return to receive more ordered logs from replicas
		return true
	}

	// after we finish the gathering of stable set, we would like to try to scan the candidate
	// replicas to find if they have ordered the tx or not
	tf.relationScanner()
	if tf.appointing {
		// we are scanning appointed candidates of gathered transaction now
		// we should return to receive more ordered logs from replicas
		return true
	}

	return false
}

func (tf *transactionsFilterImpl) pavingScanner() {
	if tf.gathering {
		// it indicates that we are now gathering the ordered logs
		// from efficient replicas
		return
	}

	if tf.appointing {
		// it indicates there is a series of transaction waiting to
		// generate a execution graph
		return
	}

	nextAmountSeq := tf.amountSeq + 1
	nextLen := len(tf.txsGraph[nextAmountSeq])
	tf.logger.Noticef("Trying to pave %d, now %d, need %d", nextAmountSeq, nextLen, tf.allReplicas())
	for id, hash := range tf.txsGraph[nextAmountSeq] {
		tf.logger.Noticef("    [%d ===> %s]", id, hash)
	}
	if nextLen < tf.allQuorumReplicas() {
		// we won't process the amount check until the length of next serial of txs no less than n-f
		return
	}
	if !tf.paving {
		// we have received ordered log with the same sequence number from n-f replicas
		// here, we will start a timer for it until all the replicas order the log for
		// current sequence number
		tf.startPavingTimer(tf.pavingExit)
	}
	if nextLen == tf.allReplicas() {
		tf.stopPavingTimer()
		var txList []string
		for _, hash := range tf.txsGraph[tf.amountSeq] {
			txList = append(txList, hash)
		}
		tf.pavedTxs = txList
		tf.amountSeq++
		return
	}
	return
}

func (tf *transactionsFilterImpl) gatheringScanner() {
	if len(tf.pavedTxs) == 0 {
		// it indicates that we haven't found a stable paved set yet and we cannot try to
		// gather the ordered logs in this condition
		return
	}
	if tf.appointing {
		// it indicates there is a series of transaction waiting to
		// generate a execution graph
		return
	}

	if !tf.gathering {
		tf.startGatheringTimer(tf.gatheringExit)
	}

	txHash := tf.pavedTxs[0]
	if tf.txRecorder[txHash].OrderLen() >= tf.allQuorumReplicas() {
		tf.stopGatheringTimer()
		tf.gatheredTxs[txHash] = true
		tf.pavedTxs = tf.pavedTxs[1:]
		if len(tf.pavedTxs) > 0 {
			tf.gatheringScanner()
		}
		if len(tf.gatheredTxs) >= types.DefaultGraphSize {
			tf.waiting = nil
			for hash := range tf.gatheredTxs {
				tf.waiting = append(tf.waiting, hash)
			}
			tf.relationScanner()
		}
	}

	// ===============================================================
	// we are scanning verification now
	// there aren't efficient replicas order current transaction
	// we should return to receive more ordered logs from replicas
	// ===============================================================
	return
}

func (tf *transactionsFilterImpl) relationScanner() {
	if len(tf.waiting) == 0 {
		return
	}

	if tf.paving {
		return
	}

	if tf.gathering {
		// it indicates that current node is trying to gather efficient ordered logs
		// we cannot start the scanner of candidates now
		return
	}

	if len(tf.gatheredTxs) == 0 {
		// if there is non-transaction paved or gathered, it means we need try to pave
		// the floor at first and we don't need to scan particular replicas for tx
		return
	}

	tf.logger.Noticef("Trying to gather the relationship")

	if !tf.appointing {
		tf.startAppointingTimer(tf.appointingExit)
	}

	self := tf.waiting[0]

	finished := true
	for _, other := range tf.waiting {
		if self == other {
			continue
		}
		switch tf.check(self, other) {
		case types.FormerPriority:
			tf.logger.Noticef("%s ===> %s", self, other)
			cert := tf.getRelationCert(other, self)
			if !cert.Finished {
				cert.Finished = true
				cert.Status = types.LatterPriority
			}
		case types.LatterPriority:
			tf.logger.Noticef("%s ===> %s", other, self)
			cert := tf.getRelationCert(other, self)
			if !cert.Finished {
				cert.Finished = true
				cert.Status = types.FormerPriority
			}
		case types.NotEfficient:
			tf.logger.Noticef("cannot compare %s and %s", self, other)
			finished = false
		}
	}

	if finished {
		tf.waiting = tf.waiting[1:]
	}

	if len(tf.waiting) == 0 {
		tf.stopAppointingTimer()
		tf.generateRawGraph()
		tf.scanner()
	}

	// we are trying to generate relation graph
	// and we need to receive more ordered logs to determinate the relation
	return
}

func (tf *transactionsFilterImpl) check(former, latter string) types.BeforeCheck {
	cert := tf.getRelationCert(former, latter)

	if cert.Finished {
		// we have already finished the determination of the order between former and latter
		return cert.Status
	}

	for id, vp := range tf.vpRecorder {
		if cert.Scanned[id] {
			continue
		}
		seqFormer, errFormer := vp.GetSequence(former)
		seqLatter, errLatter := vp.GetSequence(latter)
		if errFormer != nil || errLatter != nil {
			// current replica's logs cannot determinate the order between former and latter
			continue
		}

		// current scanned replica has provided effective relation reference
		cert.Scanned[id] = true
		if seqFormer < seqLatter {
			// current replica believes that the former has a higher priority
			cert.FormerPreferred++
		} else {
			// current replica believes that the latter has a higher priority
			cert.LatterPreferred++
		}

		if cert.FormerPreferred > tf.moreThanHalf() {
			// more than half replicas have decided the former one has a higher priority
			cert.Status = types.FormerPriority
			cert.Finished = true
			break
		}
		if cert.LatterPreferred > tf.moreThanHalf() {
			// more than half replicas have decided the latter one has a higher priority
			cert.Status = types.LatterPriority
			cert.Finished = true
			break
		}
	}

	return cert.Status
}

func (tf *transactionsFilterImpl) generateRawGraph() {
	if tf.certStore == nil {
		return
	}

	graph := make(map[string][]string)
	tf.logger.Noticef("Trying to generate graph")

	for idr, cert := range tf.certStore {
		if cert.Status == types.FormerPriority {
			//if graph[idr.From] == nil {
			//	graph[idr.From] = make(map[string]bool)
			//}
			graph[idr.From] = append(graph[idr.From], idr.To)
		}
	}

	tf.gatheredTxs = make(map[string]bool)

	tf.printGraph(graph)
	//tf.graphEngine <- graph
}

func (tf *transactionsFilterImpl) printGraph(graph map[string][]string) {
	for from := range graph {
		toList := graph[from]
		tf.logger.Noticef("%s out degree is %d", from, len(toList))
		for _, to := range toList {
			tf.logger.Noticef("    ===> %s", to)
		}
	}
}