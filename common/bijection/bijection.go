package bijection

import (
	"errors"
	"sync"
	"fmt"
)

type Bijection struct {
	data map[uint64] uint64
	backend map[uint64] uint64
	mutex sync.RWMutex
}


func New() Bijection{
	b := Bijection{
		data: make(map[uint64]uint64),
		backend: make(map[uint64]uint64)}
	return b
}

func (b *Bijection) Set(key, value uint64) error {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	if _, ok := b.data[key]; ok {
		return errors.New(fmt.Sprintf("key is existed: %d", key))
	}
	if _, ok := b.backend[value]; ok {
		return errors.New(fmt.Sprintf("value is existed: %d", value))
	}
	b.data[key] = value
	b.backend[value] = key
	return nil
}

func (b *Bijection) Get(key uint64) (uint64, error) {
	b.mutex.RLock()
	defer b.mutex.RUnlock()
	v, ok := b.data[key]
	if !ok {
		return 0, errors.New("can't find this value")
	}
	return v, nil
}