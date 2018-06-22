package event_test

import (
	"fmt"
	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/ecoball/go-ecoball/common/event"
	"testing"
	"time"
)

type Data struct {
	val int
}

func TestActorRegister(t *testing.T) {
	props := actor.FromFunc(func(c actor.Context) {
		switch msg := c.Message().(type) {
		case int32:
			fmt.Println(msg)
		case *Data:
			fmt.Println(msg.val)
		default:
			fmt.Println("unkown type")
		}
	})
	actorA, _ := actor.SpawnNamed(props, "actorA")
	actorB, _ := actor.SpawnNamed(props, "actorB")

	if err := event.RegisterActor(event.ActorTxPool, actorA); err != nil {
		t.Fatal(err)
	}
	if err := event.RegisterActor(event.ActorLedger, actorB); err != nil {
		t.Fatal(err)
	}

	actorTxPool, _ := event.GetActor(event.ActorTxPool)
	actorLedger, _ := event.GetActor(event.ActorLedger)
	var i int32 = 1000
	actorTxPool.Request(i, actorLedger)
	actorTxPool.Request(&Data{val: 99}, actorLedger)
	time.Sleep(1 * time.Second)
}
