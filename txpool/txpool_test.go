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
package txpool_test

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/ecoball/go-ecoball/common"
	"github.com/ecoball/go-ecoball/common/event"
	"github.com/ecoball/go-ecoball/common/message"
	"github.com/ecoball/go-ecoball/core/types"
	"github.com/ecoball/go-ecoball/txpool"
)

func TestTxPool(t *testing.T) {
	txPool, err := txpool.Start()
	if err != nil {
		t.Fatal(err)
	}

	txPool.Put(newTx(t))
	txPool.PengdingTx.Show()

	var pid *actor.PID
	pid, err = event.GetActor(event.ActorTxPool)
	if err != nil {
		t.Fatal(err)
	}

	pid.Tell(message.GetTxs{})
	time.Sleep(time.Duration(5) * time.Second)
}

func newTx(t *testing.T) *types.Transaction {
	fromUser, _ := hex.DecodeString("01b1a6569a557eafcccc71e0d02461fd4b601aea")
	toUser, _ := hex.DecodeString("01ca5cdd56d99a0023166b337ffc7fd0d2c42330")
	from := common.NewAddress(fromUser)
	to := common.NewAddress(toUser)
	value := big.NewInt(100)
	timeStamp := time.Now().Unix()
	fmt.Println(timeStamp)
	//生成结构体，会自动计算哈希值
	tx, err := types.NewTransfer(from, to, value, 0, timeStamp)
	if err != nil {
		t.Fatal(err)
	}
	return tx
}
