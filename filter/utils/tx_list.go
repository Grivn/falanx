package utils

import (
	"container/list"
	"errors"
	"github.com/Grivn/libfalanx/logger"
	pb "github.com/Grivn/libfalanx/zcommon/protos"
)

// Recorder List Interfaces ===========================================
type TxList interface {
	// list controller
	Add(l *pb.OrderedLog)
	Len() int
	Pop() *pb.OrderedLog

	// value controller
	GetFrontLog() *pb.OrderedLog
	GetSequence(key string) (uint64, error)
	GetByOrder(order int) *pb.OrderedLog
	GetHashList(max int) []string

	RemoveByHash(hash string)
}
// ====================================================================

func NewTxList(logger logger.Logger) *txListImpl {
	return newTxListImpl(logger)
}

func (tli *txListImpl) RemoveByHash(hash string) {
	tli.removeByHash(hash)
}

func (tli *txListImpl) GetHashList(max int) []string {
	return tli.getHashList(max)
}

func (tli *txListImpl) Pop() *pb.OrderedLog {
	e := tli.front()
	log := e.Value.(*pb.OrderedLog)
	delete(tli.presence, log.TxHash)
	tli.list.Remove(e)
	return log
}

func (tli *txListImpl) GetFrontLog() *pb.OrderedLog {
	return tli.frontLog()
}

func (tli *txListImpl) Add(l *pb.OrderedLog) {
	tli.add(l)
}

func (tli *txListImpl) GetSequence(key string) (uint64, error) {
	return tli.getSequence(key)
}

func (tli *txListImpl) Len() int {
	return tli.len()
}

func (tli *txListImpl) GetByOrder(order int) *pb.OrderedLog {
	return tli.getByOrder(order)
}

type txListImpl struct {
	// list contains the ordered logs from a replica
	list *list.List
	// presence indicates the elements in log list
	presence map[string]*list.Element

	logger logger.Logger
}

func newTxListImpl(logger logger.Logger) *txListImpl {
	return &txListImpl{
		list:     list.New(),
		presence: make(map[string]*list.Element),
		logger:   logger,
	}
}

func (tli *txListImpl) getByOrder(order int) *pb.OrderedLog {
	if tli.len() < order {
		return nil
	}

	i := 0
	element := tli.list.Front()
	for {
		if element == nil {
			panic("nil element!")
		}
		i++
		if i == order {
			break
		}
		element = element.Next()
	}

	log, ok := element.Value.(*pb.OrderedLog)
	if !ok {
		panic("parsing error")
	}

	return log
}

func (tli *txListImpl) getHashList(max int) []string {
	if tli.len() == 0 {
		return nil
	}

	var (
		next     *list.Element
		element  *list.Element
		hashList []string
	)

	i := 0
	for element = tli.list.Front(); i< max; element = next {
		if element == nil {
			panic("nil element!")
		}
		log, ok := element.Value.(*pb.OrderedLog)
		if !ok {
			panic("parsing error")
		}
		hashList = append(hashList, log.TxHash)

		i++
		next = element.Next()
	}

	return hashList
}

func (tli *txListImpl) frontLog() *pb.OrderedLog {
	e := tli.list.Front()
	log, ok := e.Value.(*pb.OrderedLog)
	if !ok {
		panic("parsing error!")
	}
	return log
}

func (tli *txListImpl) add(l *pb.OrderedLog) {
	if tli.has(l.TxHash) {
		return
	}
	e := tli.list.PushBack(l)
	tli.presence[l.TxHash] = e
}

func (tli *txListImpl) get(key string) *list.Element {
	e, ok := tli.presence[key]
	if !ok {
		return nil
	}
	return e
}

func (tli *txListImpl) getLog(key string) *pb.OrderedLog {
	e := tli.get(key)
	if e == nil {
		return nil
	}
	r, ok := e.Value.(*pb.OrderedLog)
	if !ok {
		return nil
	}
	return r
}

func (tli *txListImpl) getSequence(key string) (uint64, error) {
	e := tli.getLog(key)
	if e == nil {
		return 0, errors.New("nil element")
	}
	return e.Sequence, nil
}

func (tli *txListImpl) has(key string) bool {
	_, ok := tli.presence[key]
	return ok
}

func (tli *txListImpl) front() *list.Element {
	return tli.list.Front()
}

func (tli *txListImpl) pushBack(key string, value interface{}) *list.Element {
	if tli.has(key) {
		return nil
	}
	e := tli.list.PushBack(value)
	tli.presence[key] = e
	return e
}

func (tli *txListImpl) len() int {
	return tli.list.Len()
}

func (tli *txListImpl) removeByHash(hash string) {
	e := tli.presence[hash]
	if e == nil {
		return
	}
	v := tli.list.Remove(e)
	if v == nil {
		panic("nil value!")
	}
	log := v.(*pb.OrderedLog)
	tli.logger.Infof("[LIST] remove, (%d, %d)", log.ReplicaId, log.Sequence)
	delete(tli.presence, hash)
}
