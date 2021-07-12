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
	GetCandidates() []int
}

func NewTxRecorder(replicas []int, hash string, n int, f int) *txRecorderImpl {
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

func (tr *txRecorderImpl) GetCandidates() []int {
	return tr.getCandidates()
}

type txRecorderImpl struct {
	n          int
	f          int
	hash       string
	ordered    map[uint64]bool
	pending    map[uint64]bool
	whitelist  []int
	candidates []int
	rand       *rand.Rand
}

func newTxRecorderImpl(replicas []int, hash string, n int, f int) *txRecorderImpl {
	sort.Ints(replicas)
	candidates := replicas[:n-f]
	pendingMap := make(map[uint64]bool)
	for _, i := range candidates {
		id := uint64(i)
		pendingMap[id] = true
	}

	// generate a seed for pseudo random function
	hex := StringToHex(hash)
	seed := int64(0)
	for _, val := range hex {
		s := int64(val)
		seed += s
	}

	return &txRecorderImpl{
		n:          n,
		f:          f,
		hash:       hash,
		ordered:    make(map[uint64]bool),
		pending:    pendingMap,
		whitelist:  replicas,
		candidates: candidates,
		rand:       rand.New(rand.NewSource(seed)),
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
	candidates := whitelist[:tr.n-tr.f]
	tr.pending = make(map[uint64]bool)
	for _, i := range candidates {
		id := uint64(i)
		tr.pending[id] = true
		if tr.ordered[id] {
			tr.pending[id] = false
		}
	}
	tr.candidates = candidates
}

func (tr *txRecorderImpl) pendingLen() int {
	return len(tr.pending)
}

func (tr *txRecorderImpl) orderLen() int {
	return len(tr.ordered)
}

func (tr *txRecorderImpl) getMalicious() []uint64 {
	var malicious []uint64
	for id, pending := range tr.pending {
		if pending {
			malicious = append(malicious, id)
		}
	}
	return malicious
}

func (tr *txRecorderImpl) getCandidates() []int {
	return tr.candidates
}
