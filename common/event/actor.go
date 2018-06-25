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

package event

import (
	"fmt"
	"sync"
	"time"

	"errors"

	"github.com/AsynkronIT/protoactor-go/actor"
)

type ActorIndex int

const (
	ActorNil ActorIndex = iota
	ActorTxPool
	ActorInputTx
	ActorValidator
	ActorP2P
	ActorHttp
	ActorConsensus
	ActorLedger
	ActorValidation
	ActorNetRpc
	maxActorNumber
)

type actors struct {
	mux  sync.Mutex
	list map[ActorIndex]*actor.PID
}

var actorList = actors{sync.Mutex{}, make(map[ActorIndex]*actor.PID)}

func (a ActorIndex) String() string {
	switch a {
	case ActorTxPool:
		return "tx pool"
	case ActorInputTx:
		return "input tx"
	case ActorValidator:
		return "validator"
	case ActorP2P:
		return "p2p"
	case ActorHttp:
		return "http"
	case ActorConsensus:
		return "consensus"
	case ActorLedger:
		return "ledger"
	case ActorValidation:
		return "validation"
	default:
		return "unknown actor index"
	}
}

/**
* 各模块在创建actor时先注册，然后使用GetActor来获取其他模块的actor
 */
func RegisterActor(index ActorIndex, pid *actor.PID) error {
	if index <= ActorNil || index > maxActorNumber {
		return errors.New("invalid index since too big or too little")
	}
	actorList.mux.Lock()
	defer actorList.mux.Unlock()
	if _, ok := actorList.list[index]; ok {
		return errors.New("this actor is existed")
	}
	actorList.list[index] = pid
	return nil
}

func GetActor(index ActorIndex) (*actor.PID, error) {
	actorList.mux.Lock()
	defer actorList.mux.Unlock()
	a, ok := actorList.list[index]
	if !ok {
		return nil, errors.New(fmt.Sprintf("not found this actor:%s", index.String()))
	}
	return a, nil
}

func DelActor(index ActorIndex) {
	actorList.mux.Lock()
	defer actorList.mux.Unlock()
	if _, ok := actorList.list[index]; ok {
		delete(actorList.list, index)
	}
}

/**
** 使用异步方式发送消息，sender是调用者的Actor index，receiver是接收者的Actor index，
** 如果希望对方返回数据，那么就需要带上自身的sender index，否则可以设置为0即ActorNil
 */
func Send(sender, receiver ActorIndex, msg interface{}) error {
	if sender != ActorNil {
		s, err := GetActor(sender)
		if err != nil {
			return err
		}
		r, err := GetActor(receiver)
		if err != nil {
			return err
		}
		r.Request(msg, s)
	} else {
		r, err := GetActor(receiver)
		if err != nil {
			return err
		}
		r.Tell(msg)
	}

	return nil
}

/**
** 同步发送消息方式，需要带上时间参数
 */
func SendSync(receiver ActorIndex, msg interface{}, timeout time.Duration) (interface{}, error) {
	r, err := GetActor(receiver)
	if err != nil {
		return nil, err
	}
	value := r.RequestFuture(msg, timeout)
	res, err := value.Result()
	if err != nil {
		return nil, err
	}
	return res, nil
}

/**
** Send msg to multiple actors
** pub -- the sender actor
** sub -- the receiver actor
 */
func Publish(pub ActorIndex, msg interface{}, sub ...ActorIndex) error {
	for _, s := range sub {
		if err := Send(pub, s, msg); err != nil {
			return err
		}
	}
	return nil
}
