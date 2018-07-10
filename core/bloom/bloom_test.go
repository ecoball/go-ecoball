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

package bloom_test

import (
	"github.com/ecoball/go-ecoball/core/bloom"
	"math/big"
	"testing"
)

func TestNewBloom(t *testing.T) {
	b := bloom.NewBloom(nil)
	b.Add([]byte("pct"))
	if b.Test([]byte("pct")) != true {
		t.Fatal("error test pct")
	}
	if b.Test([]byte("pct2")) == true {
		t.Fatal("error test pct2")
	}

	data := b.Bytes()
	b2 := bloom.NewBloom(data)
	if b2.Test([]byte("pct")) != true {
		t.Fatal("error test pct")
	}
	if b2.Test([]byte("pct2")) == true {
		t.Fatal("error test pct2")
	}
}

func TestBloomCycle(t *testing.T) {
	b := bloom.NewBloom(nil)
	for i := 0; i < 100000; i++ {
		key := new(big.Int).SetInt64(int64(i))
		b.Add(key.Bytes())
	}
	for i := 0; i < 100000; i++ {
		key := new(big.Int).SetInt64(int64(i))
		if true != b.Test(key.Bytes()) {
			t.Fatal("test error", i)
		}
	}
}
