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

package ledgerimpl

import (
	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/ecoball/go-ecoball/common/event"
	"github.com/ecoball/go-ecoball/common/message"
	"github.com/ecoball/go-ecoball/core/types"
	"reflect"
)

type LedActor struct {
	ledger *LedgerImpl
	pid    *actor.PID //保存自身的pid，用于和其他Actor交互
}

/**
** 创建一个账本的Actor对象
 */
func NewLedgerActor(l *LedActor) (*actor.PID, error) {
	props := actor.FromProducer(func() actor.Actor {
		return l
	})
	pid, err := actor.SpawnNamed(props, "LedgerActor")
	if err != nil {
		return nil, err
	}
	event.RegisterActor(event.ActorLedger, pid)

	return pid, nil
}

func (l *LedActor) SetLedger(ledger *LedgerImpl) {
	l.ledger = ledger
}

/**
** Actor的接收方法，实现了此方法就会实现Actor机制，有消息会调用此函数处理
 */
func (l *LedActor) Receive(ctx actor.Context) {
	switch msg := ctx.Message().(type) {
	case *actor.Started:
	case *actor.Restarting:
	case *types.Transaction: //从交易池发来的确认交易信息的消息类型，需即刻返回结果
		log.Info("Receive Transaction:", msg.Hash.HexString())
		errCode := l.ledger.ChainTx.CheckTransaction(msg)
		log.Info("Response Transaction Check")
		ctx.Sender().Tell(errCode)
	/*case *types.TxsList:
		if err := l.AddBlock(msg); err != nil {
			log.Error(err)
		}*/
	case message.GetTransaction:
		tx, err := l.ledger.ChainTx.GetTransaction(msg.Key)
		if err != nil {
			log.Error("Get Transaction Failed:", err)
		} else {
			ctx.Sender().Tell(tx)
		}
	case *types.Block:
		if err := l.ledger.ChainTx.SaveBlock(msg); err != nil {
			log.Error("save block error:", err)
			break
		}
		if err := event.Send(event.ActorLedger, event.ActorTxPool, msg); err != nil {
			log.Error("send block to tx pool error:", err)
		}
	default:
		log.Warn("unknown type message:", msg, "type", reflect.TypeOf(msg))
	}
}
/*
func (l *LedActor) AddBlock(txList *types.TxsList) error {

	var txs []*types.Transaction
	for _, v := range txList.Txs {
		log.Debug(v.Hash.HexString())
		txs = append(txs, v)
	}
	log.Debug("Receive Txs, then Create a new Block")
	if len(txs) == 0 {
		log.Warn("no transactions now")
		//return nil
	}
	block, err := l.ledger.ChainTx.NewBlock(txs)
	if err != nil {
		return err
	}
	if err := l.ledger.SaveTxBlock(block); err != nil {
		log.Error(err)
	}
	return nil
}
*/