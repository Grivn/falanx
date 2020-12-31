package falanx

import (
	cProto "github.com/Grivn/libfalanx/zcommon/protos"
	"github.com/Grivn/libfalanx/falanx/types"
	"github.com/Grivn/libfalanx/atxcontainer"

	fCommonProto "github.com/ultramesh/flato-common/types/protos"
)

type Falanx interface {
	ProcessTransaction(tx *fCommonProto.Transaction)
	ProcessOrderedRequest(or *cProto.OrderedReq)
	ProcessOrderedLog(ol *cProto.OrderedLog)
}

type falanxImpl struct {
	author uint64

	txContainer atxcontainer.TxsContainer
}

func NewFalanx(config types.Config) *falanxImpl {
	return &falanxImpl{
		author: config.Author,
		txContainer: config.TxContainer,
	}
}

func (f *falanxImpl) ProcessTransaction(tx *fCommonProto.Transaction) {
	f.txContainer.Add(tx)
}

func (f *falanxImpl) ProcessOrderedRequest(or *cProto.OrderedReq) {

}

func (f *falanxImpl) ProcessOrderedLog(ol *cProto.OrderedLog) {

}
