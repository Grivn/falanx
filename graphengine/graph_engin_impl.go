package graphengine

import (
	"github.com/Grivn/libfalanx/graphengine/types"
	"github.com/Grivn/libfalanx/logger"
)

type graphEngineImpl struct {
	graphSize int

	// rawGraph
	// when we receive txs from filter, we would like to generate a raw graph at first.
	// here, we will assign a sequence number for every tx, and initiate the relation of txs
	// according to the filter.vpRecorder
	rawG types.Graph
	rawV map[string]types.V

	seqNo uint64
	txMap map[string]uint64
	idMap map[uint64]string

	dfn   uint64
	cVertex map[uint64]types.V

	graphEngine chan interface{}
	close  chan bool

	logger logger.Logger
}

func newGraphEngineImpl(graphEngine chan interface{}) *graphEngineImpl {
	return &graphEngineImpl{
		graphSize:   types.DefaultGraphSize,
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
			switch e := event.(type) {
			case map[string]map[string]bool:
				ge.printGraph(e)
			}
		}
	}
}

func (ge *graphEngineImpl) printGraph(graph map[string]map[string]bool) {
	for from := range graph {
		toMap := graph[from]
		if toMap == nil {
			ge.logger.Debugf("out degree is 0")
			continue
		}
		ge.logger.Debugf("out degree is %d:", len(toMap))
		for to := range toMap {
			ge.logger.Debugf("    ===> %s", to)
		}
	}
}
