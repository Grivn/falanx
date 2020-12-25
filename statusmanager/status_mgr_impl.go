package statusmanager

import "sync/atomic"

// consensus status
type statusManager struct {
	normalStatus uint32
	atomicStatus uint32
}

func NewStatusManager() *statusManager{
	return newStatusMgr()
}

func newStatusMgr() *statusManager {
	st := &statusManager{}
	st.reset()
	return st
}

// reset only resets consensus status to 0.
func (st *statusManager) reset() {
	st.normalStatus = 0
	atomic.StoreUint32(&st.atomicStatus, 0)
}
