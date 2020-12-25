package statusmanager

// On sets the status of specified positions.
func (st *statusManager) On(statusPos ...uint32) {
	for _, pos := range statusPos {
		st.setBit(pos)
	}
}

// Off resets the status of specified positions.
func (st *statusManager) Off(statusPos ...uint32) {
	for _, pos := range statusPos {
		st.clearBit(pos)
	}
}

// In returns the status of specified position.
func (st *statusManager) In(pos uint32) bool {
	return st.hasBit(pos)
}

// InOne checks the result of several status computed with each other using '||'
func (st *statusManager) InOne(poss ...uint32) bool {
	var rs = false
	for _, pos := range poss {
		rs = rs || st.hasBit(pos)
	}
	return rs
}

// ==================================================
// Normal Options to Process Normal Status
// ==================================================
// setBit sets the bit at position in integer n.
func (st *statusManager) setBit(position uint32) {
	st.normalStatus |= 1 << position
}

// clearBit clears the bit at position in integer n.
func (st *statusManager) clearBit(position uint32) {
	st.normalStatus &= ^(1 << position)
}

// hasBit checks whether a bit position is set.
func (st *statusManager) hasBit(position uint32) bool {
	val := st.normalStatus & (1 << position)
	return val > 0
}
