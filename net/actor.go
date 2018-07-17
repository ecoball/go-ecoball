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
	"github.com/AsynkronIT/protoactor-go/actor"
	eactor "github.com/ecoball/go-ecoball/common/event"
	"github.com/ecoball/go-ecoball/core/types"
	"github.com/ecoball/go-ecoball/net/message"
	"github.com/ecoball/go-ecoball/net/rpc"
	"reflect"
	"github.com/ecoball/go-ecoball/consensus/ababft"
)

type NetActor struct {
	props    *actor.Props
	node   	 *NetNode
	gossiper *Gossiper
}

func NewNetActor(node *NetNode, gossiper *Gossiper) *NetActor {
	return &NetActor{
		node:     node,
		gossiper: gossiper,
	}
}

func (this *NetActor) Start() (*actor.PID, error) {
	this.props = actor.FromProducer(func() actor.Actor { return this })
	netPid, err := actor.SpawnNamed(this.props, "net")
	eactor.RegisterActor(eactor.ActorP2P, netPid)
	return netPid, err
}

func (this *NetActor) Receive(ctx actor.Context) {
	var buffer []byte
	var msgType uint32
	msg := ctx.Message()
	switch msg.(type) {
	case *actor.Started:
		log.Debug("NetActor started")
	case *types.Transaction:
		msgType = message.APP_MSG_TRN
		buffer, _ = msg.(*types.Transaction).Serialize()
		netMsg := message.New(msgType, buffer)
		log.Debug("new transactions")
		//this.node.broadCastCh <- netMsg
		this.gossiper.AddPushMsg(netMsg)
		//TODO pubsub
		//this.node.pubSub.Publish("transaction", buffer)
	case *types.Block:
		msgType = message.APP_MSG_BLK
		buffer, _ = msg.(*types.Block).Serialize()
		netMsg := message.New(msgType, buffer)
		log.Debug("p2p push new block")
		//this.node.broadCastCh <- netMsg
		this.gossiper.AddPushMsg(netMsg)
	case *types.BlkReqMsg:
		msgType = message.APP_MSG_GOSSIP_PULL_BLK_REQ
		buffer, _ = msg.(*types.BlkReqMsg).Serialize()
		netMsg := message.New(msgType, buffer)
		this.gossiper.AddPushPullMsg(netMsg)
	case *types.BlkAckMsg:
		msgType = message.APP_MSG_GOSSIP_PULL_BLK_ACK
		buffer, _ = msg.(*types.BlkAckMsg).Serialize()
		netMsg := message.New(msgType, buffer)
		this.gossiper.AddPushPullMsg(netMsg)
	case *types.BlkAck2Msg:
		msgType = message.APP_MSG_GOSSIP_PUSH_BLKS
		buffer, _ = msg.(*types.BlkAck2Msg).Serialize()
		netMsg := message.New(msgType, buffer)
		this.gossiper.AddPushPullMsg(netMsg)
	case *rpc.ListMyIdReq:
		id := this.node.SelfId()
		ctx.Sender().Request(&rpc.ListMyIdRsp{Id:id}, ctx.Self())
	case *rpc.ListPeersReq:
		peers := this.node.Nbrs()
		log.Info(peers)
		ctx.Sender().Request(&rpc.ListPeersRsp{Peer: peers}, ctx.Self())
	case ababft.Signature_Preblock:
		// broadcast the signature for the previous block
		msgType = message.APP_MSG_SIGNPRE
		buffer, _ = msg.(*ababft.Signature_Preblock).Serialize()
		netMsg := message.New(msgType, buffer)
		this.node.broadCastCh <- netMsg
	case ababft.Block_FirstRound:
		// broadcast the first round block
		msgType = message.APP_MSG_BLKF
		buffer, _ = msg.(*ababft.Block_FirstRound).Blockfirst.Serialize()
		netMsg := message.New(msgType, buffer)
		this.node.broadCastCh <- netMsg
	case ababft.REQSyn:
		// broadcast the synchronization request to update the ledger
		msgType = message.APP_MSG_REQSYN
		buffer, _ = msg.(*ababft.REQSyn).Serialize()
		netMsg := message.New(msgType, buffer)
		this.node.broadCastCh <- netMsg
	case ababft.TimeoutMsg:
		msgType = message.APP_MSG_TIMEOUT
		buffer, _ = msg.(*ababft.TimeoutMsg).Serialize()
		netMsg := message.New(msgType, buffer)
		this.node.broadCastCh <- netMsg
		/*

	case ababft.Signature_BlkF:
		// broadcast the signature for the first-round block
		msgType = message.APP_MSG_SIGNBLKF
		buffer, _ = msg.(*ababft.Signature_BlkF).Serialize()
		netMsg := message.New(msgType, buffer)
		this.node.broadCastCh <- netMsg
	case ababft.Block_SecondRound:
		// broadcast the first round block
		msgType = message.APP_MSG_BLKS
		buffer, _ = msg.(*ababft.Block_SecondRound).Blocksecond.Serialize()
		netMsg := message.New(msgType, buffer)
		this.node.broadCastCh <- netMsg

	case ababft.Block_Syn:
		// broadcast the block according to the synchronization request
		msgType = message.APP_MSG_BLKSYN
		buffer, _ = msg.(*ababft.Block_Syn).Blksyn.Serialize()
		netMsg := message.New(msgType, buffer)
		this.node.broadCastCh <- netMsg

		*/
	default:
		log.Error("Error Xmit message ", reflect.TypeOf(ctx.Message()))
	}

	log.Debug("Actor receive msg ", reflect.TypeOf(ctx.Message()))
}
