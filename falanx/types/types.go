package types

import (
	"github.com/Grivn/libfalanx/txcontainer"
)

type Config struct {
	Author      uint64
	TxContainer txcontainer.TxsContainer
}