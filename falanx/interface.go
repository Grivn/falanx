package falanx

type Kernel interface {
	// Start starts a Falanx node instance.
	Start() error
	// Stop performs any necessary termination of the Node.
	Stop()

	// Propose proposes consensus messages to Falanx core,
	// messages are ensured to be eventually submitted to all non-fault nodes
	// unless current node crash down.
	Propose(requests [][]byte) error
	// Step advances the state machine using the given message.
	Step()
}
