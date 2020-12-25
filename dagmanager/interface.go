package dagmanager

import "github.com/Grivn/libfalanx/dagmanager/types"

type DAGManager interface {
	// Extend is used to extend DAG graph
	Extend(value *types.DAGValue)

	// Execute is used to fetch values from DAG to call execute
	Execute()

	// GetGraph is used to get the graph of DAG
	GetGraph() *types.DAG
}
