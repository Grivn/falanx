package types

type NetworkReceiver struct {
	SequenceReq chan []byte
	SequenceLog chan []byte
}
