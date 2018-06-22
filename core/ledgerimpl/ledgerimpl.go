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
	"github.com/ecoball/go-ecoball/core/ledgerimpl/ledger"
	"github.com/ecoball/go-ecoball/core/ledgerimpl/transaction"
	"github.com/ecoball/go-ecoball/core/state"
	"github.com/ecoball/go-ecoball/core/types"
)

var log = elog.NewLogger("LedgerImpl", elog.DebugLog)

type LedgerImpl struct {
	ChainTx *transaction.ChainTx
	//ChainCt *ChainContract
	//ChainAc *account.ChainAccount

}

func NewLedger(path string) (l ledger.Ledger, err error) {
	ll := new(LedgerImpl)
	ll.ChainTx, err = transaction.NewTransactionChain(path + "/Transaction")
	if err != nil {
		return nil, err
	}

	actor := &LedActor{ledger: ll}
	actor.pid, err = NewLedgerActor(actor)
	if err != nil {
		return nil, err
	}

	return ll, nil
}

func (l *LedgerImpl) GetAccountBalance(addr common.Address) (uint64, error) {
	value, err := l.ChainTx.StateDB.GetBalance(addr, state.AbaToken, []byte("aba"))
	if err != nil {
		return 0, err
	}
	return value.Uint64(), nil
}

func (l *LedgerImpl) GetContractInfo(key []byte) ([]byte, error) {
	return l.ChainTx.TxsStore.Get(key)
}

func (l *LedgerImpl) AddAccountBalance(addr common.Address, value uint64) error {
	return l.ChainTx.AddAccountBalance(addr, value)
}

func (l *LedgerImpl) SubAccountBalance(addr common.Address, value uint64) error {
	return l.ChainTx.SubAccountBalance(addr, value)
}

func (l *LedgerImpl) NewTxBlock(txs []*types.Transaction, consensusData types.ConsensusData) (*types.Block, error) {
	return l.ChainTx.NewBlock(l, txs, consensusData)
}

func (l *LedgerImpl) GetCurrentHeader() *types.Header {
	return l.ChainTx.CurrentHeader
}

func (l *LedgerImpl) GetCurrentHeight() uint64 {
	return l.ChainTx.CurrentHeader.Height
}

func (l *LedgerImpl) SaveTxBlock(block *types.Block) error {
	if err := l.ChainTx.SaveBlock(block); err != nil {
		return err
	}
	return nil
}

func (l *LedgerImpl) GetTxBlock(hash common.Hash) (*types.Block, error) {
	return l.ChainTx.GetBlock(hash)
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