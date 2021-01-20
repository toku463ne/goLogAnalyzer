package analyzer

import (
	"fmt"

	"github.com/juju/fslock"
)

// Lock .. lock object
type lock struct {
	lock *fslock.Lock
}

// NewLock .. gain lock to the lock file
func newLock(rootDir string) (*lock, error) {
	if rootDir == "" {
		return nil, nil
	}
	if !PathExist(rootDir) {
		return nil, nil
	}

	lock1 := new(lock)
	lock1.lock = fslock.New(fmt.Sprintf("%s/lock", rootDir))
	if err := lock1.lock.TryLock(); err != nil {
		return nil, err
	}
	return lock1, nil
}

// unLock .. unlock to the lock file
func (l *lock) unLock() error {
	if l.lock != nil {
		if err := l.lock.Unlock(); err != nil {
			//fmt.Printf("%+v\n", err)
			return err
		}
	}
	return nil
}
