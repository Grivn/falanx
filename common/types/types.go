package types

import fCommonProtos "github.com/ultramesh/flato-common/types/protos"

// const length
const (
	HashLength = 32
)

type (
	// Hash type used in falanx
	Hash [HashLength]byte

	Transaction *fCommonProtos.Transaction

	Author uint64
)
