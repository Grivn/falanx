package filter

import (
	"github.com/Grivn/libfalanx/filter/types"
	"github.com/Grivn/libfalanx/filter/utils"
	"github.com/Grivn/libfalanx/logger"
	pb "github.com/Grivn/libfalanx/zcommon/protos"
	"math"
)

type graphingMgr struct {
	n int
	f int

	verifiedTxs map[string]bool

	vpRecorder map[uint64]utils.TxList

	certStore map[types.RelationId]*types.RelationCert

	graphSize int

	preferSeq uint64

	recvC   chan *pb.OrderedLog
	verifyC chan string
	pavedC  chan types.PavedTxs
	finishC chan []string
	close   chan bool

	waiting  []string
	finished []string
	executed map[string]bool
	graphing bool

	logger logger.Logger
}

func newRelatingMgr(n, f int, vpRecorder map[uint64]utils.TxList, recvC chan *pb.OrderedLog, verifyC chan string, pavedC chan types.PavedTxs, close chan bool, logger logger.Logger, delC chan []string) *graphingMgr {
	return &graphingMgr{
		certStore:   make(map[types.RelationId]*types.RelationCert),
		vpRecorder:  vpRecorder,
		verifiedTxs: make(map[string]bool),
		executed:    make(map[string]bool),
		graphSize:   types.DefaultGraphSize,
		recvC:       recvC,
		finishC:     delC,
		verifyC:     verifyC,
		pavedC:      pavedC,
		close:       close,
		graphing:    false,
		preferSeq:   1,
		logger:      logger,
	}
}

func (g *graphingMgr) listener() {
	for {
		select {
		case <-g.close:
			return

		case log := <-g.recvC:
			g.add(log)
			if !g.graphing {
				continue
			}
			g.logger.Debug("[GRAPH] trying to relate")
			g.relateTxs()

		case txHash := <-g.verifyC:
			g.verified(txHash)

		case pavedTxs := <-g.pavedC:
			if g.graphing {
				g.logger.Debug("[GRAPH] reject paved txs, in graphing")
				continue
			}
			if pavedTxs.Seq != g.preferSeq {
				g.logger.Debug("[GRAPH] reject paved txs, wrong batch seq")
				continue
			}
			g.generateGraph(pavedTxs)
		}
	}
}

func (g *graphingMgr) add(log *pb.OrderedLog) {
	if log == nil {
		panic("nil log!")
	}

	if g.executed[log.TxHash] {
		return
	}

	// update vpRecorder
	g.vpRecorder[log.ReplicaId].Add(log)
}

func (g *graphingMgr) verified(hash string) {
	g.verifiedTxs[hash] = true
}

func (g *graphingMgr) generateGraph(pavedTxs types.PavedTxs) {
	if pavedTxs.Seq != g.preferSeq {
		return
	}
	if len(pavedTxs.Txs) < g.graphSize {
		panic("not efficient paved txs!")
	}

	g.waiting = nil
	for txHash := range pavedTxs.Txs {
		if !g.verifiedTxs[txHash] {
			g.logger.Debugf("[Unverified] reject paved tx %s", txHash)
			return
		}
		g.waiting = append(g.waiting, txHash)
	}
	g.finished = nil
	g.graphing = true

	g.relateTxs()
}

func (g *graphingMgr) relateTxs() {
	if len(g.waiting) == 0 {
		return
	}

	self := g.waiting[0]

	finished := true
	for _, other := range g.waiting {
		if self == other {
			continue
		}
		switch g.check(self, other) {
		case types.FormerPriority:
			g.logger.Infof("%s ===> %s", self, other)
			cert := g.getRelationCert(other, self)
			if !cert.Finished {
				cert.Finished = true
				cert.Status = types.LatterPriority
			}
		case types.LatterPriority:
			g.logger.Infof("%s ===> %s", other, self)
			cert := g.getRelationCert(other, self)
			if !cert.Finished {
				cert.Finished = true
				cert.Status = types.FormerPriority
			}
		case types.NotEfficient:
			g.logger.Infof("cannot compare %s and %s", self, other)
			finished = false
		}
	}
	if finished {
		g.waiting = g.waiting[1:]
		g.finished = append(g.finished, self)
	}

	if len(g.waiting) == 0 {
		g.generateRawGraph()
	}
}

func (g *graphingMgr) check(former, latter string) types.BeforeCheck {
	cert := g.getRelationCert(former, latter)

	if cert.Finished {
		// we have already finished the determination of the order between former and latter
		return cert.Status
	}

	for id, vp := range g.vpRecorder {
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

		if cert.FormerPreferred > g.moreThanHalf() {
			// more than half replicas have decided the former one has a higher priority
			cert.Status = types.FormerPriority
			cert.Finished = true
			break
		}
		if cert.LatterPreferred > g.moreThanHalf() {
			// more than half replicas have decided the latter one has a higher priority
			cert.Status = types.LatterPriority
			cert.Finished = true
			break
		}
	}

	return cert.Status
}

func (g *graphingMgr) getRelationCert(former, latter string) *types.RelationCert {
	idr := types.RelationId{From: former, To: latter}
	value, ok := g.certStore[idr]

	if ok {
		return value
	}

	cert := &types.RelationCert{
		Finished:        false,
		Status:          types.NotEfficient,
		Scanned:         make(map[uint64]bool),
		FormerPreferred: 0,
		LatterPreferred: 0,
	}
	g.certStore[idr] = cert
	return cert
}

func (g *graphingMgr) generateRawGraph() {
	if g.certStore == nil {
		return
	}

	graph := make(map[string][]string)
	g.logger.Infof("Trying to generate graph")

	for idr, cert := range g.certStore {
		if cert.Status == types.FormerPriority {
			graph[idr.From] = append(graph[idr.From], idr.To)
		}
	}

	g.graphing = false
	g.printGraph(graph)
	g.preferSeq++

	g.finish()
}

func (g *graphingMgr) printGraph(graph map[string][]string) {
	for from := range graph {
		toList := graph[from]
		g.logger.Infof("%s out degree is %d", from, len(toList))
		for _, to := range toList {
			g.logger.Infof("    ===> %s", to)
		}
	}
}

func (g *graphingMgr) finish() {
	g.logger.Infof("============================ Call execute %d ============================", g.preferSeq-1)
	for _, txHash := range g.finished {
		g.logger.Infof("[EXEC] %s", txHash)
		g.executed[txHash] = true
		for _, vp := range g.vpRecorder {
			vp.RemoveByHash(txHash)
		}
	}
	g.certStore = make(map[types.RelationId]*types.RelationCert)
	go g.inform()
}

func (g *graphingMgr) inform() {
	g.logger.Infof("[GRAPH] post finished event")
	g.finishC <- g.finished
}

func (g *graphingMgr) allReplicas() int {
	return g.n
}

func (g *graphingMgr) allQuorumReplicas() int {
	return g.n- g.f
}

func (g *graphingMgr) moreThanHalf() int {
	return int(math.Ceil(float64(g.n+1)/float64(2)))
}