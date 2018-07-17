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
	"github.com/ecoball/go-ecoball/core/types"
	"github.com/ecoball/go-ecoball/common"
	"time"
	"github.com/ecoball/go-ecoball/core/pb"
	"bytes"
	"github.com/ecoball/go-ecoball/crypto/secp256k1"
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
var block_firstround Block_FirstRound // temporary parameters for the first round block
var block_secondround Block_SecondRound // temporary parameters for the second round block
var currentheader *types.Header // temporary parameters for the current block header, according to the blocks saved in the local ledger
var current_payload types.AbaBftData // temporary parameters for current payload
var received_signpre_num int // the number of received signatures for the previous block
var cache_signature_preblk []pb.SignaturePreblock // cache the received signatures for the previous block



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
	var err error
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

		// todo
		// the update of current_round_num
		// current_round_num = int(current_payload.NumberRound)
		// the timeout/changeview message
		// need to check whether the update of current_round_num is necessary

		current_height_num = int(currentheader.Height)
		// signature the current highest block and broadcast
		var signature_preblock common.Signature
		signature_preblock.PubKey = actor_c.service_ababft.account.PublicKey
		signature_preblock.SigData, err = actor_c.service_ababft.account.Sign(currentheader.Hash.Bytes())
		if err != nil {
			return
		}
		// check whether self is the prime or peer
		if current_round_num % Num_peers == (Self_index-1) {
			// if is prime
			primary_tag = 1
			actor_c.status = 3
			received_signpre_num = 0
			// increase the round index
			current_round_num ++
			// set up a timer to wait for the signature_preblock from other peera
			t0 := time.NewTimer(time.Second * WAIT_RESPONSE_TIME * 2)
			go func() {
				select {
				case <-t0.C:
					// timeout for the preblock signature
					err = event.Send(event.ActorConsensus, event.ActorConsensus, PreBlockTimeout{})
					t0.Stop()
				}
			}()
		} else {
			// is peer
			primary_tag = 0
			actor_c.status = 5
			// broadcast the signature_preblock and set up a timer for receiving the data
			var signaturepre_send Signature_Preblock
			signaturepre_send.Signature_preblock.PubKey = signature_preblock.PubKey
			signaturepre_send.Signature_preblock.SigData = signature_preblock.SigData
			signaturepre_send.Signature_preblock.Round = uint32(current_round_num)
			signaturepre_send.Signature_preblock.Height = uint32(currentheader.Height)
			// broadcast
			event.Send(event.ActorConsensus, event.ActorP2P, signaturepre_send)
			// increase the round index
			current_round_num ++
			// set up a timer for receiving the data
			t1 := time.NewTimer(time.Second * WAIT_RESPONSE_TIME * 2)
			go func() {
				select {
				case <-t1.C:
					// timeout for the preblock signature
					err = event.Send(event.ActorConsensus, event.ActorConsensus, TxTimeout{})
					t1.Stop()
				}
			}()
		}
		return

	case Signature_Preblock:
		// the prime will verify the signature for the previous block
		round_in := int(msg.Signature_preblock.Round)
		height_in := int(msg.Signature_preblock.Height)
		if round_in >= current_round_num {
			// cache the Signature_Preblock
			cache_signature_preblk = append(cache_signature_preblk,msg.Signature_preblock)
		}
		if primary_tag == 1 && (actor_c.status == 2 || actor_c.status == 3){
			// verify the signature
			// first check the round number and height
			if round_in >= (current_round_num-1) && height_in >= current_height_num {
				if round_in > (current_round_num - 1) && height_in > current_height_num {
					// require synchronization, the longest chain is ok
					// send synchronization message
					var requestsyn REQSyn
					requestsyn.Reqsyn.PubKey = actor_c.service_ababft.account.PublicKey
					requestsyn.Reqsyn.SigData = []byte("none")
					requestsyn.Reqsyn.RequestHeight = uint64(current_height_num+1)
					event.Send(event.ActorConsensus,event.ActorP2P,requestsyn)
					// todo
					// attention
					// to against the height cheat, do not change the actor_c.status
				} else {
					// check the signature
					pubkey_in := msg.Signature_preblock.PubKey// signaturepre_send.signature_preblock.PubKey = signature_preblock.PubKey
					// check the pubkey_in is in the peer list
					var found_peer bool
					found_peer = false
					var peer_index int
					for index,peer := range Peers_list {
						if ok := bytes.Equal(peer.PublicKey, pubkey_in); ok == true {
							found_peer = true
							peer_index = index
							break
						}
					}
					if found_peer == false {
						// the signature is not from the peer in the list
						return
					}
					// 1. check that signature in or not in list of
					if signature_preblock_list[peer_index] != nil {
						// already receive the signature
						return
					}
					// 2. verify the correctness of the signature
					sigdata_in := msg.Signature_preblock.SigData
					header_hash := currentheader.Hash.Bytes()
					var result_verify bool
					result_verify, err = secp256k1.Verify(header_hash, sigdata_in, pubkey_in)
					if result_verify == true {
						// add the incoming signature to signature preblock list
						signature_preblock_list[peer_index] = sigdata_in
						received_signpre_num ++
					} else {
						return
					}
				}
			} else {
				// the message is old
				return
			}
		} else {
			return
		}

	case PreBlockTimeout:
		if primary_tag == 1 && (actor_c.status == 2 || actor_c.status == 3){


		} else {
			return
		}

	default :
		log.Debug(msg)
		log.Warn("unknown message")
	}
}
