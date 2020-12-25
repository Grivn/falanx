package statusmanager

import "sync/atomic"

// atomic on sets the atomic status of specified positions.
func (st *statusManager) AOn(statusPos ...uint32) {
	for _, pos := range statusPos {
		st.atomicSetBit(pos)
	}
}

// atomic off resets the atomic status of specified positions.
func (st *statusManager) AOff(statusPos ...uint32) {
	for _, pos := range statusPos {
		st.atomicClearBit(pos)
	}
}

// atomic in returns the atomic status of specified position.
func (st *statusManager) AIn(pos uint32) bool {
	return st.atomicHasBit(pos)
}

// atomic inOne checks the result of several atomic status computed with each other using '||'
func (st *statusManager) AInOne(poss ...uint32) bool {
	var rs = false
	for _, pos := range poss {
		rs = rs || st.atomicHasBit(pos)
	}
	return rs
}

// ==================================================
// Atomic Options to Process Atomic Status
// ==================================================
// atomic setBit sets the bit at position in integer n.
func (st *statusManager) atomicSetBit(position uint32) {
	// try CompareAndSwapUint64 until success
	for {
		oldStatus := atomic.LoadUint32(&st.atomicStatus)
		if atomic.CompareAndSwapUint32(&st.atomicStatus, oldStatus, oldStatus|(1<<position)) {
			break
		}
	}
}

// atomic clearBit clears the bit at position in integer n.
func (st *statusManager) atomicClearBit(position uint32) {
	// try CompareAndSwapUint64 until success
	for {
		oldStatus := atomic.LoadUint32(&st.atomicStatus)
		if atomic.CompareAndSwapUint32(&st.atomicStatus, oldStatus, oldStatus&^(1<<position)) {
			break
		}
	}
}

// atomic hasBit checks atomic whether a bit position is set.
func (st *statusManager) atomicHasBit(position uint32) bool {
	val := atomic.LoadUint32(&st.atomicStatus) & (1 << position)
	return val > 0
}
