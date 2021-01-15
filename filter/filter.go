package filter

import "github.com/Grivn/libfalanx/filter/types"

func NewTransactionFilter(c types.Config) *transactionsFilterImpl {
	return newTransactionsFilterImpl(c)
}

func (tf *transactionsFilterImpl) Start() {
	tf.start()
}

func (tf *transactionsFilterImpl) Stop() {
	tf.stop()
}