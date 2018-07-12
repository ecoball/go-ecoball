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
	"github.com/ecoball/go-ecoball/common"
	"github.com/ecoball/go-ecoball/common/config"
	"github.com/ecoball/go-ecoball/common/elog"
	"github.com/ecoball/go-ecoball/common/event"
	"github.com/ecoball/go-ecoball/common/message"
	"github.com/ecoball/go-ecoball/consensus/dpos"
	"github.com/ecoball/go-ecoball/core/ledgerimpl/ledger"
	"github.com/ecoball/go-ecoball/core/ledgerimpl/transaction"
	"github.com/ecoball/go-ecoball/core/state"
	"github.com/ecoball/go-ecoball/core/types"
	"time"
)

var log = elog.NewLogger("LedgerImpl", elog.DebugLog)

type LedgerImpl struct {
	ChainTx *transaction.ChainTx
	//ChainCt *ChainContract
	//ChainAc *account.ChainAccount

	//TODO, start
	bc   *dpos.Blockchain
	dpos *dpos.DposService
	//TODO, end
}

func NewLedger(path string) (l ledger.Ledger, err error) {
	ll := new(LedgerImpl)
	ll.ChainTx, err = transaction.NewTransactionChain(path+"/Transaction", ll)
	if err != nil {
		return nil, err
	}
	if err := ll.ChainTx.GenesesBlockInit(); err != nil {
		return nil, err
	}
	//TODO
	if config.ConsensusAlgorithm == "DPOS" {
		ll.bc, err = dpos.NewBlockChain(ll.ChainTx)
		ll.dpos, err = dpos.NewDposService()
		if err != nil {
			log.Debug("Init NewBlockChain error")
			return nil, err
		}
		log.Debug("DPOS setup")
		ll.bc.Setup(ll.dpos)
		ll.dpos.Setup(ll.bc, ll)
	}

	actor := &LedActor{ledger: ll}
	actor.pid, err = NewLedgerActor(actor)
	if err != nil {
		return nil, err
	}

	return ll, nil
}

func (l *LedgerImpl) Start() {
	//TODO start
	if config.ConsensusAlgorithm == "DPOS" {
		log.Debug("Ledger start DPOS")
		l.bc.Start()
		l.dpos.Start()
		//TODO end
	} else {
		t := time.NewTimer(time.Second * 10)

		go func() {
			for {
				select {
				case <-t.C:
					log.Debug("Request a new block")
					event.Send(event.ActorLedger, event.ActorTxPool, message.GetTxs{})
					t.Reset(time.Second * 10)
				}
			}
		}()
	}
}

func (l *LedgerImpl) StateDB() *state.State {
	return l.ChainTx.StateDB
}
func (l *LedgerImpl) ResetStateDB(hash common.Hash) error {
	return l.ChainTx.ResetStateDB(hash)
}

func (l *LedgerImpl) AccountGet(index common.AccountName) (*state.Account, error) {
	return l.ChainTx.StateDB.GetAccountByName(index)
}
func (l *LedgerImpl) AccountAdd(index common.AccountName, addr common.Address) (*state.Account, error) {
	return l.ChainTx.AccountAdd(index, addr)
}
func (l *LedgerImpl) AddPermission(index common.AccountName, perm state.Permission) error {
	return l.ChainTx.AddPermission(index, perm)
}
func (l *LedgerImpl) FindPermission(index common.AccountName, name string) (string, error) {
	return l.ChainTx.FindPermission(index, name)
}
func (l *LedgerImpl) AccountGetBalance(index common.AccountName, token string) (uint64, error) {
	value, err := l.ChainTx.AccountGetBalance(index, token)
	if err != nil {
		return 0, err
	}
	return value.Uint64(), nil
}
func (l *LedgerImpl) AccountAddBalance(index common.AccountName, token string, value uint64) error {
	return l.ChainTx.AccountAddBalance(index, token, value)
}
func (l *LedgerImpl) AccountSubBalance(index common.AccountName, token string, value uint64) error {
	return l.ChainTx.AccountSubBalance(index, token, value)
}
func (l *LedgerImpl) TokenCreate(index common.AccountName, token string, maximum uint64) error {
	return l.ChainTx.AccountAddBalance(index, token, maximum)
}
func (l *LedgerImpl) TokenIsExisted(token string) bool {
	return l.ChainTx.TokenExisted(token)
}

func (l *LedgerImpl) GetTxBlockByHeight(height uint64) (*types.Block, error) {
	return l.ChainTx.GetBlockByHeight(height)
}
func (l *LedgerImpl) GetCurrentHeader() *types.Header {
	return l.ChainTx.CurrentHeader
}
func (l *LedgerImpl) GetCurrentHeight() uint64 {
	return l.ChainTx.CurrentHeader.Height
}

func (l *LedgerImpl) GetTxBlock(hash common.Hash) (*types.Block, error) {
	return l.ChainTx.GetBlock(hash)
}
func (l *LedgerImpl) NewTxBlock(txs []*types.Transaction, consensusData types.ConsensusData) (*types.Block, error) {
	return l.ChainTx.NewBlock(l, txs, consensusData)
}
func (l *LedgerImpl) VerifyTxBlock(block *types.Block) error {
	return l.ChainTx.VerifyTxBlock(block)
}
func (l *LedgerImpl) CheckTransaction(tx *types.Transaction) error {
	if err := l.ChainTx.CheckTransaction(tx); err != nil {
		return err
	}
	//if err := l.ChainAc.CheckTransaction(tx); err != nil {
	//	return err
	//}
	return nil
}
func (l *LedgerImpl) SaveTxBlock(block *types.Block) error {
	if err := l.ChainTx.SaveBlock(block); err != nil {
		return err
	}
	return nil
}

