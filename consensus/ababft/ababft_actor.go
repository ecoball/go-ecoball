package ababft

import (
	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/ecoball/go-ecoball/common/elog"
	"gitlab.quachain.net/aba/aba/common/config"
	"github.com/ecoball/go-ecoball/core/ledgerimpl/ledger"
	"github.com/ecoball/go-ecoball/account"
	"strconv"
	"bytes"
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
var Self_index int
var current_round_num int
var current_height_num int
var current_ledger ledger.Ledger

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

	// todo
	// need to modify the config.go......
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