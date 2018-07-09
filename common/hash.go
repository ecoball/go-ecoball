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

package common

import (
	"bytes"
	"errors"
	"github.com/ecoball/go-ecoball/crypto/sha3"
	"math/big"
)

const HashLen = 32

type Hash [HashLen]byte

func NewHash(addr []byte) Hash {
	var hash Hash
	copy(hash[:], addr)
	return hash
}

func BytesToHash(b []byte) Hash {
	var h Hash
	h.SetBytes(b)
	return h
}

func BigToHash(b *big.Int) Hash { return BytesToHash(b.Bytes()) }

func HexToHash(s string) Hash   { return BytesToHash(FromHex(s)) }

func (h Hash) Bytes() []byte {
	return h[:]
}

// Sets the hash to the value of b. If b is larger than len(h), 'b' will be cropped (from the left).
func (h *Hash) SetBytes(b []byte) {
	if len(b) > len(h) {
		b = b[len(b)-HashLen:]
	}

	copy(h[HashLen-len(b):], b)
}

func SingleHash(b []byte) Hash {
	return Keccak256Hash(b)
}

func DoubleHash(b []byte) (Hash, error) {
	if len(b) == 0 {
		return Hash{}, errors.New("len of data is zero")
	}
	temp := Keccak256Hash(b)
	f := Keccak256Hash(temp[:])
	return Hash(f), nil
}

func Keccak256Hash(data ...[]byte) (hash Hash) {
	d := sha3.NewKeccak256()
	for _, b := range data {
		d.Write(b)
	}
	d.Sum(hash[:0])
	return hash
}

func (h Hash) HexString() string {
	return ToHex(h[:])
}

func (h* Hash)FormHexString(data string) Hash{
	hash := NewHash(FromHex(data))
	return hash
}

func (h *Hash) Equals(b *Hash) bool {
	if nil == h {
		return nil == b
	}
	if nil == b {
		return false
	}
	return bytes.Equal(h[:], b[:])
}

func (h *Hash) IsNil() bool {
	return h.Equals(&Hash{})
}