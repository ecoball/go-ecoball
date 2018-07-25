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
	"github.com/ecoball/go-ecoball/common/elog"
	"github.com/ecoball/go-ecoball/common/event"
	"github.com/ecoball/go-ecoball/common/message"
	"github.com/ecoball/go-ecoball/core/ledgerimpl/ledger"
	"github.com/ecoball/go-ecoball/core/ledgerimpl/transaction"
	"github.com/ecoball/go-ecoball/core/state"
	"github.com/ecoball/go-ecoball/core/types"
	"math/big"
	"time"
)

var log = elog.NewLogger("LedgerImpl", elog.DebugLog)

type LedgerImpl struct {
	ChainTx *transaction.ChainTx
	//ChainCt *ChainContract
	//ChainAc *account.ChainAccount
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

	actor := &LedActor{ledger: ll}
	actor.pid, err = NewLedgerActor(actor)
	if err != nil {
		return nil, err
	}

	return ll, nil
}

func (l *LedgerImpl) Start() {

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

func (l *LedgerImpl) NewTxBlock(txs []*types.Transaction, consensusData types.ConsensusData) (*types.Block, error) {
	return l.ChainTx.NewBlock(l, txs, consensusData)
}
func (l *LedgerImpl) GetTxBlock(hash common.Hash) (*types.Block, error) {
	return l.ChainTx.GetBlock(hash)
}
func (l *LedgerImpl) SaveTxBlock(block *types.Block) error {
	if err := l.ChainTx.SaveBlock(block); err != nil {
		return err
	}
	return nil
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
func (l *LedgerImpl) GetChainTx() ledger.ChainInterface {
	return l.ChainTx
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

func (l *LedgerImpl) AccountGet(index common.AccountName) (*state.Account, error) {
	return l.ChainTx.StateDB.GetAccountByName(index)
}
func (l *LedgerImpl) AccountAdd(index common.AccountName, addr common.Address) (*state.Account, error) {
	return l.ChainTx.StateDB.AddAccount(index, addr)
}
//func (l *LedgerImpl) SetResourceLimits(from, to common.AccountName, cpu, net float32) error {
//	return l.ChainTx.StateDB.SetResourceLimits(from, to, cpu, net)
//}
func (l *LedgerImpl) StoreSet(index common.AccountName, key, value []byte) (err error) {
	return l.ChainTx.StateDB.StoreSet(index, key, value)
}
func (l *LedgerImpl) StoreGet(index common.AccountName, key []byte) (value []byte, err error) {
	return l.ChainTx.StateDB.StoreGet(index, key)
}
func (l *LedgerImpl) SetContract(index common.AccountName, t types.VmType, des, code []byte) error {
	return l.ChainTx.StateDB.SetContract(index, t, des, code)
}
func (l *LedgerImpl) GetContract(index common.AccountName) (*types.DeployInfo, error) {
	return l.ChainTx.StateDB.GetContract(index)
}
func (l *LedgerImpl) AddPermission(index common.AccountName, perm state.Permission) error {
	return l.ChainTx.StateDB.AddPermission(index, perm)
}
func (l *LedgerImpl) FindPermission(index common.AccountName, name string) (string, error) {
	return l.ChainTx.StateDB.FindPermission(index, name)
}
func (l *LedgerImpl) CheckPermission(index common.AccountName, name string, sig []common.Signature) error {
	return l.ChainTx.StateDB.CheckPermission(index, name, sig)
}
func (l *LedgerImpl) AccountGetBalance(index common.AccountName, token string) (uint64, error) {
	value, err := l.ChainTx.StateDB.AccountGetBalance(index, token)
	if err != nil {
		return 0, err
	}
	return value.Uint64(), nil
}
func (l *LedgerImpl) AccountAddBalance(index common.AccountName, token string, value uint64) error {
	return l.ChainTx.StateDB.AccountAddBalance(index, token, new(big.Int).SetUint64(value))
}
func (l *LedgerImpl) AccountSubBalance(index common.AccountName, token string, value uint64) error {
	return l.ChainTx.StateDB.AccountSubBalance(index, token, new(big.Int).SetUint64(value))
}
func (l *LedgerImpl) TokenCreate(index common.AccountName, token string, maximum uint64) error {
	return l.ChainTx.StateDB.AccountAddBalance(index, token, new(big.Int).SetUint64(maximum))
}
func (l *LedgerImpl) TokenIsExisted(token string) bool {
	return l.ChainTx.StateDB.TokenExisted(token)
}
func (l *LedgerImpl) StateDB() *state.State {
	return l.ChainTx.StateDB
}
func (l *LedgerImpl) ResetStateDB(hash common.Hash) error {
	return l.ChainTx.StateDB.Reset(hash)
}
