package graphengine

import "github.com/Grivn/libfalanx/graphengine/types"

type GraphEngine interface {
	// GraphAnalyzer is used to find a strongly connected rawGraph
	// we could use Tarjan or Kosaraju algorithm
	GraphAnalyzer(g *types.Graph) ([]*types.Graph, error)

	// RawGraphGenerator is used to generate a relation rawGraph which might have strongly connected sub-rawGraph
	RawGraphGenerator() (*types.Graph, error)

	// AcyclicGraphGenerator is used to generate the rawGraph to make finalization decision
	AcyclicGraphGenerator() (*types.Graph, error)

	// VertexPicker is used to pick the legal value to call execute
	VertexPicker() []*types.Vertex
}
