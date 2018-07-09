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

package types

import (
	"fmt"
	"github.com/ecoball/go-ecoball/common"
	"sync"
)

/**
** 交易列表结构体
 */
type TxsList struct {
	Txs map[common.Hash]*Transaction

	mux sync.RWMutex
}

func NewTxsList() *TxsList {
	return &TxsList{Txs: make(map[common.Hash]*Transaction, 0)}
}

//删除一笔交易
func (t *TxsList) Delete(hash common.Hash) {
	t.mux.Lock()
	defer t.mux.Unlock()

	delete(t.Txs, hash)
}

//检查是否有重复的数据
func (t *TxsList) Same(hash common.Hash) bool {
	t.mux.RLock()
	defer t.mux.RUnlock()

	if value := t.Txs[hash]; nil != value {
		return true
	}

	return false
}

/**
** 入栈操作
 */
func (t *TxsList) Push(tx *Transaction) {
	t.mux.Lock()
	defer t.mux.Unlock()
	if _, ok := t.Txs[tx.Hash]; ok {
		return
	}
	t.Txs[tx.Hash] = tx
}

/**
** 出栈操作，取出后立即从列表移除
 */
func (t *TxsList) Pull() (tx *Transaction) {
	t.mux.Lock()
	defer t.mux.Unlock()
	if len(t.Txs) == 0 {
		return nil
	}
	var tt *Transaction
	for k, v := range t.Txs {
		tt = v
		delete(t.Txs, k)
		return tt
	}
	return nil
}

func (t *TxsList) Copy(txs *TxsList) {
	txs.mux.RLock()
	defer txs.mux.RUnlock()
	for k, v := range txs.Txs {
		t.Txs[k] = v
	}
}

func (t *TxsList) Show() {
	t.mux.RLock()
	defer t.mux.RUnlock()
	for _, v := range t.Txs {
		fmt.Println("Version:", v.Version)
		fmt.Println("From:", v.From)
		fmt.Println("Hash:", v.Hash.HexString())
	}
}
