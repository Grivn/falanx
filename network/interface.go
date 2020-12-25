package network

import (
	baTypes "github.com/Grivn/libfalanx/localba/types"
	seqTypes "github.com/Grivn/libfalanx/sequencepool/types"
)

type Network interface {
	BroadcastSequenceLog(log seqTypes.SequenceLog)
	BroadcastSuspectMalice(sus baTypes.SuspectMalice)
}
