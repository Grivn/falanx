package graphengine

import (
	"github.com/Grivn/libfalanx/graphengine/types"
)

type graphEngineImpl struct {
	graphSize int

	// rawGraph
	// when we receive txs from filter, we would like to generate a raw graph at first.
	// here, we will assign a sequence number for every tx, and initiate the relation of txs
	// according to the filter.vpRecorder
	seqNo    uint64
	txMap    map[string]uint64
	rawGraph map[uint64]map[uint64]bool
	vertex   map[uint64]types.V

	dfn   uint64
	cVertex map[uint64]types.V

	graphEngine chan types.GraphEvent
	close  chan bool
}

func newGraphEngineImpl(graphEngine chan types.GraphEvent) *graphEngineImpl {
	return &graphEngineImpl{
		graphSize:   types.DefaultGraphSize,
		seqNo:       0,
		txMap:       make(map[string]uint64),
		graphEngine: graphEngine,
		close:       make(chan bool),
	}
}

func (ge *graphEngineImpl) start() {
	go ge.listenEvent()
}

func (ge *graphEngineImpl) stop() {
	close(ge.close)
}

func (ge *graphEngineImpl) listenEvent() {
	for {
		select {
		case <-ge.close:
			return

		case event := <-ge.graphEngine:
			switch event.Type {
			case types.TypeGraphPassedTxs:
				txs, ok := event.Event.([]string)
				if ok {
					ge.processPassedTxs(txs)
				}
			case types.TypeGraphTxsRelation:
				ge.processTxsRelation(event.Event)
			}
		}
	}
}

func (ge *graphEngineImpl) processPassedTxs(txs []string) {
	for _, hash := range txs {
		ge.seqNo++
		if len(ge.txMap) == ge.graphSize {
			ge.graphGenerator()
			ge.garbageCollector()
		}
		ge.txMap[hash] = ge.seqNo
	}
}

func (ge *graphEngineImpl) graphGenerator() {
	graph := make(map[uint64]map[uint64]bool, len(ge.txMap))
	for _, id := range ge.txMap {
		value := make(map[uint64]bool, len(ge.txMap))
		v := types.V{
			ID: id,
			DFN: 0,
			Low: 0,
			Pushed: false,
		}
		graph[id] = value
		ge.vertex[id] = v
	}
	ge.rawGraph = graph
}

func (ge *graphEngineImpl) garbageCollector() {

}

func (ge *graphEngineImpl) processTxsRelation(relation interface{}) {

}
