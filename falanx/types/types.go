package types

import (
	"github.com/Grivn/libfalanx/api"
)

type Config struct {
	Author      uint64
	TxContainer api.TxsContainer
}