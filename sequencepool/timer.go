package sequencepool

import (
	"strconv"
	"sync"
	"time"

	"github.com/Grivn/libfalanx/logger"
	"github.com/Grivn/libfalanx/sequencepool/types"
)

// titleTimer
type titleTimer struct {
	timerName string        // the unique timer name
	timeout   time.Duration // default timeout of this timer
	isActive  sync.Map      // track all the timers with this timerName if it is active now
}

func (tt *titleTimer) store(key, value interface{}) {
	tt.isActive.Store(key, value)
}

func (tt *titleTimer) delete(key interface{}) {
	tt.isActive.Delete(key)
}

func (tt *titleTimer) has(key string) bool {
	_, ok := tt.isActive.Load(key)
	return ok
}

func (tt *titleTimer) count() int {
	length := 0
	tt.isActive.Range(func(_, _ interface{}) bool {
		length++
		return true
	})
	return length
}

func (tt *titleTimer) clear() {
	tt.isActive.Range(func(key, _ interface{}) bool {
		tt.isActive.Delete(key)
		return true
	})
}

// timerManager manages common used timers.
type timerManager struct {
	tTimers map[string]*titleTimer
	eventC  chan interface{}
	logger  logger.Logger
}

// NewTimerMgr news an instance of timerManager.
func NewTimerMgr(c types.Config) *timerManager {
	tm := &timerManager{
		tTimers: make(map[string]*titleTimer),
		eventC:  c.TimeoutC,
		logger:  c.Logger,
	}

	return tm
}

// newTimer news a titleTimer with the given name and default timeout, then add this timer to timerManager
func (tm *timerManager) newTimer(name string, d time.Duration) {
	tm.tTimers[name] = &titleTimer{
		timerName: name,
		timeout:   d,
	}
}

// Stop stops all timers managed by timerManager
func (tm *timerManager) Stop() {
	for timerName := range tm.tTimers {
		tm.stopTimer(timerName)
	}
}

// startTimer starts the timer for particular transaction
func (tm *timerManager) startTimer(name string, event types.TimeoutEvent) string {
	tm.stopTimer(name)

	timestamp := time.Now().UnixNano()
	key := strconv.FormatInt(timestamp, 10)
	tm.tTimers[name].store(key, true)

	send := func() {
		if tm.tTimers[name].has(key) {
			tm.eventC <- event
		}
	}
	time.AfterFunc(tm.tTimers[name].timeout, send)
	return key
}

// stopTimer stops all timers with the same timerName.
func (tm *timerManager) stopTimer(name string) {
	if !tm.containsTimer(name) {
		tm.logger.Errorf("Stop timer failed, timer %s not created yet!", name)
		return
	}

	tm.tTimers[name].clear()
}

// containsTimer returns true if there exists a timer named timerName
func (tm *timerManager) containsTimer(timerName string) bool {
	_, ok := tm.tTimers[timerName]
	return ok
}
