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
	"github.com/ecoball/go-ecoball/common/config"
	"github.com/ecoball/go-ecoball/core/bloom"
	"github.com/ecoball/go-ecoball/core/ledgerimpl/ledger"
	"github.com/ecoball/go-ecoball/core/types"
	"time"
	"github.com/ecoball/go-ecoball/core/state"
	"fmt"
	"encoding/json"
	"github.com/ecoball/go-ecoball/smartcontract/wasmservice"
)

func GenesisBlockInit(ledger ledger.Ledger) (*types.Block, error) {
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
}

func PresetContract(ledger ledger.Ledger, t int64) ([]*types.Transaction, error) {
	var txs []*types.Transaction
	if ledger == nil {
		return nil, errors.New("ledger is nil")
	}
	index := common.NameToIndex("root")
	addr := common.AddressFromPubKey(common.FromHex(config.RootPubkey))
	fmt.Println("preset insert a root account:", addr.HexString())
	if acc, err := ledger.AccountAdd(index, addr); err != nil {
		return nil, err
	} else {
		acc.Show()
	}

	//TODO
	if err := ledger.AccountAddBalance(index, state.AbaToken, 10000); err != nil {
		return nil, err
	}
	code, err := wasmservice.ReadWasm("../../test/root/root.wasm")
	if err != nil {
		return nil, err
	}
	tokenContract, err := types.NewDeployContract(index, index, state.Active, types.VmWasm, "system control", code, 0, t)
	if err != nil {
		return nil, err
	}
	if err := tokenContract.SetSignature(&config.Root); err != nil {
		return nil, err
	}
	txs = append(txs, tokenContract)

	invoke, err := types.NewInvokeContract(index, index, state.Owner, types.VmWasm, "new_account",
		[]string{"worker1", common.AddressFromPubKey(config.Worker1.PublicKey).HexString()}, 0, t)
	invoke.SetSignature(&config.Root)
	txs = append(txs, invoke)

	invoke, err = types.NewInvokeContract(index, index, state.Owner, types.VmWasm, "new_account",
		[]string{"worker2", common.AddressFromPubKey(config.Worker2.PublicKey).HexString()}, 1, t)
	invoke.SetSignature(&config.Root)
	txs = append(txs, invoke)

	invoke, err = types.NewInvokeContract(index, index, state.Owner, types.VmWasm, "new_account",
		[]string{"worker3", common.AddressFromPubKey(config.Worker3.PublicKey).HexString()}, 2, t)
	invoke.SetSignature(&config.Root)
	txs = append(txs, invoke)

	perm := state.NewPermission(state.Active, state.Owner, 2, []state.KeyFactor{}, []state.AccFactor{{Actor: common.NameToIndex("worker1"), Weight: 1, Permission: "active"}, {Actor: common.NameToIndex("worker2"), Weight: 1, Permission: "active"}, {Actor: common.NameToIndex("worker3"), Weight: 1, Permission: "active"}})
	param, err := json.Marshal(perm)
	if err != nil {
		return nil, err
	}
	invoke, err = types.NewInvokeContract(index, index, state.Active, types.VmWasm, "set_account", []string{"root", string(param)}, 0, time.Now().Unix())
	invoke.SetSignature(&config.Root)
	txs = append(txs, invoke)
	//END

	return txs, nil
}
