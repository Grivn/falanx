package filter

import "time"

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

func (tf *transactionsFilterImpl) allReplicas() int {
	return tf.n
}

func (tf *transactionsFilterImpl) allQuorumReplicas() int {
	return tf.n-tf.f
}
