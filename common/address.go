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

	"github.com/ecoball/go-ecoball/core/pb"
	"github.com/gogo/protobuf/proto"

	"github.com/btcsuite/btcutil/base58"
)

const AddressLen = 20

type Address [AddressLen]byte

func NewAddress(addr []byte) Address {
	var address Address
	copy(address[:], addr)
	return address
}

func AddressFromBase58(data string) Address {
	return NewAddress(base58.Decode(data))
}
func (a Address) Bytes() []byte {
	return a[:]
}

/* Return address string. */
func (a Address) ToBase58() string {
	return base58.Encode(a[:])
}

// ToHexString returns  hex string representation of Address
func (a Address) HexString() string {
	//return fmt.Sprintf("%x", a[:])
	return ToHex(a[:])
}

func FormHexString(data string) Address{
	return NewAddress(FromHex(data))
}

/* Equals compare two Address. True is equal, otherwise false. */
func (a *Address) Equals(b *Address) bool {
	if nil == a {
		return nil == b
	}
	if nil == b {
		return false
	}
	return bytes.Equal(a[:], b[:])
}

type Signature struct {
	PubKey  []byte //签名者的公钥
	SigData []byte //签名数据
}

func (s *Signature) Serialize() ([]byte, error) {
	sig := pb.Signature{
		SigData: s.SigData,
		PubKey:  s.PubKey,
	}
	d, err := proto.Marshal(&sig)
	if err != nil {
		return nil, err
	}
	return d, nil
}

func (s *Signature) Deserialize(data []byte) error {
	if len(data) == 0 {
		return errors.New("input data's length is zero")
	}
	var sig pb.Signature
	if err := proto.Unmarshal(data, &sig); err != nil {
		return err
	}
	s.PubKey = sig.PubKey
	s.SigData = sig.SigData
	return nil
}
