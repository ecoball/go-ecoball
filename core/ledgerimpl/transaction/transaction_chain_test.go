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

package transaction_test

import (
	"fmt"
	"github.com/ecoball/go-ecoball/common"
	"github.com/ecoball/go-ecoball/core/ledgerimpl"
	"github.com/ecoball/go-ecoball/core/ledgerimpl/transaction"
	"github.com/ecoball/go-ecoball/core/state"
	"github.com/ecoball/go-ecoball/core/types"
	"github.com/ecoball/go-ecoball/smartcontract/wasmservice"
	"testing"
	"github.com/ecoball/go-ecoball/test/example"
)

func TestNewTransactionChain(t *testing.T) {
	txChain, err := transaction.NewTransactionChain("/tmp/Tx", nil)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(txChain.CurrentHeader.Hash.HexString())
	block, err := txChain.GetBlock(txChain.CurrentHeader.Hash)
	if err != nil {
		t.Fatal(err)
	}
	block.Show()
}

func TestBlockAdd(t *testing.T) {
	c, err := transaction.NewTransactionChain("/tmp/quaker/Tx", nil)
	if err != nil {
		t.Fatal(err)
	}

	re, err := c.BlockStore.SearchAll()
	if err != nil {
		t.Fatal(err)
	}
	for k, v := range re {
		hash := common.NewHash([]byte(k))
		block := new(types.Block)
		block.Deserialize([]byte(v))
		fmt.Println(hash.HexString())
		fmt.Println(block)
	}
}

func TestLedgerTxAdd(t *testing.T) {
	l, err := ledgerimpl.NewLedger("/tmp/quaker")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("Start LedgerImpl Module, hash:", l.GetCurrentHeader().Hash.HexString())
	example.ExampleAddAccount(l.StateDB())
	tx := example.ExampleTestTx()
	l.AccountAddBalance(tx.From, state.IndexAbaToken, 150)
	var txs []*types.Transaction
	txs = append(txs, tx)
	conData := types.ConsensusData{Type: types.ConSolo, Payload: &types.SoloData{}}
	block, err := l.NewTxBlock(txs, conData)
	if err != nil {
		t.Fatal(err)
	}
	if err := l.SaveTxBlock(block); err != nil {
		t.Fatal(err)
	}
	value, err := l.AccountGetBalance(tx.From, state.IndexAbaToken)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("value:", value)
	value, err = l.AccountGetBalance(tx.Addr, state.IndexAbaToken)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("value:", value)
}

func TestLedgerDeployAdd(t *testing.T) {
	l, err := ledgerimpl.NewLedger("/tmp/deploy")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("Start LedgerImpl Module, hash:", l.GetCurrentHeader().Hash.HexString())
	code, err := wasmservice.ReadWasm("../../../test/token.wasm")
	tx := example.ExampleTestDeploy(code)
	var txs []*types.Transaction
	txs = append(txs, tx)
	conData := types.ConsensusData{Type: types.ConSolo, Payload: &types.SoloData{}}
	block, err := l.NewTxBlock(txs, conData)
	if err != nil {
		t.Fatal(err)
	}
	if err := l.SaveTxBlock(block); err != nil {
		t.Fatal(err)
	}
	//Invoke Contract
	invoke := example.ExampleTestInvoke("create")
	var txs2 []*types.Transaction
	txs2 = append(txs, invoke)
	block, err = l.NewTxBlock(txs2, conData)
	if err != nil {
		t.Fatal(err)
	}
	if err := l.SaveTxBlock(block); err != nil {
		t.Fatal(err)
	}
}
/*
func TestLedgerInterface(t *testing.T) {
	l, err := ledgerimpl.NewLedger("/tmp/quaker")
	if err != nil {
		t.Fatal(err)
	}
	addr := common.NewAddress(common.FromHex("01b1a6569a557eafcccc71e0d02461fd4b601aea"))
	if err := l.AccountAddBalance(addr, "Test", 100001); err != nil {
		t.Fatal(err)
	}

	value, err := l.AccountGetBalance(addr, "Test")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("Token Abc Value:", value)

}

func TestLedgerInterfaceTokenCreate(t *testing.T) {
	l, err := ledgerimpl.NewLedger("/tmp/quaker")
	if err != nil {
		t.Fatal(err)
	}
	addr := common.NewAddress(common.FromHex("01b1a6569a557eafcccc71e0d02461fd4b601aea"))
	if err := l.TokenCreate(addr, "Test", 100001); err != nil {
		t.Fatal(err)
	}
	fmt.Println("Token Test Value:", l.TokenIsExisted("Test"))

}
*/