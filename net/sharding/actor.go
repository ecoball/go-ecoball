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
	"github.com/AsynkronIT/protoactor-go/actor"
	eactor "github.com/ecoball/go-ecoball/common/event"
)

type ShardingActor struct {
	props    *actor.Props
}

func NewShardingActor() *ShardingActor {
	return &ShardingActor{
	}
}

func (this *ShardingActor) Start() (*actor.PID, error) {
	this.props = actor.FromProducer(func() actor.Actor { return this })
	netPid, err := actor.SpawnNamed(this.props, "net")
	eactor.RegisterActor(eactor.ActorP2P, netPid)
	return netPid, err
}

func (this *ShardingActor) Receive(ctx actor.Context) {
	//var buffer []byte
	//var msgType uint32
	msg := ctx.Message()
	switch msg.(type) {
	case *actor.Started:
	//	log.Debug("NetActor started")
	default:
	//	log.Error("Error Xmit message ", reflect.TypeOf(ctx.Message()))
	}

	//log.Debug("Actor receive msg ", reflect.TypeOf(ctx.Message()))
}

