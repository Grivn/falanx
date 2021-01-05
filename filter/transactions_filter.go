package filter

import (
	"github.com/Grivn/libfalanx/zcommon/protos"
	"math"
	"time"

	"github.com/Grivn/libfalanx/filter/utils"
)

type TransactionsFilter interface {

}

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
	// we would like to order the set {{t1,t2,t2,t3},{t2,t1,t1,t1}}
	// in other words, we would like to order the set T={t1,t2,t3}
	// here the transactions in T should meet some essential conditions
	n     uint64
	f     uint64
	multi uint64 // default 1

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
	// vpRecorder
	// the log-order from every replica, struct: replica id ==> transaction list
	//
	// txRecorder
	// indicates the essential messages for every transaction, including
	// 1) the replicas who have ordered current transaction
	// 2) the candidates for current transaction
	// 3) whether candidates have ordered the transaction or not
	//
	amountSeq  uint64
	verifySeq  uint64
	txsGraph   map[uint64]map[uint64]string
	vpRecorder map[uint64]utils.TxList
	txRecorder map[string]utils.TxRecorder

	// timer =======================================================================
	// there are 2 types of timer for every sequence number:
	//
	// amount timer:
	// we need to trigger a timer for sequence number when there are n-f replicas have send
	// the ordered log. we would like to 'remove' the replicas which haven't send ordered
	// logs with specific sequence number if the timer expires
	//
	// verify timer:
	// we need to start a timer for a serial of transactions when we has collected them for
	// particular sequence number. we would like to 'remove' the transactions which haven't
	// been verified by efficient(n-f) replicas before the timer expires.
	//
	// notes:
	// only one instance will be maintained here for every type of timer, and such two types
	// of timer should be started one by one, which means we would like to track the only one
	// sequence number's legality at one moment.
	//
	whitelist []int
	delay     time.Duration
	inAmount  bool
	inVerify  bool
	pendingTx []string
	passedTx  []string

	// channel =====================================================================
	// order:       channel used to deliver the ordered logs from replicas
	// passed:      channel used to deliver the transactions to candidate_filter
	// amountTimer: channel used to process timeout events for amount check
	// verifyTimer: channel used to process timeout events for verify check
	// close:       channel used to stop go-routine
	//
	// replica_order ------- order -----------> recorder
	// recorder ------------ amountTimer -----> timeout or finished_amount
	// finished_amount ----- verifyTimer -----> timeout or finished_verify
	// finished_verify ----- passed ----------> candidate_filter
	//
	order          chan *protos.OrderedLog
	passed         chan []string
	amountTimer    chan bool
	amountExit     chan bool
	amountFinished chan bool
	verifyTimer    chan bool
	verifyExit     chan bool
	verifyFinished chan bool
	close          chan bool
}

func newTransactionsFilterImpl(n uint64, order chan *protos.OrderedLog, passed chan []string) *transactionsFilterImpl {
	f := uint64(math.Floor((float64(n)-1)/4))
	vpRecorder := make(map[uint64]utils.TxList)
	multi := uint64(1)
	for i:=uint64(0); i<n; i++ {
		vpRecorder[i+1] = utils.NewTxList()
	}
	return &transactionsFilterImpl{
		n:           n,
		f:           f,
		multi:       multi,
		amountSeq:   uint64(0),
		verifySeq:   uint64(0),
		vpRecorder:  vpRecorder,
		order:       order,
		passed:      passed,
		amountTimer: make(chan bool),
		amountExit:  make(chan bool),
		verifyTimer: make(chan bool),
		verifyExit:  make(chan bool),
		close:       make(chan bool),
	}
}

func (tf *transactionsFilterImpl) listenTimerEvent() {
	for {
		select {
		case <-tf.close:
			return

		case <-tf.amountTimer:
			tf.stopAmountTimer()
			// TODO(wgr): trigger ba remove

		case <-tf.verifyTimer:
			tf.stopVerifyTimer()
			// TODO(wgr): trigger ba remove

		case <-tf.amountFinished:
			tf.stopAmountTimer()
			tf.amountSeq++
			var txList []string
			for _, hash := range tf.txsGraph[tf.amountSeq] {
				txList = append(txList, hash)
			}
			tf.pendingTx = txList

		case <-tf.verifyFinished:
			tf.stopVerifyTimer()
			tf.verifySeq++
		}
	}
}

func (tf *transactionsFilterImpl) listenOrderedLogs() {
	for {
		select {
		case <-tf.close:
			return

		case log := <-tf.order:
			tf.add(log)
			tf.amountScanner()
			tf.verifyScanner()
		}
	}
}

func (tf *transactionsFilterImpl) add(l *protos.OrderedLog) {
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

func (tf *transactionsFilterImpl) amountScanner() {
	nextAmountSeq := tf.amountSeq + 1

	if len(tf.pendingTx) != 0 {
		return
	}
	if len(tf.txsGraph[nextAmountSeq]) >= tf.allQuorumReplicas() && !tf.inAmount {
		tf.startAmountTimer(tf.amountExit)
	}
	if len(tf.txsGraph[nextAmountSeq]) == tf.allReplicas() {
		<-tf.amountFinished
	}
}

func (tf *transactionsFilterImpl) verifyScanner() {

}

func (tf *transactionsFilterImpl) startAmountTimer(exitCh chan bool) {
	go func() {
		tf.inAmount = true
		timer := time.NewTimer(tf.delay)
		select {
		case <-timer.C:
			tf.amountTimer <- true
		case <-exitCh:
			return
		}
	}()
}

func (tf *transactionsFilterImpl) stopAmountTimer() {
	close(tf.amountExit)
	tf.amountExit = make(chan bool)
	tf.inAmount = false
}

func (tf *transactionsFilterImpl) startVerifyTimer(exitCh chan bool) {
	go func() {
		tf.inVerify = true
		timer := time.NewTimer(tf.delay)
		select {
		case <-timer.C:
			tf.amountTimer <- true
		case <-exitCh:
			return
		}
	}()
}

func (tf *transactionsFilterImpl) stopVerifyTimer() {
	close(tf.verifyExit)
	tf.verifyExit = make(chan bool)
	tf.inVerify = false
}

func (tf *transactionsFilterImpl) allReplicas() int {
	return int(tf.n)
}

func (tf *transactionsFilterImpl) allQuorumReplicas() int {
	return int(tf.n-tf.f)
}
