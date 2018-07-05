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
package sharding

import (
	"gx/ipfs/QmaWsab8a1KQgoxWP3RjK7mBhSi5PB9pR6NwZUrSXvVd1i/go-libp2p-floodsub"
	"time"
)

type Pubkey string

type shard struct {
	key           Pubkey
	leader        string
	leaderBackup  string
	memberCount   uint32
}

type Committee struct {
	shards        map[uint32]shard
	leader        string
	pnPeers       []string
	fnPeers       []string
	topics        map[string]*floodsub.PubSub
	shardLeaders  []string
	epoch         *time.Timer
	preColletWait *time.Timer
	node          *ShardNode
}

func NewCommittee(node *ShardNode) *Committee {
	return &Committee{
		shards:  make(map[uint32]shard),
		pnPeers: make([]string, 256), //TODO:move to config
		fnPeers: make([]string, 1024), //TODO:move to config
		topics:  make(map[string]*floodsub.PubSub),
		shardLeaders: make([]string, 256), //TODO:move to config
		//epoch: time.NewTimer(7 * 24 * time.Hour),//one week //TODO:move to config
		//preColletWait: time.NewTimer(20 * time.Second), //TODO:move to config
		node: node,
	}
}

func (cm *Committee) prePowerNodeCollect()  {
	cm.preColletWait = time.NewTimer(20 * time.Second)
	go func() {
		timerCh := cm.preColletWait.C
		for {
			select {
			case <-timerCh:
				//TODO
			}
		}
	}()
	
	if cm.node.isFirstSharding() {
		
	}
}

func (cm *Committee) firstSharding()  {

}


