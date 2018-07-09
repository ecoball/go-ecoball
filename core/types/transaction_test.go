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

package types_test

import (
	"fmt"
	"github.com/ecoball/go-ecoball/core/types"
	"testing"
	"github.com/ecoball/go-ecoball/test/example"
)

func TestTransfer(t *testing.T) {
	fmt.Println("test create transaction")
	tx := example.ExampleTestTx()
	fmt.Println("Hash1:", tx.Hash.HexString())
	tx.Show()

	transferData, err := tx.Serialize()
	if err != nil {
		t.Fatal(err)
	}
	tx2 := &types.Transaction{}
	if err := tx2.Deserialize(transferData); err != nil {
		t.Fatal(err)
	}
	fmt.Println("Hash2:", tx2.Hash.HexString())
	if !tx2.Hash.Equals(&tx.Hash) {
		t.Fatal("hash wrong")
	}
	tx2.Show()
}

func TestDeploy(t *testing.T) {
	deploy := example.ExampleTestDeploy([]byte("test"))
	deploy.Show()
	fmt.Println("--------------")
	data, err := deploy.Serialize()
	if err != nil {
		t.Fatal(err)
	}
	dep := &types.Transaction{Payload: new(types.DeployInfo)}
	if err := dep.Deserialize(data); err != nil {
		t.Fatal(err)
	}
	if !dep.Hash.Equals(&deploy.Hash) {
		t.Fatal("hash mismatch")
	}
	dep.Show()
}

func TestInvoke(t *testing.T) {
	i := example.ExampleTestInvoke("main")
	i.Show()
	data, err := i.Serialize()
	if err != nil {
		t.Fatal(err)
	}

	i2 := new(types.Transaction)
	if err := i2.Deserialize(data); err != nil {
		t.Fatal(err)
	}
	if !i2.Hash.Equals(&i.Hash) {
		t.Fatal("hash mismatch")
	}
	i2.Show()
}