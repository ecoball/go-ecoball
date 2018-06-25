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
	"errors"
	"github.com/ecoball/go-ecoball/common"
	"github.com/ecoball/go-ecoball/core/pb"
	"github.com/ecoball/go-ecoball/common/elog"
)

const (
	Second             = int64(1000)
	BlockInterval      = int64(15000)
	GenerationInterval = GenerationSize * BlockInterval * 10
	GenerationSize     = 4
	ConsensusThreshold = GenerationSize*2/3 + 1
	MaxProduceDuration = int64(5250)
	MinProduceDuration = int64(2250)
)

var log = elog.NewLogger("Consensus", elog.DebugLog)

// Errors in dpos state
var (
	ErrTypeWrong               = errors.New("wrong type")
	ErrTimeNegative            = errors.New("Negative Time")
	ErrNotBlockForgTime = errors.New("current is not time to forge block")
	ErrFoundNilLeader   = errors.New("found a nil leader")
	ErrNilArgument      = errors.New("arguments is nil")
)

type State struct {
	timestamp int64
	leader common.Hash

	//TODO
	bookkeepers []common.Hash
}

func (ds *State) Timestamp() int64 {
	return ds.timestamp
}

func (ds *State) Leader() common.Hash  {
	return ds.leader
}

func (ds *State) NextConsensusState(passedSecond int64) (ConsensusState, error){
	elapsedSecondInMs := passedSecond * Second
	if elapsedSecondInMs <= 0 || elapsedSecondInMs %BlockInterval != 0 {
		return nil, ErrNotBlockForgTime
	}
	//TODO
	bookkeepers := ds.bookkeepers

	consensusState := &State{
		timestamp: ds.timestamp + passedSecond,
		bookkeepers: bookkeepers,
	}

	log.Debug("consensusState, timestamp ", consensusState.timestamp)
	log.Debug(ds.timestamp, passedSecond)
	currentInMs := consensusState.timestamp * Second
	offsetInMs := currentInMs % GenerationInterval
	log.Debug("timestamp %", offsetInMs, (offsetInMs*Second)%BlockInterval)
	var err error
	consensusState.leader, err = FindLeader(consensusState.timestamp, bookkeepers)
	if err != nil {
		log.Debug(err)
		return nil, err
	}
	return consensusState, nil
}

func FindLeader(current int64, bookkeepers []common.Hash) (leader common.Hash, err error) {
	currentInMs := current * Second
	offsetInMs := currentInMs % GenerationInterval
	log.Debug("offsetMs = ", offsetInMs)
	if offsetInMs %BlockInterval != 0 {
		log.Debug("In FindLeader, mod not 0")
		return common.NewHash(nil), ErrNotBlockForgTime
	}
	offset := offsetInMs / BlockInterval
	offset %= GenerationSize

	if offset >= 0 && int(offset) < len(bookkeepers) {
		log.Debug("offset = ", offset)
		leader = bookkeepers[offset]
	} else {
		log.Warn("Can't find Leader")
		return common.NewHash(nil), ErrFoundNilLeader
	}
	return leader, nil
}

func (ds *State) Bookkeepers() ([]common.Hash, error) {
	return ds.bookkeepers, nil
}

func GenesisStateInit(timestamp int64) (ConsensusState, error)  {

	//TODO, bookkeepers
	bookkeepers := []common.Hash{}

	addr1 := common.Address{1,2,3,4,5,6,7,8,9,10,11,12,13,1,2,3,4,5,6,7}
	s1 := addr1.ToBase58()

	addr2 := common.Address{1,2,3,4,5,6,7,8,9,10,11,12,13,1,2,3,4,5,6,8}
	s2 := addr2.ToBase58()

	addr3 := common.Address{1,2,3,4,5,6,7,8,9,10,11,12,13,1,2,3,4,5,6,9}
	s3 := addr3.ToBase58()

	addr4 := common.Address{1,2,3,4,5,6,7,8,9,10,11,12,13,1,2,3,4,5,6,6}
	s4 := addr4.ToBase58()


	addresses := []string{}
	addresses = append(addresses, s1)
	addresses = append(addresses, s2)
	addresses = append(addresses, s3)
	addresses = append(addresses, s4)

	for _, v := range addresses {
		hash := common.NewHash(common.AddressFromBase58(v).Bytes())
		bookkeepers = append(bookkeepers, hash)
	}

	state := &State{
		timestamp: timestamp,
		bookkeepers: bookkeepers,
	}
	return state, nil
}

//TODO
func (s *State) Serialize() ([]byte, error)  {
	p, err := s.protoBuf()
	if err != nil {
		return nil, err
	}
	b, err := p.Marshal()
	if err != nil {
		return nil, err
	}
	return b, nil
}

//TODO
func (s *State) Deserialize(data []byte) error {
	if len(data) == 0 {
		return errors.New("input data's length is zero")
	}
	var state pb.ConsensusState
	if err := state.Unmarshal(data); err != nil {
		return err
	}

	s.timestamp = state.Timestamp
	s.leader = common.NewHash(state.Hash)
	var keepers []common.Hash
	for i := 0; i < len(state.Bookkeepers); i++ {
		bookkeeper := state.Bookkeepers[i]
		keepers = append(keepers, common.NewHash(bookkeeper.Hash))
	}
	s.bookkeepers = keepers
	return nil
}


func (state *State) protoBuf() (*pb.ConsensusState, error) {
	var bookkeepers []*pb.Miner
	for i := 0; i < len(state.bookkeepers); i++ {
		bookkeeper := &pb.Miner{
			Hash: state.bookkeepers[i].Bytes(),
		}
		bookkeepers = append(bookkeepers, bookkeeper)
	}
	consensusState := &pb.ConsensusState{
		state.leader.Bytes(),
		bookkeepers,
		state.timestamp,
	}
	return consensusState, nil
}

