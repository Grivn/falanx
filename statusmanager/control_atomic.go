package statusmanager

import "sync/atomic"

// atomic on sets the atomic status of specified positions.
func (sp *StatusProcessor) AOn(statusPos ...uint32) {
	for _, pos := range statusPos {
		sp.atomicSetBit(pos)
	}
}

// atomic off resets the atomic status of specified positions.
func (sp *StatusProcessor) AOff(statusPos ...uint32) {
	for _, pos := range statusPos {
		sp.atomicClearBit(pos)
	}
}

// atomic in returns the atomic status of specified position.
func (sp *StatusProcessor) AIn(pos uint32) bool {
	return sp.atomicHasBit(pos)
}

// atomic inOne checks the result of several atomic status computed with each other using '||'
func (sp *StatusProcessor) AInOne(poss ...uint32) bool {
	var rs = false
	for _, pos := range poss {
		rs = rs || sp.atomicHasBit(pos)
	}
	return rs
}

// ==================================================
// Atomic Options to Process Atomic Status
// ==================================================
// atomic setBit sets the bit at position in integer n.
func (sp *StatusProcessor) atomicSetBit(position uint32) {
	// try CompareAndSwapUint64 until success
	for {
		oldStatus := atomic.LoadUint32(&sp.atomicStatus)
		if atomic.CompareAndSwapUint32(&sp.atomicStatus, oldStatus, oldStatus|(1<<position)) {
			break
		}
	}
}

// atomic clearBit clears the bit at position in integer n.
func (sp *StatusProcessor) atomicClearBit(position uint32) {
	// try CompareAndSwapUint64 until success
	for {
		oldStatus := atomic.LoadUint32(&sp.atomicStatus)
		if atomic.CompareAndSwapUint32(&sp.atomicStatus, oldStatus, oldStatus&^(1<<position)) {
			break
		}
	}
}

// atomic hasBit checks atomic whether a bit position is set.
func (sp *StatusProcessor) atomicHasBit(position uint32) bool {
	val := atomic.LoadUint32(&sp.atomicStatus) & (1 << position)
	return val > 0
}
