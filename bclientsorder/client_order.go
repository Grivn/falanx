package bclientsorder

import (
	"github.com/Grivn/libfalanx/zcommon/protos"
)

// Interface ======================================================
type ClientOrder interface {
	ReceiveOrderedReq(r *protos.OrderedReq)
}
