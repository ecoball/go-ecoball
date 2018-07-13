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