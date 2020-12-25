package statusmanager

import "sync/atomic"

// consensus status
type StatusProcessor struct {
	normalStatus uint32
	atomicStatus uint32
}

func NewStatusProcessor() *StatusProcessor {
	sp := &StatusProcessor{}
	sp.reset()
	return sp
}

// reset only resets consensus status to 0.
func (sp *StatusProcessor) reset() {
	sp.normalStatus = 0
	atomic.StoreUint32(&sp.atomicStatus, 0)
}
