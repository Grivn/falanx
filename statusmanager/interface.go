package statusmanager

type StatusManager interface {
	NormalStatus
	AtomicStatus
}

type NormalStatus interface {
	On(statusPos ...uint32)
	Off(statusPos ...uint32)
	In(pos uint32) bool
	InOne(poss ...uint32) bool
}

type AtomicStatus interface {
	AOn(statusPos ...uint32)
	AOff(statusPos ...uint32)
	AIn(pos uint32) bool
	AInOne(poss ...uint32) bool
}
