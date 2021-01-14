package replicasorder

import "github.com/Grivn/libfalanx/replicasorder/types"

func NewReplicaOrder(c types.Config) *replicaOrderImpl {
	return newReplicaOrderImpl(c)
}

func (r *replicaOrderImpl) Start() {
	r.start()
}

func (r *replicaOrderImpl) Stop() {
	r.stop()
}
