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
	"github.com/ecoball/go-ecoball/net/p2p"
	"github.com/ecoball/go-ecoball/net/ipfs"
	"fmt"
	"context"
	"github.com/ipfs/go-ipfs/core"
	peer "gx/ipfs/QmcJukH2sAFjY3HdBKq35WDzWoL3UUu2gt9wdfqZTUyM74/go-libp2p-peer"
	floodsub "gx/ipfs/QmaWsab8a1KQgoxWP3RjK7mBhSi5PB9pR6NwZUrSXvVd1i/go-libp2p-floodsub"
	"github.com/ecoball/go-ecoball/net/message"
	"github.com/ecoball/go-ecoball/common/elog"
	"github.com/AsynkronIT/protoactor-go/actor"
	"bytes"
	"os"
)

var log = elog.NewLogger("net", elog.DebugLog)

//TODO move to config
var ecoballChainId uint32 = 1

type NetNode struct {
	ctx          context.Context
	ipfsNode     *core.IpfsNode
	self         peer.ID
	network      p2p.EcoballNetwork
	broadCastCh  chan message.EcoBallNetMsg
	handlers 	 map[uint32]message.HandlerFunc
	actorId      *actor.PID
	pubSub       *floodsub.PubSub

	//TODO msg channel
	//msgChannel notifications.PubSub
	//TODO cache check
	//netMsgCache  *lru.Cache
}

func New(parent context.Context, ipfs *core.IpfsNode, network p2p.EcoballNetwork) *NetNode {
	netNode := &NetNode{
		ctx: parent,
		ipfsNode: ipfs,
		self: ipfs.Identity,
		network: network,
		broadCastCh: make(chan message.EcoBallNetMsg, 4 * 1024),//TODO move to config
		handlers: message.MakeHandlers(),
		pubSub: ipfs.Floodsub,
	}

	netNode.broadcastLoop()
	netNode.subTxLoop()
	network.SetDelegate(netNode)
	return netNode
}

func (node *NetNode) broadcastLoop() {
	go func() {
		for {
			select {
			case msg := <-node.broadCastCh:
				//TODO cache check
				//node.netMsgCache.Add(msg.DataSum, msg.Size)
				node.broadcastMessage(msg)
			}
		}
	}()
}

func (node *NetNode) broadcastMessage(msg message.EcoBallNetMsg) {
	peers := node.connectedPeerIds()
	for _, pid := range peers {
		err := node.network.SendMessage(context.Background(), pid, msg)
		if err != nil {
			log.Error("send msg to ", pid.Pretty(), err.Error())
		}
	}
}

func (node *NetNode)subTxLoop()  {
	go func() {
		sub, err := node.pubSub.Subscribe("transaction")
		if err != nil {
			return
		}
		self := []byte(node.self)
		for {
			msg, err := sub.Next(context.Background())
			if err != nil {
				return
			}
			if !bytes.Equal(msg.From, self) {
				message.HdTransactionMsg(msg.Data)
			}
		}
	}()
}

func (node *NetNode) connectedPeerIds() []peer.ID  {
	peers := []peer.ID{}
	host := node.ipfsNode.PeerHost
	if host == nil {
		return peers
	}
	conns := host.Network().Conns()
	for _, c := range conns {
		pid := c.RemotePeer()
		peers = append(peers, pid)
	}
	return peers
}
func (bs *NetNode) ReceiveMessage(ctx context.Context, p peer.ID, incoming message.EcoBallNetMsg) {
	handler, ok := bs.handlers[incoming.Type()]
	if !ok {
		log.Error("get msg ", incoming.Type(), "handler failed")
		return
	}
	err := handler(incoming.Data())
	if err != nil {
		log.Error(err.Error())
	}
}


func (bs *NetNode) ReceiveError(err error) {
	// TODO log the network error
	// TODO bubble the network error up to the parent context/error logger
}

// Connected/Disconnected warns net about peer connections
func (bs *NetNode) PeerConnected(p peer.ID) {
//TODO
}

// Connected/Disconnected warns bitswap about peer connections
func (bs *NetNode) PeerDisconnected(p peer.ID) {
//TODO
}
func (node *NetNode) SelfId() string {
	return node.self.Pretty()
}

func (node *NetNode) Nbrs() []string  {
	peers := []string{}
	host := node.ipfsNode.PeerHost
	if host == nil {
		return peers
	}
	conns := host.Network().Conns()
	for _, c := range conns {
		pid := c.RemotePeer()
		peers = append(peers, pid.Pretty())
	}
	return peers
}

func (node *NetNode) SetActorPid(pid *actor.PID) {
	node.actorId = pid
}

func (node *NetNode) GetActorPid() *actor.PID {
	return node.actorId
}

func SetChainId(id uint32)  {
	ecoballChainId = id
}

func GetChainId() uint32 {
	return ecoballChainId
}

func StartNetWork()  {
	//TODO load config
	//configFile, err := ioutil.ReadFile(ConfigFile)
	//if err != nil {
	//
	//}
	//TODO move to config file
	//InitIpfsConfig(path)
	var path = "./store"
	ipfsNode, err := ipfs.StartIpfsNode(path)
	if err != nil {
		panic(err)
		os.Exit(1)
	}
	network := p2p.NewFromIpfsHost(ipfsNode.PeerHost, ipfsNode.Routing)
	netNode := New(context.Background(), ipfsNode, network)
	netActor := NewNetActor(netNode)
	actorId, _ := netActor.Start()
	netNode.SetActorPid(actorId)
	fmt.Printf("i am %s \n", netNode.SelfId())
}