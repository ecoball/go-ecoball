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
	"time"
	"sync/atomic"
	//"github.com/ecoball/go-ecoball/core/types"
	"github.com/ecoball/go-ecoball/net/message"
	eactor "github.com/ecoball/go-ecoball/common/event"
	"github.com/ecoball/go-ecoball/core/ledgerimpl/ledger"
	"gx/ipfs/QmcJukH2sAFjY3HdBKq35WDzWoL3UUu2gt9wdfqZTUyM74/go-libp2p-peer"
	"github.com/ecoball/go-ecoball/core/types"
)

type msgSendCallback func(int, message.EcoBallNetMsg) []peer.ID

type Gossip interface {
	// Add adds a push message
	AddPushMsg(interface{})

	// Add adds a push/pull message
	AddPushPullMsg(interface{})

	// Set the external callback for sending gossip message
	SetMsgSendCallback(cb msgSendCallback)

	// Start starts the component
	Start()

	// Stop stops the component
	Stop()
}

// Gossiper is the implementer of Gossip interface
type Gossiper struct{
	netNode            *NetNode
	nodeLedger         ledger.Ledger
	pushPeerCount      int
	interval           time.Duration
	blkSyncFsm         *BlkSyncFsm
	msgPushChan        chan message.EcoBallNetMsg
	msgPushpullChan    chan message.EcoBallNetMsg
	stopFlag           int32
}

func NewGossiper(node *NetNode, ledg ledger.Ledger) *Gossiper {
	gossiper := &Gossiper{
		netNode:          node,
		nodeLedger:       ledg,
		pushPeerCount:    3,
		interval:         time.Second * 5,
		blkSyncFsm:       NewBlkSyncFsm(node, ledg),
		msgPushChan:      make(chan message.EcoBallNetMsg),
		msgPushpullChan:  make(chan message.EcoBallNetMsg),
		stopFlag:         int32(0),
	}
	return gossiper
}

func (this *Gossiper) Start() {
	atomic.StoreInt32(&(this.stopFlag), int32(0))
	go this.run()
}

func (this *Gossiper) Stop() {
	atomic.StoreInt32(&(this.stopFlag), int32(1))
}

func (this *Gossiper) isDead() bool {
	return atomic.LoadInt32(&(this.stopFlag)) == int32(1)
}

func (this *Gossiper) AddPushMsg(msg message.EcoBallNetMsg) {
	this.msgPushChan <- msg
}

func (this *Gossiper) AddPushPullMsg(msg message.EcoBallNetMsg) {
	this.msgPushpullChan <- msg
}

func (this *Gossiper) run() {
	timer := time.NewTimer(this.interval)
	for !this.isDead() {
		select {
		// Anti-Entropy
		case <- timer.C:
			log.Debug("gossip Anti-Entropy timer emit")
			if running := this.blkSyncFsm.IsRunning(); !running {
				go this.blkSyncFsm.Execute()
				this.blkSyncFsm.SetFsmInput(nil)
			}

			timer.Reset(this.interval)
		case gossipMsg :=  <- this.msgPushpullChan:
			this.handlePushPullMsg(gossipMsg)

		// Rumor-Mongering
		case msgs2bePush := <- this.msgPushChan:
			go this.netNode.SendMsg2Peers(this.pushPeerCount, msgs2bePush)
		}
	}
}

func (this *Gossiper) handlePushPullMsg(msg message.EcoBallNetMsg) {
	if msg.Type() == message.APP_MSG_GOSSIP_PULL_BLK_REQ{
		this.handlePullReqMsg(msg)
	}
	if msg.Type() == message.APP_MSG_GOSSIP_PULL_BLK_ACK {
		if err := this.blkSyncFsm.SetFsmInput(msg); err != nil {
			log.Error(err)
		}
	}
	if msg.Type() == message.APP_MSG_GOSSIP_PUSH_BLKS{
		this.handlePushMsg(msg)
	}
}

func (this *Gossiper) handlePullReqMsg(msg message.EcoBallNetMsg) {
	blkReqMsg := new(types.BlkReqMsg)
	blkReqMsg.Deserialize(msg.Data())
	log.Debug("handle gossip pull blocks request from", blkReqMsg.Peer)
	header := this.nodeLedger.GetCurrentHeader()
	height := header.Height
	peerMaxHeight := blkReqMsg.BlkHeight
	var blkCount uint64
	if height > peerMaxHeight {
		blkCount = height - peerMaxHeight
	} else {
		blkCount = 0
	}
	hash := header.Hash
	blkAck := &types.BlkAckMsg{
		Peer:this.netNode.SelfRawId(),
		ChainID:1,
		BlkHeight:height,
		BlkCount: blkCount,
		Data:make([]*types.Block, blkCount),
	}
	log.Debug("height ", height, "peer height ", peerMaxHeight, blkCount)
	// it is better to limit the blk count threshold
	for blkCount>0 {
		blk,err := this.nodeLedger.GetTxBlock(hash)
		if err != nil {
			blkAck.BlkCount = 0
			blkAck.Data = []*types.Block{}
			break
		}
		blkAck.Data[blkCount-1] = blk
		hash = blk.Header.PrevHash
		blkCount -= 1
	}
	log.Debug("send pull blocks response to ", blkReqMsg.Peer.Pretty())
	data, _ := blkAck.Serialize()
	netMsg := message.New(message.APP_MSG_GOSSIP_PULL_BLK_ACK, data)
	this.netNode.SendMsg2Peer(blkReqMsg.Peer, netMsg)
}

func (this *Gossiper) handlePushMsg(msg message.EcoBallNetMsg) {
	log.Debug("handle gossip push blocks msg")
	blkAck2Msg := new(types.BlkAck2Msg)
	blkAck2Msg.Deserialize(msg.Data())

	header := this.nodeLedger.GetCurrentHeader()
	height := header.Height
	//merge the remote peer's blocks to local ledger
	for _, blk := range blkAck2Msg.Data {
		if height < blk.Header.Height {
			eactor.Send(0, eactor.ActorLedger, blk)
		}
	}
}