package types

type NodeStatus struct {
	ID     uint64
	H      uint64
	Status StatusType
}

// StatusType defines the Falanx internal status.
type StatusType int

// consensus status type.
const (
	// normal status, which will only be used in falanx
	Normal = iota + 1 // normal consensus state

	// atomic status, which might be used by consensus service
	Pending // node cannot process consensus messages

	// internal status
	byzantine // byzantine
)
