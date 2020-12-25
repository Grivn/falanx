package network

type Network interface {
	BroadcastSequenceLog(log []byte)
	BroadcastSuspectMalice(sus []byte)
}
