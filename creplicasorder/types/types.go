package types

// TimeoutEvent is used to process local ba progress
type TimeoutEvent struct {
	// TxHash is used to track the particular transaction waiting for quorum candidates
	TxHash string
}
