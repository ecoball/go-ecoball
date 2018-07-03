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
	"github.com/ecoball/go-ecoball/core/types"
	"github.com/ecoball/go-ecoball/common"
	"github.com/ecoball/go-ecoball/common/elog"
	"errors"
)
var (
	ErrNotBlockForgTime = errors.New("current is not time to forge block")
	ErrFoundNilLeader   = errors.New("found a nil leader")
	ErrNilArgument      = errors.New("arguments is nil")

	ErrTypeWrong               = errors.New("wrong type")
	ErrTimeNegative            = errors.New("Negative Time")

)

var log = elog.NewLogger("Consensus", elog.DebugLog)

type DposBlock struct {
	*types.Block
	state ConsensusState
}

func (block *DposBlock) Timestamp() int64{
	return block.Header.TimeStamp
}

//TODO
func (block *DposBlock) Pack() error {
	return nil
}

//TODO
func (block *DposBlock) ReturnTransactions() {

}

//TODO
func (block *DposBlock) VerifyIntegrity() error {
	return nil
}

//TODO
func (block *DposBlock) LinkParentBlock(chain *Blockchain, parentBlock *DposBlock) error {
	return nil
}

//TODO
func (block *DposBlock) String() string  {
	return ""
}


//TODO
func LoadBlockFromStorage(hash common.Hash, chain *Blockchain) (*DposBlock, error) {
	block, err := chain.chainTx.GetBlock(hash)
	if err != nil {
		log.Debug("GetBlock err")
		return nil, errors.New("Invalid hash")
	}
	//state, err := chain.chainTx.GetConsensusState(hash)

	state := block.ConsensusData.Payload.GetObject().(ConsensusState)

	if err != nil {
		log.Error(err)
		return nil, err
	}

	dposBlock := &DposBlock{
		block,
		state,
	}
	return dposBlock, nil
}

func (block *DposBlock) DposState() (ConsensusState) {
	return block.state
}


