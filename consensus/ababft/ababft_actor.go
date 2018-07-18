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
	"fmt"
	"github.com/ecoball/go-ecoball/common/message"
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
var block_first_cal *types.Block // cache the first-round block
var received_signblkf_num int // temporary parameters for received signatures for first round block
var TimeoutMsgs = make(map[string]int, 1000) // cache the timeout message

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
			// 1. check the cache cache_signature_preblk
			header_hash := currentheader.Hash.Bytes()
			for _,signpreblk := range cache_signature_preblk {
				round_in := signpreblk.Round
				if int(round_in) != current_round_num {
					continue
				}
				// check the signature
				pubkey_in := signpreblk.PubKey// signaturepre_send.signature_preblock.PubKey = signature_preblock.PubKey
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
					continue
				}
				// first check that signature in or not in list of
				if signature_preblock_list[peer_index] != nil {
					// already receive the signature
					continue
				}
				// second, verify the correctness of the signature
				sigdata_in := signpreblk.SigData
				var result_verify bool
				result_verify, err = secp256k1.Verify(header_hash, sigdata_in, pubkey_in)
				if result_verify == true {
					// add the incoming signature to signature preblock list
					signature_preblock_list[peer_index] = sigdata_in
					received_signpre_num ++
				} else {
					continue
				}

			}
			// clean the cache_signature_preblk
			cache_signature_preblk = make([]pb.SignaturePreblock,len(Peers_list)*2)
			fmt.Println("valid sign_pre:",received_signpre_num)
			// 2. check the number of the preblock signature
			if received_signpre_num >= int(len(Peers_list)/3+1) {
				// enough preblock signature, so generate the first-round block, only including the preblock signatures and
				// prepare the ConsensusData
				var signpre_send []common.Signature
				for index,signpre := range signature_preblock_list {
					if signpre != nil {
						var sign_tmp common.Signature
						sign_tmp.SigData = signpre
						sign_tmp.PubKey = Peers_list[index].PublicKey
						signpre_send = append(signpre_send, sign_tmp)
					}
				}
				conData := types.ConsensusData{Type: types.ConABFT, Payload: &types.AbaBftData{uint32(current_round_num),signpre_send}}
				fmt.Println("conData for blk firstround",conData)
				// prepare the tx list
				value, err := event.SendSync(event.ActorTxPool, message.GetTxs{}, time.Second*1)
				if err != nil {
					log.Error("AbaBFT Consensus error:", err)
					return
				}
				txList, ok := value.(*types.TxsList)
				if !ok {
					// log.Error("The format of value error [solo]")
					return
				}
				var txs []*types.Transaction
				for _, v := range txList.Txs {
					txs = append(txs, v)
				}
				// generate the first-round block
				var block_first *types.Block
				block_first,err = actor_c.service_ababft.ledger.NewTxBlock(txs,conData)
				block_first.SetSignature(actor_c.service_ababft.account)
				// broadcast the first-round block to peers for them to verify the transactions and wait for the corresponding signatures back
				block_firstround.Blockfirst = *block_first
				event.Send(event.ActorConsensus, event.ActorP2P, block_firstround)
				// change the statue
				actor_c.status = 4
				// initial the received_signblkf_num to count the signatures for txs (i.e. the first round block)
				received_signblkf_num = 0
				// set the timer for collecting the signature for txs (i.e. the first round block)
				t2 := time.NewTimer(time.Second * WAIT_RESPONSE_TIME)
				go func() {
					select {
					case <-t2.C:
						// timeout for the preblock signature
						err = event.Send(event.ActorConsensus, event.ActorConsensus, SignTxTimeout{})
						t2.Stop()
					}
				}()
			} else {
				// did not receive enough preblock signature in the assigned time interval
				actor_c.status = 7
				primary_tag = 0 // reset to zero, and the next primary will take the turn
				// send out the timeout message
				var timeoutmsg TimeoutMsg
				timeoutmsg.Toutmsg.RoundNumber = uint64(current_round_num)
				timeoutmsg.Toutmsg.PubKey = actor_c.service_ababft.account.PublicKey
				timeoutmsg.Toutmsg.SigData = []byte("none")
				event.Send(event.ActorConsensus,event.ActorP2P,timeoutmsg)
				// start/enter the next turn
				event.Send(event.ActorConsensus, event.ActorConsensus, ABABFTStart{})
			}
		} else {
			return
		}

	case Block_FirstRound:
		if primary_tag == 0 && (actor_c.status == 2 || actor_c.status == 5) {
			// to verify the first round block
			blockfirst_received := msg.Blockfirst
			// the protocal type is ababft
			if blockfirst_received.ConsensusData.Type == types.ConABFT {
				data_preblk_received := blockfirst_received.ConsensusData.Payload.(*types.AbaBftData)
				// 1. check the round number
				// 1a. current round number
				if data_preblk_received.NumberRound < uint32(current_round_num) {
					return
				} else if data_preblk_received.NumberRound > uint32(current_round_num) {
					// require synchronization, the longest chain is ok
					// in case that somebody may skip the current generator, only the different height can call the synchronization
					if (current_height_num+1) < int(blockfirst_received.Header.Height) {
						// send synchronization message
						var requestsyn REQSyn
						requestsyn.Reqsyn.PubKey = actor_c.service_ababft.account.PublicKey
						requestsyn.Reqsyn.SigData = []byte("none")
						requestsyn.Reqsyn.RequestHeight = uint64(current_height_num+1)
						event.Send(event.ActorConsensus,event.ActorP2P,requestsyn)
						// todo
						// attention:
						// to against the height cheat, do not change the actor_c.status
					}
				} else {
					// 1b. the round number corresponding to the block generator
					index_g := (current_round_num-1) % Num_peers + 1
					pukey_g_in := blockfirst_received.Signatures[0].PubKey
					var index_g_in int
					index_g_in = -1
					for _, peer := range Peers_list {
						if ok := bytes.Equal(peer.PublicKey, pukey_g_in); ok == true {
							index_g_in = int(peer.Index)
							break
						}
					}
					if index_g != index_g_in {
						// illegal block generator
						return
					}
					// 1c. check the block header, except the consensus data
					var valid_blk bool
					valid_blk,err = actor_c.verify_header(&blockfirst_received, current_round_num)
					if valid_blk==false {
						println("header check fail")
						return
					}
					// 2. check the preblock signature
					sign_preblk_list := data_preblk_received.PerBlockSignatures
					header_hash := currentheader.Hash.Bytes()
					var num_verified int
					num_verified = 0
					for index,sign_preblk := range sign_preblk_list {
						// 2a. check the peers in the peer list
						var peerin_tag bool
						peerin_tag = false
						for _, peer := range Peers_list {
							if ok := bytes.Equal(peer.PublicKey, sign_preblk.PubKey); ok == true {
								peerin_tag = true
								break
							}
						}
						if peerin_tag == false {
							// there exists signature not from the peer list
							fmt.Println("the signature is not from the peer list, its index is:", index)
							return
						}
						// 2b. verify the correctness of the signature
						pubkey_in := sign_preblk.PubKey
						sigdata_in := sign_preblk.SigData
						var result_verify bool
						result_verify, err = secp256k1.Verify(header_hash, sigdata_in, pubkey_in)
						if result_verify == true {
							num_verified++
						}
					}
					// 2c. check the valid signature number
					if num_verified < int(len(Peers_list)/3+1){
						// not enough signature
						return
					}
					// 3. check the txs
					txs_in := blockfirst_received.Transactions
					for index1,tx_in := range txs_in {
						err = actor_c.service_ababft.ledger.CheckTransaction(tx_in)
						if err != nil {
							println("wrong tx, index:", index1)
							return
						}
					}
					// 4. sign the received first-round block
					var sign_blkf_send Signature_BlkF
					sign_blkf_send.Signature_blkf.PubKey = actor_c.service_ababft.account.PublicKey
					sign_blkf_send.Signature_blkf.SigData,err = actor_c.service_ababft.account.Sign(blockfirst_received.Header.Hash.Bytes())
					// 5. broadcast the signature of the first round block
					event.Send(event.ActorConsensus, event.ActorP2P, sign_blkf_send)
					// 6. change the status
					actor_c.status = 6
					fmt.Println("sign_blkf_send:",sign_blkf_send)
					// clean the cache_signature_preblk
					cache_signature_preblk = make([]pb.SignaturePreblock,len(Peers_list)*2)
					// send the received first-round block to other peers in case that network is not good
					block_firstround.Blockfirst = blockfirst_received
					event.Send(event.ActorConsensus,event.ActorP2P,block_firstround)
					// 7. set the timer for waiting the second-round(final) block
					t3 := time.NewTimer(time.Second * WAIT_RESPONSE_TIME)
					go func() {
						select {
						case <-t3.C:
							// timeout for the second-round(final) block
							err = event.Send(event.ActorConsensus, event.ActorConsensus, BlockSTimeout{})
							t3.Stop()
						}
					}()
				}
			}
		}

	case TxTimeout:
		if primary_tag == 0 && (actor_c.status == 2 || actor_c.status == 5) {
			// not receive the first round block
			// change the status
			actor_c.status = 8
			primary_tag = 0
			// send out the timeout message
			var timeoutmsg TimeoutMsg
			timeoutmsg.Toutmsg.RoundNumber = uint64(current_round_num)
			timeoutmsg.Toutmsg.PubKey = actor_c.service_ababft.account.PublicKey
			timeoutmsg.Toutmsg.SigData = []byte("none")
			event.Send(event.ActorConsensus,event.ActorP2P,timeoutmsg)
			// start/enter the next turn
			event.Send(event.ActorConsensus, event.ActorConsensus, ABABFTStart{})
			// todo
			// the above needed to be checked
			// here, enter the next term and broadcast the preblock signature with the increased round number has the same effect as the changeview/ nextround message
			// handle of the timeout message has been added, please check case TimeoutMsg
			return
		}

	case Signature_BlkF:
		// the prime will verify the signatures of first-round block from peers
		if primary_tag == 1 && actor_c.status == 4 {
			// verify the signature
			// 1. check the peer in the peers list
			pubkey_in := msg.Signature_blkf.PubKey
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
			// 2. verify the correctness of the signature
			if signature_BlkF_list[peer_index] != nil {
				// already receive the signature
				return
			}
			sigdata_in := msg.Signature_blkf.SigData
			header_hash := block_firstround.Blockfirst.Header.Hash.Bytes()
			var result_verify bool
			result_verify, err = secp256k1.Verify(header_hash, sigdata_in, pubkey_in)
			if result_verify == true {
				// add the incoming signature to signature preblock list
				signature_BlkF_list[peer_index] = sigdata_in
				received_signblkf_num ++
				return
			} else {
				return
			}

		}

	case SignTxTimeout:
		if primary_tag == 1 && actor_c.status == 4 {
			// check the number of the signatures of first-round block from peers
			if received_signblkf_num >= int(2*len(Peers_list)/3+1) {
				// enough first-round block signatures, so generate the second-round(final) block
				// 1. add the first-round block signatures into ConsensusData
				pubkey_tag_b := []byte(pubkey_tag)
				signdata_tag_b := []byte(signdata_tag)
				var sign_tag common.Signature
				sign_tag.PubKey = pubkey_tag_b
				sign_tag.SigData = signdata_tag_b

				ababftdata := block_firstround.Blockfirst.ConsensusData.Payload.(*types.AbaBftData)
				// prepare the ConsensusData
				// add the tag to distinguish preblock signature and second round signature
				ababftdata.PerBlockSignatures = append(ababftdata.PerBlockSignatures, sign_tag)
				for index,signblkf := range signature_BlkF_list {
					if signblkf != nil {
						var sign_tmp common.Signature
						sign_tmp.SigData = signblkf
						sign_tmp.PubKey = Peers_list[index].PublicKey
						ababftdata.PerBlockSignatures = append(ababftdata.PerBlockSignatures, sign_tmp)
					}
				}

				conData := types.ConsensusData{Type: types.ConABFT, Payload: &types.AbaBftData{uint32(current_round_num),ababftdata.PerBlockSignatures}}
				// 2. generate the second-round(final) block
				var block_second types.Block
				block_second,err =  actor_c.update_block(block_firstround.Blockfirst, conData)
				block_second.SetSignature(actor_c.service_ababft.account)
				// 3. broadcast the second-round(final) block
				block_secondround.Blocksecond = &block_second
				event.Send(event.ActorConsensus, event.ActorP2P, block_secondround)
				// 4. save the second-round(final) block to ledger
				if err = actor_c.service_ababft.ledger.SaveTxBlock(&block_second); err != nil {
					// log.Error("save block error:", err)
					println("save block error:", err)
					return
				}
				// 5. change the status
				actor_c.status = 7
				primary_tag = 0
				// start/enter the next turn
				event.Send(event.ActorConsensus, event.ActorConsensus, ABABFTStart{})
				return
			} else {
				// 1. did not receive enough signatures of first-round block from peers in the assigned time interval
				actor_c.status = 7
				primary_tag = 0 // reset to zero, and the next primary will take the turn
				// 2. reset the stateDB
				err = actor_c.service_ababft.ledger.ResetStateDB(currentheader.Hash)
				if err != nil {
					log.Debug("ResetStateDB fail")
					return
				}
				// send out the timeout message
				var timeoutmsg TimeoutMsg
				timeoutmsg.Toutmsg = new(pb.ToutMsg)
				timeoutmsg.Toutmsg.RoundNumber = uint64(current_round_num)
				timeoutmsg.Toutmsg.PubKey = actor_c.service_ababft.account.PublicKey
				timeoutmsg.Toutmsg.SigData = []byte("none")
				event.Send(event.ActorConsensus,event.ActorP2P,timeoutmsg)
				// 3. start/enter the next turn
				event.Send(event.ActorConsensus, event.ActorConsensus, ABABFTStart{})
			}
		}

	case Block_SecondRound:
		if primary_tag == 0 && (actor_c.status == 6 || actor_c.status == 2 || actor_c.status == 5) {
			// to verify the first round block
			blocksecond_received := msg.Blocksecond
			// check the protocal type is ababft
			if blocksecond_received.ConsensusData.Type == types.ConABFT {
				data_blks_received := blocksecond_received.ConsensusData.Payload.(*types.AbaBftData)
				// 1. check the round number and height
				// 1a. current round number
				if data_blks_received.NumberRound < uint32(current_round_num) || blocksecond_received.Header.Height <= uint64(current_height_num) {
					return
				} else if (blocksecond_received.Header.Height-1) > uint64(current_height_num) {
					// send synchronization message
					var requestsyn REQSyn
					requestsyn.Reqsyn.PubKey = actor_c.service_ababft.account.PublicKey
					requestsyn.Reqsyn.SigData = []byte("none")
					requestsyn.Reqsyn.RequestHeight = uint64(current_height_num+1)
					event.Send(event.ActorConsensus,event.ActorP2P,requestsyn)

					// todo
					// attention:
					// to against the height cheat, do not change the actor_c.status
				} else {
					// here, the add new block into the ledger, data_blks_received.NumberRound >= current_round_num is ok instead of data_blks_received.NumberRound == current_round_num
					// 1b. the round number corresponding to the block generator
					index_g := (int(data_blks_received.NumberRound)-1) % Num_peers + 1
					pukey_g_in := blocksecond_received.Signatures[0].PubKey
					var index_g_in int
					index_g_in = -1
					for _, peer := range Peers_list {
						if ok := bytes.Equal(peer.PublicKey, pukey_g_in); ok == true {
							index_g_in = int(peer.Index)
							break
						}
					}
					if index_g != index_g_in {
						// illegal block generator
						return
					}
					// 1c. check the block header, except the consensus data
					var valid_blk bool
					valid_blk,err = actor_c.verify_header(blocksecond_received, int(data_blks_received.NumberRound))
					// todo
					// can check the hash and statdb and merker root instead of the total head to speed up
					if valid_blk==false {
						println("header check fail")
						return
					}
					// 2. check the signatures ( for both previous and current blocks) in ConsensusData
					preblkhash := currentheader.Hash
					valid_blk, err = actor_c.verify_signatures(data_blks_received, preblkhash, blocksecond_received.Header)
					if valid_blk==false {
						println("previous and first-round blocks signatures check fail")
						return
					}
					// 3.save the second-round block into the ledger
					if err = actor_c.service_ababft.ledger.SaveTxBlock(blocksecond_received); err != nil {
						// log.Error("save block error:", err)
						println("save block error:", err)
						return
					}
					// 4. change status
					actor_c.status = 8
					primary_tag = 0
					// update the current_round_num
					current_round_num = int(data_blks_received.NumberRound)
					// start/enter the next turn
					event.Send(event.ActorConsensus, event.ActorConsensus, ABABFTStart{})
					// 5. broadcast the received second-round block, which has been checked valid
					// to let other peer know this block
					block_secondround.Blocksecond = blocksecond_received
					event.Send(event.ActorConsensus, event.ActorP2P, block_secondround)
					return
				}
			}
		}
	case BlockSTimeout:
		if primary_tag == 0 && actor_c.status == 5 {
			actor_c.status = 8
			primary_tag = 0
			// reset the state of merkle tree, statehash and so on
			err = actor_c.service_ababft.ledger.ResetStateDB(currentheader.Hash)
			if err != nil {
				log.Debug("ResetStateDB fail")
				return
			}
			// send out the timeout message
			var timeoutmsg TimeoutMsg
			timeoutmsg.Toutmsg.RoundNumber = uint64(current_round_num)
			timeoutmsg.Toutmsg.PubKey = actor_c.service_ababft.account.PublicKey
			timeoutmsg.Toutmsg.SigData = []byte("none")
			event.Send(event.ActorConsensus,event.ActorP2P,timeoutmsg)
			// start/enter the next turn
			event.Send(event.ActorConsensus, event.ActorConsensus, ABABFTStart{})
			return
		}

	case REQSyn:
		// receive the shronization request
		height_req := msg.Reqsyn.RequestHeight
		// 1. get the response block from the ledger
		blk_syn,err := actor_c.service_ababft.ledger.GetTxBlockByHeight(height_req)
		if err== nil && blk_syn!=nil {
			// find the corresponding block
			var blksyn_send Block_Syn
			blksyn_send.Blksyn = blk_syn
			// 2. send the required block to
			event.Send(event.ActorConsensus,event.ActorP2P,blksyn_send)
		}

	case Block_Syn:
		height_syn := msg.Blksyn.Header.Height
		// 1. compare the height
		if (current_height_num + 1) == int(height_syn) {
			// 2. to check and save the block if it passes the verification
			// to verify the first round block
			blocksecond_received := msg.Blksyn
			// check the protocal type is ababft
			if blocksecond_received.ConsensusData.Type == types.ConABFT {
				data_blks_received := blocksecond_received.ConsensusData.Payload.(*types.AbaBftData)
				// 1. no need to check the round number, maybe the peer misses one block in the previous round.
				// Only the height is important
				round_num_in := int(data_blks_received.NumberRound)
				index_g := (int(data_blks_received.NumberRound)-1) % Num_peers + 1
				pukey_g_in := blocksecond_received.Signatures[0].PubKey
				var index_g_in int
				index_g_in = -1
				for _, peer := range Peers_list {
					if ok := bytes.Equal(peer.PublicKey, pukey_g_in); ok == true {
						index_g_in = int(peer.Index)
						break
					}
				}
				if index_g != index_g_in {
					// illegal block generator
					return
				}
				// 1c. check the block header, except the consensus data
				var valid_blk bool
				valid_blk,err = actor_c.verify_header(blocksecond_received, round_num_in)
				if valid_blk==false {
					println("header check fail")
					return
				}
				// 2. check the signatures ( for both previous and current blocks) in ConsensusData
				preblkhash := currentheader.Hash
				valid_blk, err = actor_c.verify_signatures(data_blks_received, preblkhash, blocksecond_received.Header)
				if valid_blk==false {
					println("previous and first-round blocks signatures check fail")
					return
				}
				// 3.save the second-round block into the ledger
				if err = actor_c.service_ababft.ledger.SaveTxBlock(blocksecond_received); err != nil {
					// log.Error("save block error:", err)
					println("save block error:", err)
					return
				}
				// 4. only the block is sucessfully saved, then change the status
				actor_c.status = 8
				primary_tag = 0
				// update the current_round_num
				current_round_num = int(data_blks_received.NumberRound)
				// start/enter the next turn
				event.Send(event.ActorConsensus, event.ActorConsensus, ABABFTStart{})
				return

				// todo
				// only need to check the hash and signature is enough?
				// this may help to speed up the ababft
			}
		}

	case TimeoutMsg:
		pubkey_in := msg.Toutmsg.PubKey
		round_in := int(msg.Toutmsg.RoundNumber)
		// check the peer in the peers list
		if round_in < current_round_num {
			return
		}
		for _, peer := range Peers_list {
			if ok := bytes.Equal(peer.PublicKey, pubkey_in); ok == true {
				// legal peer
				TimeoutMsgs[string(pubkey_in)] = round_in
				// to count the number is enough
				var count_r []int
				var max_r int
				max_r = 0
				for _,v := range TimeoutMsgs {
					if v > current_round_num {
						count_r[v-current_round_num]++
					}
					if v > max_r {
						max_r = v
					}
				}
				var total_count int
				total_count = 0
				for i := max_r-current_round_num; i > 0; i-- {
					total_count = total_count + count_r[i]
					if total_count > int(2*len(Peers_list)/3+1) {
						// reset the round number
						current_round_num += i
						// start/enter the next turn
						actor_c.status = 8
						primary_tag = 0
						event.Send(event.ActorConsensus, event.ActorConsensus, ABABFTStart{})
						break
					}
				}
				break
			}
		}
		return

	default :
		log.Debug(msg)
		log.Warn("unknown message")
		return
	}
}

func (actor_c *Actor_ababft) verify_header(block_in *types.Block, current_round_num_in int) (bool,error){
	var err error
	header_in := block_in.Header
	txs := block_in.Transactions
	data_preblk_received := block_in.ConsensusData.Payload.(*types.AbaBftData)
	signpre_send := data_preblk_received.PerBlockSignatures
	condata_c := types.ConsensusData{Type:types.ConABFT, Payload:&types.AbaBftData{uint32(current_round_num_in),signpre_send}}

	fmt.Println("before reset")
	// reset the stateDB
	err = actor_c.service_ababft.ledger.ResetStateDB(currentheader.Hash)
	// fmt.Println("after reset",err)

	// generate the block_first_cal for comparison
	block_first_cal,err = actor_c.service_ababft.ledger.NewTxBlock(txs,condata_c)

	fmt.Println("block_first_cal:",block_first_cal)

	var num_txs int
	num_txs = int(block_in.CountTxs)
	if num_txs != len(txs) {
		println("tx number is wrong")
		return false,nil
	}
	// check Height        uint64
	if current_height_num >= int(header_in.Height) {
		println("the height is not higher than current height")
		return false,nil
	}
	// ConsensusData is checked in the Receive function

	// check PrevHash      common.Hash
	if ok :=bytes.Equal(block_in.PrevHash.Bytes(),currentheader.Hash.Bytes()); ok != true {
		println("prehash is wrong")
		return false,nil
	}
	// check MerkleHash    common.Hash
	if ok := bytes.Equal(block_first_cal.MerkleHash.Bytes(),block_in.MerkleHash.Bytes()); ok != true {
		println("MercleHash is wrong")
		return false,nil
	}
	fmt.Println("mercle:",block_first_cal.MerkleHash.Bytes(),block_in.MerkleHash.Bytes())

	// check StateHash     common.Hash
	if ok := bytes.Equal(block_first_cal.StateHash.Bytes(),block_in.StateHash.Bytes()); ok != true {
		println("StateHash is wrong")
		return false,nil
	}

	fmt.Println("statehash:",block_first_cal.StateHash.Bytes(),block_in.StateHash.Bytes())


	// check Bloom         bloom.Bloom
	if ok := bytes.Equal(block_first_cal.Bloom.Bytes(), block_in.Bloom.Bytes()); ok != true {
		println("bloom is wrong")
		return false,nil
	}
	// check Hash common.Hash
	header_cal,err1 := types.NewHeader(header_in.Version, header_in.Height, header_in.PrevHash,
		header_in.MerkleHash, header_in.StateHash, header_in.ConsensusData, header_in.Bloom, header_in.TimeStamp)
	if ok := bytes.Equal(header_cal.Hash.Bytes(),header_in.Hash.Bytes()); ok != true {
		println("Hash is wrong")
		return false,err1
	}
	// check Signatures    []common.Signature
	signpre_in := block_in.Signatures[0]
	pubkey_g_in := signpre_in.PubKey
	signdata_in := signpre_in.SigData
	var sign_verify bool
	sign_verify, err = secp256k1.Verify(header_in.Hash.Bytes(), signdata_in, pubkey_g_in)
	if sign_verify != true {
		println("signature is wrong")
		return false,err
	}
	return true,err
}

func (actor_c *Actor_ababft) update_block(block_first types.Block, condata types.ConsensusData) (types.Block,error){
	var block_second types.Block
	var err error
	header_in := block_first.Header
	header, _ := types.NewHeader(header_in.Version, header_in.Height, header_in.PrevHash, header_in.MerkleHash,
		header_in.StateHash, condata, header_in.Bloom, header_in.TimeStamp)
	block_second = types.Block{header, uint32(len(block_first.Transactions)), block_first.Transactions}
	return block_second,err
}

func (actor_c *Actor_ababft) verify_signatures(data_blks_received *types.AbaBftData, preblkhash common.Hash, curheader *types.Header) (bool,error){
	var err error
	// 1. devide the signatures into two part
	var sign_blks_preblk []common.Signature
	var sign_blks_curblk []common.Signature
	pubkey_tag_byte := []byte(pubkey_tag)
	sigdata_tag_byte := []byte(signdata_tag)
	var tag_sign int
	tag_sign = 0
	for _,sign := range data_blks_received.PerBlockSignatures {
		ok1 := bytes.Equal(sign.PubKey, pubkey_tag_byte);
		ok2 := bytes.Equal(sign.SigData,sigdata_tag_byte);
		if ok1 == true && ok2 == true {
			tag_sign = 1
			continue
		}
		if tag_sign == 0 {
			sign_blks_preblk = append(sign_blks_preblk,sign)
		} else if tag_sign == 1 {
			sign_blks_curblk = append(sign_blks_curblk,sign)
		}
	}

	// 2. check the preblock signature
	var num_verified int
	num_verified = 0
	for index,sign_preblk := range sign_blks_preblk {
		// 2a. check the peers in the peer list
		var peerin_tag bool
		peerin_tag = false
		for _, peer := range Peers_list {
			if ok := bytes.Equal(peer.PublicKey, sign_preblk.PubKey); ok == true {
				peerin_tag = true
				break
			}
		}
		if peerin_tag == false {
			// there exists signature not from the peer list
			fmt.Println("the signature is not from the peer list, its index is:", index)
			return false,nil
		}
		// 2b. verify the correctness of the signature
		pubkey_in := sign_preblk.PubKey
		sigdata_in := sign_preblk.SigData
		var result_verify bool
		result_verify, err = secp256k1.Verify(preblkhash.Bytes(), sigdata_in, pubkey_in)
		if result_verify == true {
			num_verified++
		}
	}
	// 2c. check the valid signature number
	if num_verified < int(len(Peers_list)/3+1){
		fmt.Println(" not enough signature for the previous block:", num_verified)
		return false,nil
	}

	// 3. check the current block signature
	num_verified = 0
	// calculate firstround block header hash for the check of the first-round block signatures
	conData := types.ConsensusData{Type: types.ConABFT, Payload: &types.AbaBftData{uint32(current_round_num),sign_blks_preblk}}
	header_recal, _ := types.NewHeader(curheader.Version, curheader.Height, curheader.PrevHash, curheader.MerkleHash,
		curheader.StateHash, conData, curheader.Bloom, curheader.TimeStamp)
	blkFhash := header_recal.Hash
	for index,sign_curblk := range sign_blks_curblk {
		// 3a. check the peers in the peer list
		var peerin_tag bool
		peerin_tag = false
		for _, peer := range Peers_list {
			if ok := bytes.Equal(peer.PublicKey, sign_curblk.PubKey); ok == true {
				peerin_tag = true
				break
			}
		}
		if peerin_tag == false {
			// there exists signature not from the peer list
			fmt.Println("the signature is not from the peer list, its index is:", index)
			return false,nil
		}
		// 3b. verify the correctness of the signature
		pubkey_in := sign_curblk.PubKey
		sigdata_in := sign_curblk.SigData
		var result_verify bool
		result_verify, err = secp256k1.Verify(blkFhash.Bytes(), sigdata_in, pubkey_in)
		if result_verify == true {
			num_verified++
		}
	}
	// 3c. check the valid signature number
	if num_verified < int(2*len(Peers_list)/3+1){
		fmt.Println(" not enough signature for first round block:", num_verified)
		return false,nil
	}
	return  true,err

	// todo
	// use CheckPermission(index common.AccountName, name string, sig []common.Signature) instead
	/*
	// 4. check the current block signature by using function CheckPermission
	// 4a. check the peers permission
	err = actor_c.service_ababft.ledger.CheckPermission(0, "active",sign_blks_curblk)
	if err != nil {
		log.Debug("signature permission check fail")
		return false,err
	}
	num_verified = 0
	// calculate firstround block header hash for the check of the first-round block signatures
	conData := types.ConsensusData{Type: types.ConABFT, Payload: &types.AbaBftData{uint32(current_round_num),sign_blks_preblk}}
	header_recal, _ := types.NewHeader(curheader.Version, curheader.Height, curheader.PrevHash, curheader.MerkleHash,
		curheader.StateHash, conData, curheader.Bloom, curheader.TimeStamp)
	blkFhash := header_recal.Hash
	for _,sign_curblk := range sign_blks_curblk {
		// 4b. verify the correctness of the signature
		pubkey_in := sign_curblk.PubKey
		sigdata_in := sign_curblk.SigData
		var result_verify bool
		result_verify, err = secp256k1.Verify(blkFhash.Bytes(), sigdata_in, pubkey_in)
		if result_verify == true {
			num_verified++
		}
	}
	// 4c. check the valid signature number
	if num_verified < int(2*len(Peers_list)/3+1){
		fmt.Println(" not enough signature for first round block:", num_verified)
		return false,nil
	}
	return  true,err
	*/
}