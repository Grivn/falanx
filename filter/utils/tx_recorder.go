package utils

import (
	"math/rand"
	"sort"
)

type TxRecorder interface {
	Add(id uint64)
	Update(whitelist []int)
	PendingLen() int
	OrderLen() int
	GetMalicious() []uint64
}

func NewTxRecorder(replicas []int, hash string, n uint64, f uint64) *txRecorderImpl {
	return newTxRecorderImpl(replicas, hash, n, f)
}

func (tr *txRecorderImpl) Add(id uint64) {
	tr.add(id)
}

func (tr *txRecorderImpl) Update(whitelist []int) {
	tr.update(whitelist)
}

func (tr *txRecorderImpl) PendingLen() int {
	return tr.pendingLen()
}

func (tr *txRecorderImpl) OrderLen() int {
	return tr.orderLen()
}

func (tr *txRecorderImpl) GetMalicious() []uint64 {
	return tr.getMalicious()
}

type txRecorderImpl struct {
	n          uint64
	f          uint64
	hash       string
	ordered    map[uint64]bool
	candidates map[uint64]bool
	whitelist  []int
	rand       rand.Rand
}

func newTxRecorderImpl(replicas []int, hash string, n uint64, f uint64) *txRecorderImpl {
	sort.Ints(replicas)
	whitelist := replicas
	candidates := whitelist[:int(n-f)]
	candidatesMap := make(map[uint64]bool)
	for _, i := range candidates {
		id := uint64(i)
		candidatesMap[id] = true
	}

	return &txRecorderImpl{
		n:          n,
		f:          f,
		hash:       hash,
		ordered:    make(map[uint64]bool),
		whitelist:  replicas,
		candidates: candidatesMap,
	}
}

func (tr *txRecorderImpl) add(id uint64) {
	_, ok := tr.ordered[id]
	if ok {
		return
	}
	tr.ordered[id] = true
}

func (tr *txRecorderImpl) update(whitelist []int) {
	tr.whitelist = whitelist
	candidates := whitelist[:int(tr.n-tr.f)]
	tr.candidates = make(map[uint64]bool)
	for _, i := range candidates {
		id := uint64(i)
		tr.candidates[id] = true
		if tr.ordered[id] {
			tr.candidates[id] = false
		}
	}
}

func (tr *txRecorderImpl) pendingLen() int {
	return len(tr.candidates)
}

func (tr *txRecorderImpl) orderLen() int {
	return len(tr.ordered)
}

func (tr *txRecorderImpl) getMalicious() []uint64 {
	var malicious []uint64
	for id, pending := range tr.candidates {
		if pending {
			malicious = append(malicious, id)
		}
	}
	return malicious
}
