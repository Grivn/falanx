package clientsorder

import (
	"github.com/Grivn/libfalanx/clientsorder/types"
)

func NewClientOrder(c types.Config) *clientOrderImpl {
	return newClientOrderImpl(c)
}
func (c *clientOrderImpl) Start() {
	c.start()
}
func (c *clientOrderImpl) Stop() {
	c.stop()
}