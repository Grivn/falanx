package localorder

import "github.com/Grivn/libfalanx/localorder/types"

func NewLocalOrder(c types.Config) *localOrderImpl {
	return newLocalOrderImpl(c)
}

func (local *localOrderImpl) Start() {
	local.start()
}

func (local *localOrderImpl) Stop() {
	local.stop()
}
