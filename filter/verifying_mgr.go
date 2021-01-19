package filter

import (
	"github.com/Grivn/libfalanx/filter/utils"
	"github.com/Grivn/libfalanx/logger"
	pb "github.com/Grivn/libfalanx/zcommon/protos"
)

type verifyingMgr struct {
	n         int
	f         int
	whitelist []int

	// recorder ====================================================================
	// txRecorder
	// indicates the essential messages for every transaction, including
	// 1) the replicas who have ordered current transaction
	// 2) the candidates for current transaction
	// 3) whether candidates have ordered the transaction or not
	txRecorder map[string]utils.TxRecorder

	pendingTxs  []string
	verifiedTxs map[string]bool

	recvC chan *pb.OrderedLog
	commC chan string
	close chan bool

	gathering bool

	logger logger.Logger
}

func newGatheringMgr(n, f int, whitelist []int, recvC chan *pb.OrderedLog, commC chan string, close chan bool, logger logger.Logger) *verifyingMgr {
	return &verifyingMgr{
		n:           n,
		f:           f,
		whitelist:   whitelist,
		txRecorder:  make(map[string]utils.TxRecorder),
		pendingTxs:  nil,
		verifiedTxs: make(map[string]bool),
		recvC:       recvC,
		commC:       commC,
		close:       close,
		gathering:   false,
		logger:      logger,
	}
}

func (v *verifyingMgr) listener() {
	for {
		select {
		case <-v.close:
			return

		case log := <-v.recvC:
			v.add(log)
			v.scanner()
		}
	}
}

func (v *verifyingMgr) add(log *pb.OrderedLog) {
	if log == nil {
		panic("nil log!")
	}

	// update txRecorder
	if v.txRecorder[log.TxHash] == nil {
		v.txRecorder[log.TxHash] = utils.NewTxRecorder(v.whitelist, log.TxHash, v.n, v.f)
	}
	v.txRecorder[log.TxHash].Add(log.ReplicaId)

	// update pending txs set
	if !v.verifiedTxs[log.TxHash] {
		v.pendingTxs = append(v.pendingTxs, log.TxHash)
	}
}

func (v *verifyingMgr) scanner() {
	if len(v.pendingTxs) == 0 {
		return
	}

	v.softStartTimer()
	txHash := v.pendingTxs[0]
	if v.verifiedTxs[txHash] || v.txRecorder[txHash].OrderLen() >= v.allQuorumReplicas() {
		v.stopTimer()

		if !v.verifiedTxs[txHash] {
			v.verifiedTxs[txHash] = true
			v.communicate(txHash)
		}

		v.pendingTxs = v.pendingTxs[1:]
		if len(v.pendingTxs) > 0 {
			v.scanner()
		}
	}
}

func (v *verifyingMgr) communicate(hash string) {
	v.commC <- hash
}

func (v *verifyingMgr) softStartTimer() {
	if v.gathering {
		return
	}
	v.gathering = true
}

func (v *verifyingMgr) stopTimer() {
	v.gathering = false
}

func (v *verifyingMgr) allReplicas() int {
	return v.n
}

func (v *verifyingMgr) allQuorumReplicas() int {
	return v.n- v.f
}