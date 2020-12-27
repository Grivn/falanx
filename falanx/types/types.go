package types

import "github.com/Grivn/libfalanx/sequencepool"

type Config struct {
	Author      uint64
	TxContainer sequencepool.TxsContainer
}