package sequencepool

import (
	cTypes "github.com/Grivn/libfalanx/common/types"
	"github.com/Grivn/libfalanx/sequencepool/types"
)

type PoolProcessor struct {
	Author cTypes.Author

	SequenceLogs map[cTypes.Author]map[string]uint64

	//
	HashMap map[string]*cTypes.Transaction

	//

}

func (pp *PoolProcessor) ReceiveTransactions(or *types.OrderedRequest) {

}

func (pp *PoolProcessor) ReceiveReplicaLogs(ol *types.OrderedLog) {

}

func (pp *PoolProcessor) LocalLogOrder() {

}
