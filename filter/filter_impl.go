package filter

type filterImpl struct {
	// transactionsFilter
	// transactions will be send into such a filter at first to find if there is a malicious client

	// candidatesFilter
	// for every transaction delivered from transaction filter
	// we need to initiate a candidates filter
	// to find selected quorum replicas to order
}
