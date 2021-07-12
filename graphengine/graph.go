package graphengine

func NewGraphEngine(graphEngine chan interface{}) *graphEngineImpl {
	return newGraphEngineImpl(graphEngine)
}

func (g *graphEngineImpl) Start() {
	g.start()
}

func (g *graphEngineImpl) Stop() {
	g.stop()
}
