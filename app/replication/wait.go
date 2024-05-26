package replication

import (
	"fmt"
	"sync"
)

func (repl *Replicator) Blocked() bool {
	return repl.followerCounter.Locked()
}

type followerCounter struct {
	mutex       sync.Mutex
	locked      bool
	responded   map[int]bool
	portChannel chan int
}

func (lock *followerCounter) Locked() bool {
	return lock.locked
}

func (lock *followerCounter) StartBlock() error {
	if lock.locked {
		return fmt.Errorf("followerCounter isn't locked")
	}
	lock.mutex.Lock()
	defer lock.mutex.Unlock()
	lock.locked = true
	lock.responded = make(map[int]bool)
	lock.portChannel = make(chan int)
	return nil
}

func (lock *followerCounter) EndBlock() error {
	if !lock.locked {
		return fmt.Errorf("followerCounter isn't locked")
	}
	lock.mutex.Lock()
	defer lock.mutex.Unlock()
	lock.responded = nil
	close(lock.portChannel)
	lock.locked = false
	return nil
}

func (lock *followerCounter) AddRespondedFollower(port int) error {
	if !lock.locked {
		return fmt.Errorf("followerCounter isn't locked")
	}
	lock.mutex.Lock()
	defer lock.mutex.Unlock()
	lock.responded[port] = true
	lock.portChannel <- port
	return nil
}

func (lock *followerCounter) NumRespondedFollowers() int {
	lock.mutex.Lock()
	defer lock.mutex.Unlock()
	return len(lock.responded)
}
