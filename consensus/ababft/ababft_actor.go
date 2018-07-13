package ababft

import (
	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/ecoball/go-ecoball/common/elog"
	"github.com/ecoball/go-ecoball/core/ledgerimpl/ledger"
	"github.com/ecoball/go-ecoball/common/event"
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

var primary_tag int

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
	// var err error
	// log.Debug("ababft service receives the message")

	// deal with the message
	switch msg := ctx.Message().(type) {
	default :
		log.Debug(msg)
		log.Warn("unknown message")
	}
}
