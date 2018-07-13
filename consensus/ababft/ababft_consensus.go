package ababft

import (
	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/ecoball/go-ecoball/core/ledgerimpl/ledger"
	"github.com/ecoball/go-ecoball/account"
	"github.com/ecoball/go-ecoball/common/event"
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


