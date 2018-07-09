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

package geneses

import (
	"errors"
	"github.com/ecoball/go-ecoball/common"
	"github.com/ecoball/go-ecoball/core/bloom"
	"github.com/ecoball/go-ecoball/core/ledgerimpl/ledger"
	"github.com/ecoball/go-ecoball/core/types"
	"time"
)

func GenesesBlockInit(ledger ledger.Ledger) (*types.Block, error) {
	if ledger == nil {
		return nil, errors.New("ledger is nil")
	}
	tm, err := time.Parse("02/01/2006 15:04:05 PM", "21/02/1990 00:00:00 AM")
	if err != nil {
		return nil, err
	}
	timeStamp := tm.Unix()

	//TODO start
	SecondInMs := int64(1000)
	BlockIntervalInMs := int64(15000)
	timeStamp = int64((timeStamp*SecondInMs-SecondInMs)/BlockIntervalInMs) * BlockIntervalInMs
	timeStamp = timeStamp / SecondInMs
	//TODO end

	hash := common.NewHash([]byte("EcoBall Geneses Block"))
	conData := types.GenesesBlockInitConsensusData(timeStamp)
	txs, err := GenesesContract(ledger, timeStamp)
	if err != nil {
		return nil, err
	}
	hashState := ledger.StateDB().GetHashRoot()
	header, err := types.NewHeader(types.VersionHeader, 1, hash, hash, hashState, *conData, bloom.Bloom{}, timeStamp)
	if err != nil {
		return nil, err
	}
	block := types.Block{Header: header, CountTxs: uint32(len(txs)), Transactions: txs}
	return &block, nil
}

func GenesesContract(ledger ledger.Ledger, t int64) ([]*types.Transaction, error) {
	var txs []*types.Transaction
	if ledger == nil {
		return nil, errors.New("ledger is nil")
	}
	index := common.NameToIndex("token")
	if err := ledger.AccountAdd(index, common.NewAddress([]byte("system token"))); err != nil {
		return nil, err
	}
	tokenContract, err := types.NewDeployContract(0, index, types.VmNative, "system token", nil, 0, t)
	if err != nil {
		return nil, err
	}
	txs = append(txs, tokenContract)

	index = common.NameToIndex("account")
	if err := ledger.AccountAdd(index, common.NewAddress([]byte("system account"))); err != nil {
		return nil, err
	}
	accountContract, err := types.NewDeployContract(0, index, types.VmNative, "system account", nil, 0, t)
	if err != nil {
		return nil, err
	}
	txs = append(txs, accountContract)
	return txs, nil
}
