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
	"github.com/ecoball/go-ecoball/common"
	"math/big"
	"encoding/json"
)

type TransferInfo struct {
	Value *big.Int `json:"value"`
}

func NewTransferInfo(v *big.Int) *TransferInfo {
	t := new(TransferInfo)
	t.Value = new(big.Int).Set(v)
	return t
}

func NewTransfer(from, to common.AccountName, perm string, value *big.Int, nonce uint64, time int64) (*Transaction, error) {
	payload := NewTransferInfo(value)
	return NewTransaction(TxTransfer, from, to, perm, payload, nonce, time)
}

/**
 *  @brief converts a structure into a sequence of characters
 *  @return []byte - a sequence of characters
 */
func (t *TransferInfo) Serialize() ([]byte, error) {
	data, err := t.Value.GobEncode()
	if err != nil {
		return nil, err
	}
	return data, nil
}

/**
 *  @brief converts a sequence of characters into a structure
 *  @param data - a sequence of characters
 */
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

func (t *TransferInfo) JsonString() string {
	data, _ := json.Marshal(t)
	return string(data)
}