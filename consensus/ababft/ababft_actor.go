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
		/*
				if primary_tag == 0 && (actor_c.status == 2 || actor_c.status == 5) {

			if blockfirst_received.ConsensusData.Type == types.ConABFT {

				if data_preblk_received.NumberRound < uint32(current_round_num) {
					return
				} else if data_preblk_received.NumberRound > uint32(current_round_num) {

					if (current_height_num+1) < int(blockfirst_received.Header.Height) {

					}
				} else {














				}
			}
		}

		 */


	default :
		log.Debug(msg)
		log.Warn("unknown message")
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