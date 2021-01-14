package filter

import (
	"container/list"
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

	// verifier for transactions
	// to check the client
	verifiedSeq uint64
	pendingTxs  utils.TxList
	verifiedTxs utils.TxList

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
	pavedSeq   uint64
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
	pavedTxs     utils.TxList
	waiting      []string
	gatheredTxs  []string
	appointedTxs map[string]bool
	graphSize    int

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

	graph map[string]map[string]bool
	certStore map[types.RelationId]*types.RelationCert

	pavingTimer     chan bool
	pavingExit      chan bool
	gatheringTimer  chan bool
	gatheringExit   chan bool
	appointingTimer chan bool
	appointingExit  chan bool
	close           chan bool

	// logger
	logger logger.Logger
}

func newTransactionsFilterImpl(c types.Config) *transactionsFilterImpl {
	n := len(c.Replicas)
	f := int(math.Floor((float64(n)-1)/4))
	multi := 1

	vpRecorder := make(map[uint64]utils.TxList)
	for _, id := range c.Replicas {
		vpRecorder[uint64(id)] = utils.NewTxList()
	}
	return &transactionsFilterImpl{
		n:     n,
		f:     f,
		multi: multi,

		pavedSeq:   uint64(0),
		txsGraph:   make(map[uint64]map[uint64]string),
		vpRecorder: vpRecorder,

		whitelist:    c.Replicas,
		delay:        6 * time.Second,
		paving:       false,
		gathering:    false,
		appointing:   false,
		pavedTxs:     nil,
		gatheredTxs:  nil,
		appointedTxs: make(map[string]bool),

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

	if !tf.verifiedTxs.Has(l.TxHash) {
		// there is an unverified tx, append it into the pending tx set
		tf.pendingTxs.Add(l)
	}

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

	// firstly, try to check the pending txs to guarantee there are efficient replicas have ordered them
	tf.gatheringScanner()

	// we should finish the paving stage at first to find a stable set
	tf.pavingScanner()
	if tf.paving {
		// try to pave
		// we should return to receive more ordered logs from replicas
		return true
	}

	return false
}

func (tf *transactionsFilterImpl) gatheringScanner() {
	if tf.pendingTxs.Len() == 0 {
		// no txs to be checked
		return
	}

	if !tf.gathering {
		tf.startGatheringTimer(tf.gatheringExit)
	}

	log := tf.pendingTxs.FrontLog()
	txHash := log.TxHash
	if tf.txRecorder[txHash].OrderLen() >= tf.allQuorumReplicas() {
		tf.stopGatheringTimer()

		// efficient replicas have ordered such a transaction,
		// and we will remove the tx from pending txs and update the verified set
		tf.pendingTxs.RemoveLog(txHash)
		tf.verifiedTxs.Add(log)
		tf.verifiedSeq++

		if tf.pendingTxs.Len() > 0 {
			// pending transactions there need to be checked, restart another round of gathering scanner
			tf.gatheringScanner()
		}
	}

	// ===============================================================
	// we are trying to gather efficient ordered logs for one tx now
	// return to receive more logs
	// ===============================================================
	return
}

func (tf *transactionsFilterImpl) pavingScanner() {
	if tf.verifiedTxs.Len() == 0 {
		// there aren't any verified transactions for us to pave
		return
	}

	if tf.appointing {
		return
	}

	nextPavingSeq := tf.pavedSeq + 1
	if nextPavingSeq > tf.verifiedSeq {
		// there aren't efficient verified transactions for us to pave
		return
	}

	nextLen := len(tf.txsGraph[nextPavingSeq])
	if nextLen < tf.allQuorumReplicas() {
		// we won't try to pave until the length of next serial of txs no less than n-f
		return
	}

	if !tf.paving {
		// we have received ordered log with the same sequence number from n-f replicas
		// here, we will start a timer for it until all the replicas order the log for
		// current sequence number
		tf.startPavingTimer(tf.pavingExit)
	}

	if nextLen == tf.allReplicas() {
		// all the replicas has an ordered log on such a sequence number, which means 'paved'
		for _, hash := range tf.txsGraph[tf.pavedSeq] {
			log := tf.verifiedTxs.GetLog(hash)
			if log == nil {
				// we need to guarantee that all the paved txs are verified by efficient replicas
				return
			}
			tf.pavedTxs.Add(log)
		}
		tf.pavedSeq++
		tf.stopPavingTimer()

		if tf.pavedTxs.Len() > tf.graphSize {
			var (
				hashList []string
				next     *list.Element
			)
			i := 0
			for element := tf.pavedTxs.Front(); i < tf.pavedTxs.Len(); element = next {
				log, ok := element.Value.(*pb.OrderedLog)
				if !ok {
					return
				}
				i++
				next = element.Next()
				hash := log.TxHash
				hashList = append(hashList, hash)
			}
			tf.waiting = hashList
			tf.relationScanner()
		}
	}
}

func (tf *transactionsFilterImpl) relationScanner() {
	if len(tf.waiting) == 0 {
		return
	}

	if !tf.appointing {
		tf.startAppointingTimer(tf.appointingExit)
	}

	if tf.graph == nil {
		tf.graph = make(map[string]map[string]bool)
	}

	log := tf.pavedTxs.FrontLog()
	if log == nil {
		return
	}
	self := log.TxHash

	if tf.graph[self] == nil {
		tf.graph[self] = make(map[string]bool)
	}

	finished := true
	for _, other := range tf.waiting {
		if self == other {
			continue
		}
		switch tf.check(self, other) {
		case types.FormerPriority:
			cert := tf.getRelationCert(other, self)
			if !cert.Finished {
				cert.Finished = true
				cert.Status = types.LatterPriority
			}
		case types.LatterPriority:
			cert := tf.getRelationCert(other, self)
			if !cert.Finished {
				cert.Finished = true
				cert.Status = types.FormerPriority
			}
		case types.NotEfficient:
			finished = false
		}
	}

	if finished {
		tf.pavedTxs.RemoveLog(self)
	}

	if tf.pavedTxs.Len() == 0 {
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

	graph := make(map[string]map[string]bool)

	for idr, cert := range tf.certStore {
		if cert.Status == types.FormerPriority {
			if graph[idr.From] == nil {
				graph[idr.From] = make(map[string]bool)
			}
			graph[idr.From][idr.To] = true
		}
	}

	tf.graphEngine <- graph
}
