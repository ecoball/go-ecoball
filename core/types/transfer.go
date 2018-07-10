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

package types

import (
	"errors"
	"fmt"
	"math/big"
	"github.com/ecoball/go-ecoball/common"
)

type TransferInfo struct {
	Value *big.Int
}

func NewTransferInfo(v *big.Int) *TransferInfo {
	t := new(TransferInfo)
	t.Value = new(big.Int).Set(v)
	return t
}

func NewTransfer(from, to common.AccountName, value *big.Int, nonce uint64, time int64) (*Transaction, error) {
	payload := NewTransferInfo(value)
	return NewTransaction(TxTransfer, from, to, payload, nonce, time)
}

func (t *TransferInfo) Serialize() ([]byte, error) {
	data, err := t.Value.GobEncode()
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (t *TransferInfo) Deserialize(data []byte) error {
	if len(data) == 0 {
		return errors.New("data len is 0")
	}
	t.Value = new(big.Int)
	return t.Value.GobDecode(data)
}

func (t TransferInfo) GetObject() interface{} {
	return t
}

func (t *TransferInfo) Show() {
	fmt.Println("\tValue          :", t.Value)
}
