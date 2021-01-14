package filter

import (
	"math"
	"time"

	"github.com/Grivn/libfalanx/filter/types"
)

func (tf *transactionsFilterImpl) startPavingTimer(exitCh chan bool) {
	tf.paving = true
	go func() {
		timer := time.NewTimer(tf.delay)
		select {
		case <-timer.C:
			tf.pavingTimer <- true
		case <-exitCh:
			return
		}
	}()
}

func (tf *transactionsFilterImpl) stopPavingTimer() {
	close(tf.pavingExit)
	tf.pavingExit = make(chan bool)
	tf.paving = false
}

func (tf *transactionsFilterImpl) startGatheringTimer(exitCh chan bool) {
	tf.gathering = true
	go func() {
		timer := time.NewTimer(tf.delay)
		select {
		case <-timer.C:
			tf.gatheringTimer <- true
		case <-exitCh:
			return
		}
	}()
}

func (tf *transactionsFilterImpl) stopGatheringTimer() {
	close(tf.gatheringExit)
	tf.gatheringExit = make(chan bool)
	tf.gathering = false
}

func (tf *transactionsFilterImpl) startAppointingTimer(exitCh chan bool) {
	tf.appointing = true
	go func() {
		timer := time.NewTimer(tf.delay)
		select {
		case <-timer.C:
			tf.appointingTimer <- true
		case <-exitCh:
			return
		}
	}()
}

func (tf *transactionsFilterImpl) stopAppointingTimer() {
	close(tf.appointingExit)
	tf.appointingExit = make(chan bool)
	tf.appointing = false
}

func (tf *transactionsFilterImpl) getRelationCert(former, latter string) *types.RelationCert {
	idr := types.RelationId{From: former, To: latter}
	value, ok := tf.certStore[idr]

	if ok {
		return value
	}

	cert := &types.RelationCert{
		Finished:        false,
		Status:          types.NotEfficient,
		Scanned:         make(map[uint64]bool),
		FormerPreferred: 0,
		LatterPreferred: 0,
	}
	tf.certStore[idr] = cert
	return cert
}

func (tf *transactionsFilterImpl) allReplicas() int {
	return tf.n
}

func (tf *transactionsFilterImpl) allQuorumReplicas() int {
	return tf.n-tf.f
}

func (tf *transactionsFilterImpl) moreThanHalf() int {
	return int(math.Ceil(float64(tf.n)/float64(2)))
}
