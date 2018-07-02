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

package net

import (
	"github.com/ecoball/go-ecoball/core/types"
	"github.com/ecoball/go-ecoball/net/message"
	eactor "github.com/ecoball/go-ecoball/common/event"
	"github.com/ecoball/go-ecoball/core/ledgerimpl/ledger"

	"gx/ipfs/QmVmDhyTTUcQXFD1rRQ64fGLMSAoaQvNH3hwuaCFAPq2hy/errors"
)

const (
	BLK_SYNC_IDLE uint32 = iota
	BLK_PULL_REQ
	BLK_WAIT_RES
	BLK_SYNC_END
)

type FsmAction func(msg message.EcoBallNetMsg)(uint32,bool)

type FSM interface {
	AddState(uint32, FsmAction) error
	IsRunning() bool
	SetFsmInput(interface{}) error
	Execute()
}

// BlkSyncFsm for supporting Gossip Anti-Entropy mode
type BlkSyncFsm struct {
	FSM

	states            map[uint32]FsmAction
	defaultState      uint32
	currentState      uint32
	inputChan         chan message.EcoBallNetMsg
	netNode           *NetNode
	nodeLedger        ledger.Ledger

}

func NewBlkSyncFsm(node *NetNode, ledger ledger.Ledger) *BlkSyncFsm {
	blkSyncFsm := &BlkSyncFsm{
		states:          make(map[uint32]FsmAction, 3),
		defaultState:    BLK_SYNC_IDLE,
		currentState:    BLK_SYNC_IDLE,
		inputChan:       make(chan message.EcoBallNetMsg),
		netNode:         node,
		nodeLedger:      ledger,
	}

	blkSyncFsm.AddState(BLK_PULL_REQ, blkSyncFsm.PullBlkRequest)
	blkSyncFsm.AddState(BLK_WAIT_RES, blkSyncFsm.HandlePullBlkAckMsg)
	return blkSyncFsm
}

func (this *BlkSyncFsm)AddState(state uint32, handler FsmAction) error {
	if state < BLK_SYNC_END {
		this.states[state] = handler
		return nil
	}

	return errors.New("invalid fsm state")
}

func (this *BlkSyncFsm)IsRunning() bool {
	if this.currentState != BLK_SYNC_IDLE {
		return true
	}
	return false
}

func (this *BlkSyncFsm)SetFsmInput(msg message.EcoBallNetMsg) error {
	this.inputChan <- msg
	return nil
}

// Send local blk state to a random remote peer for a new sync session
// Gossip{key, version}
func (this *BlkSyncFsm)PullBlkRequest(msg message.EcoBallNetMsg) (uint32, bool) {
	log.Debug("send gossip pull blocks request msg")
	height := this.nodeLedger.GetCurrentHeight()
	msgType := message.APP_MSG_GOSSIP_PULL_BLK_REQ
	peers := this.netNode.SelectRandomPeers(1)
	if len(peers) >0 {
		id := this.netNode.SelfRawId()
		blkReq := types.BlkReqMsg{Peer:id, ChainID:1, BlkHeight: height}
		data, _:= blkReq.Serialize()
		netMsg := message.New(msgType, data)
		if err := this.netNode.SendMsg2Peer(peers[0], netMsg); err == nil { //only select a peer to push state
			return BLK_WAIT_RES, false
		}
	}

	return BLK_SYNC_IDLE, true
}

// Handle the block state pull response from remote peer,
// And then send the latest verstion to remote peer
// Gossip {key, value, version}
func (this *BlkSyncFsm) HandlePullBlkAckMsg(msg message.EcoBallNetMsg)(uint32, bool) {
	log.Debug("handle gossip pull blocks response msg")
	blkAckMsg := new(types.BlkAckMsg)
	blkAckMsg.Deserialize(msg.Data())

	header := this.nodeLedger.GetCurrentHeader()
	height := header.Height

	//merge the remote peer's blocks to local ledger
	for _, blk := range blkAckMsg.Data {
		if height < blk.Header.Height {
			eactor.Send(0, eactor.ActorLedger, blk)
		}
	}
	log.Debug("height ", height, "peer height",blkAckMsg.BlkHeight, "block count ", blkAckMsg.BlkCount)
	//send the local latest blosks to remote peer
	var blkCount uint64
	if height > blkAckMsg.BlkHeight {
		blkCount = height - blkAckMsg.BlkHeight
	} else {
		blkCount = 0
	}
	if blkCount > 0 {
		hash := header.Hash
		blkAck2 := &types.BlkAck2Msg{
			ChainID:1,
			BlkCount: blkCount,
			Data:make([]*types.Block, blkCount),
		}

		// it is better to limit the blk count threshold
		for blkCount>0 {
			blk,err := this.nodeLedger.GetTxBlock(hash)
			if err != nil {
				blkAck2.BlkCount = 0
				blkAck2.Data = []*types.Block{}
				break
			}
			blkAck2.Data[blkCount-1] = blk
			hash = blk.Header.PrevHash
			blkCount -= 1
		}
		data, _ := blkAck2.Serialize()
		netMsg := message.New(message.APP_MSG_GOSSIP_PUSH_BLKS, data)
		this.netNode.SendMsg2Peer(blkAckMsg.Peer, netMsg)
	}

	return BLK_SYNC_IDLE, true
}

func (this *BlkSyncFsm)Execute() {
	this.currentState = BLK_PULL_REQ
	if curAction,exist := this.states[this.currentState]; exist {
		for {
			log.Debug("block sync fsm enter:", this.currentState)
			msg := <- this.inputChan
			nextKey, finished := curAction(msg)
			if finished {
				break
			}
			if curAction, exist = this.states[nextKey]; !exist {
				break;
			}
			this.currentState = nextKey
		}
		log.Debug("block sync fsm exit:", this.currentState)
		this.currentState = BLK_SYNC_IDLE
	}
}