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

package solo

import (
	"github.com/ecoball/go-ecoball/common/elog"
	"github.com/ecoball/go-ecoball/common/event"
	"github.com/ecoball/go-ecoball/common/message"
	"github.com/ecoball/go-ecoball/core/ledgerimpl/ledger"
	"github.com/ecoball/go-ecoball/core/types"
	"time"
)

var log = elog.NewLogger("Solo", elog.NoticeLog)

type Solo struct {
	ledger ledger.Ledger
}

func NewSoloConsensusServer(l ledger.Ledger) (*Solo, error) {
	return &Solo{ledger: l}, nil
}

func (s *Solo) Start() error {
	t := time.NewTimer(time.Second * 5)
	conData := types.ConsensusData{Type: types.ConSolo, Payload: &types.SoloData{}}

	go func() {
		for {
			select {
			case <-t.C:
				log.Debug("Request transactions from tx pool")
				value, err := event.SendSync(event.ActorTxPool, message.GetTxs{}, time.Second*1)
				if err != nil {
					log.Error("Solo Consensus error:", err)
					continue
				}
				txList, ok := value.(*types.TxsList)
				if !ok {
					log.Error("The format of value error [solo]")
					continue
				}
				var txs []*types.Transaction
				for _, v := range txList.Txs {
					txs = append(txs, v)
				}
				block, err := s.ledger.NewTxBlock(txs, conData)
				if err != nil {
					log.Error("new block error:", err)
					continue
				}
				if err := s.ledger.SaveTxBlock(block); err != nil {
					log.Error("save block error:", err)
					continue
				}

				t.Reset(time.Second * 5)
			}
		}
	}()
	return nil
}

