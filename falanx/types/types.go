package types

import (
	"github.com/Grivn/libfalanx/atxcontainer"
)

type Config struct {
	Author      uint64
	TxContainer atxcontainer.TxsContainer
}