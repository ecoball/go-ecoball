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

package ledgerimpl

import (
	"time"
	"github.com/ecoball/go-ecoball/core/types"
	"github.com/ecoball/go-ecoball/common"
	"github.com/ecoball/go-ecoball/core/bloom"
	"github.com/ecoball/go-ecoball/smartcontract/wasmservice"
)

func GenesesBlockInit() (*types.Block, error) {
	tm, err := time.Parse("02/01/2006 15:04:05 PM", "21/02/1990 00:00:00 AM")
	if err != nil {
		return nil, err
	}
	timeStamp := tm.Unix()



	//TODO start
	SecondInMs               := int64(1000)
	BlockIntervalInMs        := int64(15000)
	timeStamp = int64((timeStamp*SecondInMs-SecondInMs)/BlockIntervalInMs) * BlockIntervalInMs
	timeStamp = timeStamp/SecondInMs
	//TODO end

	hash := common.NewHash([]byte("EcoBall Geneses Block"))
	conData := types.GenesesBlockInitConsensusData(timeStamp)
	header, err := types.NewHeader(types.VersionHeader, 1, hash, hash, hash, *conData, bloom.Bloom{}, timeStamp)
	if err != nil {
		return nil, err
	}
	block := types.Block{header, 0, nil}
	return &block, nil
}

func createTransactions() ([]*types.Transaction, error) {
	code, err := wasmservice.ReadWasm("../../../test/token.wasm")
	if err != nil {
		return nil, err
	}
	tx := types.NewTestDeploy(code)
	tx = types.NewTestDeploy(code)
	var txs []*types.Transaction
	txs = append(txs, tx)
	return txs, nil
}