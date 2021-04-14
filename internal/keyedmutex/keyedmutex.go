package keyedmutex

import (
	"github.com/cespare/xxhash"
	"sync"
)

type KeyedMutex []sync.Mutex

func (k KeyedMutex) locate(key []byte) int {
	return int(xxhash.Sum64(key) % uint64(len(k)))
}

func (k KeyedMutex) Lock(key []byte) {
	k[k.locate(key)].Lock()
}

func (k KeyedMutex) Unlock(key []byte) {
	k[k.locate(key)].Unlock()
}
