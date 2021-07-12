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

func (g *graphEngineImpl) start() {
	go g.listenEvent()
}

func (g *graphEngineImpl) stop() {
	close(g.close)
}

func (g *graphEngineImpl) listenEvent() {
	for {
		select {
		case <-g.close:
			return

		case event := <-g.graphEngine:
			switch e := event.(type) {
			case map[string][]string:
				g.printGraph(e)
			}
		}
	}
}

func (g *graphEngineImpl) printGraph(graph map[string][]string) {
	for from := range graph {
		toList := graph[from]
		g.logger.Infof("%s out degree is %d", from, len(toList))
		for to := range toList {
			g.logger.Infof("    ===> %s", to)
		}
	}
}
