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
	"github.com/ecoball/go-ecoball/core/ledgerimpl/ledger"
	"github.com/ecoball/go-ecoball/account"
	"strconv"
	"bytes"
	"github.com/ecoball/go-ecoball/common/config"
)

// in this version, the peers take turns to generate the block
const (
	WAIT_RESPONSE_TIME = 6
)

type State_ababft byte
const (
	Initialization      State_ababft = 0x00
	Primary             State_ababft = 0x01
	Backup              State_ababft = 0x02
)
type Service_ababft struct {
	Actor *Actor_ababft // save the actor object
	pid   *actor.PID
	ledger ledger.Ledger
	account *account.Account
}

type Peer_info struct {
	PublicKey  []byte
	Index      int
}

func Service_ababft_gen(l ledger.Ledger, account *account.Account) (service_ababft *Service_ababft, err error) {
	var pid *actor.PID

	service_ababft = new(Service_ababft)

	actor_ababft := &Actor_ababft{}
	pid, err = Actor_ababft_gen(actor_ababft)
	if err != nil {
		return nil, err
	}
	actor_ababft.pid = pid
	actor_ababft.status = 1
	actor_ababft.service_ababft = service_ababft
	service_ababft.Actor = actor_ababft
	service_ababft.pid = pid
	service_ababft.ledger = l
	service_ababft.account = account

	current_ledger = l
	primary_tag = 0

	return service_ababft, err
}

func (this *Service_ababft) Start() error {
	var err error
	// start the ababft service
	// build the peers list
	Num_peers = len(config.PeerIndex)
	Peers_list = make([]Peer_info, Num_peers)
	for i := 0; i < Num_peers; i++ {
		Peers_list[i].PublicKey =  []byte(config.PeerList[i])
		Peers_list[i].Index, err = strconv.Atoi(config.PeerIndex[i])
		//fmt.Println("peer information:", i, Peers_list[i].PublicKey, Peers_list[i].index)
		if bytes.Equal(Peers_list[i].PublicKey,this.account.PublicKey) {
			Self_index = Peers_list[i].Index
		}
	}
	return err
}

func (this *Service_ababft) Stop() error {
	// stop the ababft
	return nil
}