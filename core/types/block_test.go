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
	"github.com/ecoball/go-ecoball/account"
	"github.com/ecoball/go-ecoball/common"
	"github.com/ecoball/go-ecoball/core/bloom"
	"github.com/ecoball/go-ecoball/core/types"
	"testing"
	"time"
)

func TestHeader(t *testing.T) {
	conData := types.ConsensusData{Type:types.ConSolo, Payload:&types.SoloData{}}
	h, err := types.NewHeader(types.VersionHeader, 10, common.Hash{}, common.Hash{}, common.Hash{}, conData, bloom.Bloom{}, time.Now().Unix())
	if err != nil {
		t.Fatal(err)
	}
	acc, err := account.NewAccount(0)
	if err != nil {
		t.Fatal(err)
	}
	if err := h.SetSignature(&acc); err != nil {
		t.Fatal(err)
	}
	data, err := h.Serialize()
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("Header1:", h.Hash.HexString())

	h.Show()
	h2 := new(types.Header)
	if err := h2.Deserialize(data); err != nil {
		t.Fatal(err)
	}
	fmt.Println("Header2:", h.Hash.HexString())
	if !h2.Hash.Equals(&h.Hash) {
		t.Fatal("header error")
	}
	h2.Show()
}

func TestBlockCreate(t *testing.T) {
	g, err := types.GenesesBlockInit()
	if err != nil {
		t.Fatal(err)
	}
	//Sig
	acc, err := account.NewAccount(0)
	if err != nil {
		t.Fatal(err)
	}
	if err := g.SetSignature(&acc); err != nil {
		t.Fatal(err)
	}
	g.Show()
	data, err := g.Serialize()
	if err != nil {
		t.Fatal(err)
	}

	block := new(types.Block)
	if err := block.Deserialize(data); err != nil {
		t.Fatal(err)
	}
	if !block.Hash.Equals(&g.Hash) {
		t.Fatal("error")
	}
	block.Show()
}
