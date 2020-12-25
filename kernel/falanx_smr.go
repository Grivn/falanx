package kernel

import (
	"github.com/Grivn/libfalanx/dagmanager"
	"github.com/Grivn/libfalanx/executor"
	"github.com/Grivn/libfalanx/graphengine"
	"github.com/Grivn/libfalanx/localba"
	"github.com/Grivn/libfalanx/network"
	"github.com/Grivn/libfalanx/sequencepool"
	"github.com/Grivn/libfalanx/statusmanager"
	"github.com/Grivn/libfalanx/storage"
)

type FalanxSMR struct {
	Event *EventProcessor
	DAG *dagmanager.DAGProcessor
	Execute *executor.ExecuteProcessor
	Graph *graphengine.GraphProcessor
	BA *localba.BAProcessor
	Status *statusmanager.StatusProcessor

	Pool sequencepool.SequencePool
	Sender network.Network
	Storage storage.Storage
}
