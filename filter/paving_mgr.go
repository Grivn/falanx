package filter

import (
	"github.com/Grivn/libfalanx/filter/types"
	"github.com/Grivn/libfalanx/filter/utils"
	"github.com/Grivn/libfalanx/logger"
	pb "github.com/Grivn/libfalanx/zcommon/protos"
)

type pavingMgr struct {
	n         int
	f         int
	whitelist []int

	round uint64
	pavedTxs map[string]bool

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
	// the log-replicaOrder from every replica, struct: replica id ==> transaction list
	//
	//txsGraph   map[uint64]map[uint64]string
	vpRecorder map[uint64]utils.TxList

	pavedRecorder map[uint64]map[string]bool

	recvC chan *pb.OrderedLog
	commC chan types.PavedTxs
	delC  chan []string
	close chan bool

	batchSeq uint64

	maxLen int

	logger logger.Logger
}

func newPavingMgr(n, f int, vpRecorder map[uint64]utils.TxList, recvC chan *pb.OrderedLog, commC chan types.PavedTxs, close chan bool, logger logger.Logger, delC chan []string) *pavingMgr {
	return &pavingMgr{
		n:           n,
		f:           f,
		round:       0,
		//txsGraph:    make(map[uint64]map[uint64]string),
		pavedTxs:    make(map[string]bool),
		batchSeq:    1,
		vpRecorder:  vpRecorder,
		recvC:       recvC,
		delC:        delC,
		commC:       commC,
		close:       close,
		maxLen:      types.DefaultGraphSize,
		logger:      logger,
	}
}

func (p *pavingMgr) listener() {
	for {
		select {
		case <-p.close:
			return

		case log := <-p.recvC:
			p.add(log)
			p.scanner()

		case finishedTxs := <-p.delC:
			p.finish(finishedTxs)
			p.scanner()
		}
	}
}

func (p *pavingMgr) add(log *pb.OrderedLog) {
	if log == nil {
		panic("nil log!")
	}

	// update vpRecorder
	p.vpRecorder[log.ReplicaId].Add(log)
}

func (p *pavingMgr) finish(finishedTxs []string) {
	p.logger.Infof("[PAVE] received finished event, try to remove")
	for _, txHash := range finishedTxs {
		for _, vp := range p.vpRecorder {
			vp.RemoveByHash(txHash)
		}
	}
	p.batchSeq++
	p.round = 0
	p.pavedTxs = make(map[string]bool)
}

func (p *pavingMgr) scanner() {
	round := p.round
	for len(p.pavedTxs) < p.maxLen {
		id := p.roundID(round)
		seq := p.roundSEQ(round)
		p.logger.Infof("[PAVE] read log, (%d, %d)", id, seq)
		log := p.vpRecorder[id].GetByOrder(int(seq))
		if log == nil {
			if round > p.round {
				p.round = round
			}
			p.logger.Infof("[PAVE] not efficient txs, len %d, round %d", len(p.pavedTxs), p.round)
			return
		}
		p.pavedTxs[log.TxHash] = true
		round++
	}
	p.communicate()
}

func (p *pavingMgr) communicate() {
	comm := types.PavedTxs{
		Seq: p.batchSeq,
		Txs: p.pavedTxs,
	}
	p.commC <- comm
}

func (p *pavingMgr) roundID(round uint64) uint64 {
	return round%uint64(p.n) + 1
}

func (p *pavingMgr) roundSEQ(round uint64) uint64 {
	return round/uint64(p.n) + 1
}
