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

package store_test

import (
	"bytes"
	"fmt"
	"github.com/ecoball/go-ecoball/core/store"
	"testing"
)

func TestStore(t *testing.T) {
	s, err := store.NewBlockStore("/tmp/test")
	if err != nil {
		t.Fatal(err)
	}
	r, _ := s.SearchAll()
	fmt.Println(len(r))
	//存储
	s.Put([]byte("key1"), []byte("value1"))
	s.Put([]byte("key2"), []byte("value2"))
	s.Put([]byte("key3"), []byte("value3"))
	//读取
	if v, err := s.Get([]byte("key1")); err != nil {
		if !bytes.Equal(v, []byte("value1")) {
			t.Fatal("value1 error")
		}
	}
	if v, err := s.Get([]byte("key3")); err != nil {
		if !bytes.Equal(v, []byte("value3")) {
			t.Fatal("value1 error")
		}
	}
	//批处理存储
	s.BatchPut([]byte("key01"), []byte("value01"))
	s.BatchPut([]byte("key02"), []byte("value02"))
	s.BatchPut([]byte("key03"), []byte("value03"))
	s.BatchPut([]byte("key04"), []byte("value04"))
	s.BatchCommit()
	if v, err := s.Get([]byte("key03")); err != nil {
		if !bytes.Equal(v, []byte("value03")) {
			t.Fatal("value1 error")
		}
	}
	//迭代器遍历
	it := s.NewIterator()
	for it.Next() {
		fmt.Println(string(it.Key()), string(it.Value()))
	}
	it.Release()
	if err := it.Error(); err != nil {
		t.Fatal(err)
	}

	//函数遍历
	re, err := s.SearchAll()
	if err != nil {
		t.Fatal(err)
	}
	for k, v := range re {
		fmt.Println(k, v)
	}

}
