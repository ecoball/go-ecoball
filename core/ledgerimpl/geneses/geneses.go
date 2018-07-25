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
	"fmt"
	"github.com/ecoball/go-ecoball/common"
	"github.com/ecoball/go-ecoball/common/config"
	"github.com/ecoball/go-ecoball/core/types"
	//"github.com/ecoball/go-ecoball/core/bloom"
	"github.com/ecoball/go-ecoball/core/state"
	"math/big"
)

/*
func GenesisBlockInit(ledger ledger.Ledger, timeStamp int64) (*types.Block, error) {
	if ledger == nil {
		return nil, errors.New("ledger is nil")
	}

	//TODO start
	SecondInMs := int64(1000)
	BlockIntervalInMs := int64(15000)
	timeStamp = int64((timeStamp*SecondInMs-SecondInMs)/BlockIntervalInMs) * BlockIntervalInMs
	timeStamp = timeStamp / SecondInMs
	//TODO end

	hash := common.NewHash([]byte("EcoBall Geneses Block"))
	conData := types.GenesesBlockInitConsensusData(timeStamp)
	txs, err := PresetContract(ledger, timeStamp)
	if err != nil {
		return nil, err
	}


	hashState := ledger.StateDB().GetHashRoot()
	header, err := types.NewHeader(types.VersionHeader, 1, hash, hash, hashState, *conData, bloom.Bloom{}, timeStamp)
	if err != nil {
		return nil, err
	}
	block := types.Block{Header: header, CountTxs: uint32(len(txs)), Transactions: txs}

	if err := block.SetSignature(&config.Root); err != nil {
		return nil, err
	}
	return &block, nil
}*/

func PresetContract(s *state.State, t int64) ([]*types.Transaction, error) {
	var txs []*types.Transaction
	if s == nil {
		return nil, errors.New("state is nil")
	}
	root := common.NameToIndex("root")
	delegate := common.NameToIndex("delegate")
	addr := common.AddressFromPubKey(config.Root.PublicKey)
	fmt.Println("preset insert a root account:", addr.HexString())
	if _, err := s.AddAccount(root, addr); err != nil {
		return nil, err
	}
	if err := s.AccountAddBalance(root, state.AbaToken, new(big.Int).SetUint64(1000)); err != nil {
		return nil, err
	}
	fmt.Println("set root account's resource to [cpu:100, net:100]")
	if err := s.SetResourceLimits(root, root, 100, 100); err != nil {
		fmt.Println(err)
		return nil, err
	}

	if _, err := s.AddAccount(delegate, common.AddressFromPubKey(config.Delegate.PublicKey)); err != nil {
		return nil, err
	}
	if err := s.AccountAddBalance(delegate, state.AbaToken, new(big.Int).SetUint64(1000)); err != nil {
		return nil, err
	}
	fmt.Println("set root account's resource to [cpu:100, net:100]")
	if err := s.SetResourceLimits(delegate, delegate, 100, 100); err != nil {
		fmt.Println(err)
		return nil, err
	}


	return txs, nil
}
