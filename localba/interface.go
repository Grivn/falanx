package localba

type BAEngine interface {
	// ElectCandidates is used to find a list of nodes to make finalization
	ElectCandidates() []string

	// RemoveCandidate is used to remove the node whose network might be abnormal
	RemoveCandidate() error
}
