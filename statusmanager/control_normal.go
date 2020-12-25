package statusmanager

// On sets the status of specified positions.
func (sp *StatusProcessor) On(statusPos ...uint32) {
	for _, pos := range statusPos {
		sp.setBit(pos)
	}
}

// Off resets the status of specified positions.
func (sp *StatusProcessor) Off(statusPos ...uint32) {
	for _, pos := range statusPos {
		sp.clearBit(pos)
	}
}

// In returns the status of specified position.
func (sp *StatusProcessor) In(pos uint32) bool {
	return sp.hasBit(pos)
}

// InOne checks the result of several status computed with each other using '||'
func (sp *StatusProcessor) InOne(poss ...uint32) bool {
	var rs = false
	for _, pos := range poss {
		rs = rs || sp.hasBit(pos)
	}
	return rs
}

// ==================================================
// Normal Options to Process Normal Status
// ==================================================
// setBit sets the bit at position in integer n.
func (sp *StatusProcessor) setBit(position uint32) {
	sp.normalStatus |= 1 << position
}

// clearBit clears the bit at position in integer n.
func (sp *StatusProcessor) clearBit(position uint32) {
	sp.normalStatus &= ^(1 << position)
}

// hasBit checks whether a bit position is set.
func (sp *StatusProcessor) hasBit(position uint32) bool {
	val := sp.normalStatus & (1 << position)
	return val > 0
}
