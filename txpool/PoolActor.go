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

package txpool

import (
	"errors"
	"reflect"
	"time"

	"github.com/ecoball/go-ecoball/common"
	"github.com/ecoball/go-ecoball/common/message"
	"github.com/ecoball/go-ecoball/crypto/secp256k1"

	"github.com/AsynkronIT/protoactor-go/actor"
	innerError "github.com/ecoball/go-ecoball/common/errors"
	"github.com/ecoball/go-ecoball/common/event"
	"github.com/ecoball/go-ecoball/core/types"
)

type PoolActor struct {
	txPool *TxPool
}

func NewTxPoolActor(pool *TxPool) (pid *actor.PID, err error) {
	props := actor.FromProducer(func() actor.Actor {
		return &PoolActor{txPool: pool}
	})

	if pid, err = actor.SpawnNamed(props, "TxPoolActor"); nil != err {
		return nil, err
	}

	event.RegisterActor(event.ActorTxPool, pid)

	return
}

func (l *PoolActor) Receive(ctx actor.Context) {
	switch msg := ctx.Message().(type) {
	case *actor.Started:
	case *actor.Restarting:
	case *types.Transaction:
		log.Info("receive tx:", msg.Hash.HexString())
		l.handleTransaction(msg)
	case message.GetTxs:
		log.Debug("Ledger request txs")
		txs := types.NewTxsList()
		txs.Copy(l.txPool.PengdingTx)
		ctx.Sender().Tell(txs)
	case *types.Block:
		log.Debug("new block delete transactions")
		l.handleNewBlock(msg)
	default:
		log.Warn("unknown type message:", msg, "type", reflect.TypeOf(msg))
	}
}

//Determine whether a transaction already exists
func (this *PoolActor) isSameTransaction(hash common.Hash) bool {
	if tr := this.txPool.PengdingTx.Same(hash); tr {
		return true
	}

	return false
}

func (this *PoolActor) handleTransaction(tx *types.Transaction) error {
	if exist := this.isSameTransaction(tx.Hash); exist {
		log.Warn("transaction already in the txn pool")
		return errors.New("transaction already in the txn pool: " + tx.Hash.HexString())
	}

	//check transaction signatures
	data := tx.Hash.Bytes()
	for _, v := range tx.Signatures {
		if hasSign, err := secp256k1.Verify(data, v.SigData, v.PubKey); nil != err || !hasSign {
			log.Warn("check transaction signatures failed:" + tx.Hash.HexString())
			return errors.New("check transaction signatures fail:" + tx.Hash.HexString())
		}
	}

	//Send the transaction to ledger to Check legitimacy
	res, err := event.SendSync(event.ActorLedger, tx, time.Second*2)
	if nil != err {
		log.Warn("send message to ledger actor error: ", tx.Hash.HexString())
		return errors.New("send message to ledger actor error: " + tx.Hash.HexString())
	} else if v := res.(innerError.ErrCode); v != innerError.ErrNoError {
		log.Warn(tx.Hash.HexString(), " Check legitimacy failed: ", v.ErrorInfo())
		return errors.New(tx.Hash.HexString() + " Check legitimacy failed: " + v.ErrorInfo())
	}

	//Verify by adding to the transaction pool
	this.txPool.PengdingTx.Push(tx)

	//Broadcast transactions on p2p
	if err := event.Send(event.ActorNil, event.ActorP2P, tx); nil != err {
		log.Warn("broadcast transaction failed:" + tx.Hash.HexString())
		return errors.New("broadcast transaction failed:" + tx.Hash.HexString())
	}

	return nil
}

func (this *PoolActor) handleNewBlock(block *types.Block) {
	for _, v := range block.Transactions {
		this.txPool.PengdingTx.Delete(v.Hash)
	}
}
