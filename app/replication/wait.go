package replication

import (
	"sync"
	"time"
)

func (repl *Replicator) Blocked() bool {
	return repl.waitLock.Locked()
}

func (repl *Replicator) StartBlock(durationSeconds int) bool {
	return repl.waitLock.StartBlock(durationSeconds)
}

func (repl *Replicator) EndBlockEarly() bool {
	return repl.waitLock.EndBlock()
}

type waitLock struct {
	mutex           sync.Mutex
	locked          bool
	startTime       time.Time
	durationSeconds int
}

func (lock *waitLock) Locked() bool {
	return lock.locked
}

func (lock *waitLock) StartBlock(durationSeconds int) bool {
	if !lock.locked || durationSeconds <= 0 {
		return false
	}
	lock.mutex.Lock()
	lock.locked = true
	lock.durationSeconds = durationSeconds
	lock.startTime = time.Now()
	go func() {
		time.Sleep(time.Duration(durationSeconds))
		lock.EndBlock()
	}()
	return true
}

func (lock *waitLock) EndBlock() bool {
	if !lock.locked {
		return false
	}
	lock.mutex.Unlock()
	lock.locked = false
	return true
}
