// Copyright 2018 The go-ecoball Authors
// This file is part of the go-ecoball.
//
// The go-ecoball is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ecoball is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ecoball. If not, see <http://www.gnu.org/licenses/>.


package dpos

import (
	"github.com/ecoball/go-ecoball/common"
	"github.com/ecoball/go-ecoball/core/types"
	"github.com/ecoball/go-ecoball/core/ledgerimpl/ledger"
)

type ChainInterface interface {
	GetBlock(hash common.Hash) (*types.Block, error)
	GetConsensusState(hash common.Hash) (ConsensusState, error)
	SaveBlock(block *types.Block) error
	SaveConsensusState(block *DposBlock) error
	//NewBlock(ledger ledger.Ledger, txs []*types.Transaction, consensusData types.ConsensusData) (*types.Block, error)
	NewBlock(ledger ledger.Ledger, txs []*types.Transaction, consensusData types.ConsensusData) (*types.Block, error)
	GetTailBlockHash() (common.Hash)
}
