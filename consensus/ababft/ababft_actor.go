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
//
// The following is the ababft consensus algorithm.
// Author: Xu Wang, 2018.07.16

package ababft

import (
	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/ecoball/go-ecoball/common/elog"
	"github.com/ecoball/go-ecoball/core/ledgerimpl/ledger"
	"github.com/ecoball/go-ecoball/common/event"
)
type Actor_ababft struct {
	status uint // 1: actor generated,
	// 2: running,
	// 3: as prime, start the new round, collect the tx and previous block signature, then broadcast the first round block
	// 4: as prime, start collect the tx signature and generate the new block, then broadcast
	// 5: as peer, start the new round, signature the current newest block and broadcast
	// 6: as peer, wait for the new block generation, and then update the local ledger
	// 7: as prime, the round end and enters to the next round
	// 8: as peer, the round end and enters to the next round
	pid *actor.PID // actor pid
	service_ababft *Service_ababft
}

const(
	pubkey_tag = "ababft"
	signdata_tag = "ababft"
)

var log = elog.NewLogger("ABABFT", elog.NoticeLog)

var Num_peers int
var Peers_list []Peer_info // Peer information for consensus
var Self_index int // the index of this peer in the peers list
var current_round_num int // current round number
var current_height_num int // current height, according to the blocks saved in the local ledger
var current_ledger ledger.Ledger

var primary_tag int // 0: verification peer; 1: is the primary peer, who generate the block at current round;
var signature_preblock_list [][]byte // list for saving the signatures for the previous block
var signature_BlkF_list [][]byte // list for saving the signatures for the first round block
var block_firstround Block_FirstRound // temporary parameters for saving the first round block
var block_secondround Block_SecondRound // temporary parameters for saving the second round block


func Actor_ababft_gen(actor_ababft *Actor_ababft) (*actor.PID, error) {
	props := actor.FromProducer(func() actor.Actor {
		return actor_ababft
	})
	pid, err := actor.SpawnNamed(props, "Actor_ababft")
	if err != nil {
		return nil, err
	}
	event.RegisterActor(event.ActorConsensus, pid)
	return pid, err
}

func (actor_c *Actor_ababft) Receive(ctx actor.Context) {
	// var err error
	// log.Debug("ababft service receives the message")

	// deal with the message
	switch msg := ctx.Message().(type) {
	case ABABFTStart:
		actor_c.status = 2
		// initialization
		// clear and initialize the signature preblock array
		signature_preblock_list = make([][]byte, len(Peers_list))
		signature_BlkF_list = make([][]byte, len(Peers_list))
		block_firstround = Block_FirstRound{}
		block_secondround = Block_SecondRound{}

		// get the current round number of the block
		currentheader = current_ledger.GetCurrentHeader()
		if currentheader.ConsensusData.Type != types.ConABFT {
			//log.Warn("wrong ConsensusData Type")
			return
		}
		if v,ok:= currentheader.ConsensusData.Payload.(* types.AbaBftData); ok {
			current_payload = *v
		}
	default :
		log.Debug(msg)
		log.Warn("unknown message")
	}
}
