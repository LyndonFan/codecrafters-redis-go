package replication

import (
	"sync"
)

func (repl *Replicator) Blocked() bool {
	return repl.waitLock.Locked()
}

func (repl *Replicator) startBlock() bool {
	return repl.waitLock.StartBlock()
}

func (repl *Replicator) endBlock() bool {
	return repl.waitLock.EndBlock()
}

type waitLock struct {
	mutex  sync.Mutex
	locked bool
}

func (lock *waitLock) Locked() bool {
	return lock.locked
}

func (lock *waitLock) StartBlock() bool {
	if lock.locked {
		return false
	}
	lock.mutex.Lock()
	lock.locked = true
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
