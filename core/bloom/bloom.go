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

package bloom

import (
	"fmt"
	"github.com/ecoball/go-ecoball/common"
	"math/big"
)

const (
	BloomByteLength = 1024
	BloomHashCycle  = 2
	BloomBitLength  = 8 * BloomByteLength
)

type Bloom [BloomByteLength]byte

func NewBloom(b []byte) Bloom {
	var bloom Bloom
	bloom.SetBytes(b)
	return bloom
}
func (b *Bloom) SetBytes(d []byte) {
	if len(b) < len(d) {
		panic(fmt.Sprintf("bloom bytes too big %d %d", len(b), len(d)))
	}
	copy(b[BloomByteLength-len(d):], d)
}

func (b Bloom) Bytes() []byte {
	return b[:]
}

func (b *Bloom) Add(data []byte) {
	b.add(new(big.Int).SetBytes([]byte(data)))
}

func (b *Bloom) add(d *big.Int) {
	bin := new(big.Int).SetBytes(b[:])
	bin.Or(bin, bloom(d.Bytes()))
	b.SetBytes(bin.Bytes())
}

func (b Bloom) Test(test []byte) bool {
	return b.test(new(big.Int).SetBytes(test))
}

func (b Bloom) test(test *big.Int) bool {
	return bloomLookup(b, test.Bytes())
}

func (b Bloom) Big() *big.Int {
	return new(big.Int).SetBytes(b[:])
}

func bloom(b []byte) *big.Int {
	b = common.Keccak256Hash(b[:]).Bytes()
	r := new(big.Int)
	for i := 0; i < BloomHashCycle; i += 2 {
		t := big.NewInt(1)
		b := (uint(b[i+1]) + (uint(b[i]) << 8)) & 2047
		r.Or(r, t.Lsh(t, b))
	}
	return r
}
func bloomLookup(bin Bloom, key []byte) bool {
	b := bin.Big()
	cmp := bloom(key)
	return b.And(b, cmp).Cmp(cmp) == 0
}
