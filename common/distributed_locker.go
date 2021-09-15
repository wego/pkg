package common

import (
	"hash/fnv"
	"sync"
)

// UnLocker unlock function
type UnLocker func()

// DistributedLocker a distributed lock type
type DistributedLocker interface {
	// Lock the specific key
	Lock(key string) (UnLocker, error)
}

// DMutex a simple distributed mutex for lock with keys
type DMutex struct {
	// hold mutex for each key
	// TODO: need a way to clean used ones
	mutexes sync.Map
}

// Lock the specific key
func (dm *DMutex) Lock(key string) (UnLocker, error) {
	value, _ := dm.mutexes.LoadOrStore(dm.hash(key), &sync.Mutex{})
	mtx := value.(*sync.Mutex)
	mtx.Lock()

	return func() { mtx.Unlock() }, nil
}

func (dm *DMutex) hash(value string) uint32 {
	h := fnv.New32a()
	_, _ = h.Write([]byte(value))
	return h.Sum32()
}
