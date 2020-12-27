package sequencepool

type SequencePool interface {
	Start()
	Stop()

	ReplicaManager
	ClientManager
}

type ReplicaManager interface {
	RegisterReplica(id uint64)
	UnregisterReplica(id uint64)
}

type ClientManager interface {
	RegisterClient(id uint64)
	UnregisterClient(id uint64)
}

func (sp *sequencePoolImpl) Start() {
	sp.start()
}

func (sp *sequencePoolImpl) Stop() {
	sp.stop()
}

func (sp *sequencePoolImpl) RegisterReplica(id uint64) {

}

func (sp *sequencePoolImpl) UnregisterReplica(id uint64) {

}

func (sp *sequencePoolImpl) RegisterClient(id uint64) {

}

func (sp *sequencePoolImpl) UnregisterClient(id uint64) {

}
