// Copyright 2018 The go-ecoball Authors
// This file is part of the go-ecoball library.
//
// The go-ecoball library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ecoball library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ecoball library. If not, see <http://www.gnu.org/licenses/>.
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