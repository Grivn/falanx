package executor

type Executor interface {
	// Execute
	Execute(value [][]byte)
}
