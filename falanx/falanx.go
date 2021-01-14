package falanx

import "github.com/Grivn/libfalanx/zcommon/types"

type Falanx interface {
	Start()
	Stop()
}

func NewFalanx(c types.Config) *falanxImpl {
	return newFalanxImpl(c)
}

func (falanx *falanxImpl) Start() {
	falanx.start()
}

func (falanx *falanxImpl) Stop() {
	falanx.stop()
}
