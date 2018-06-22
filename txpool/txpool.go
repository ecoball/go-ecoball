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
	"github.com/ecoball/go-ecoball/common/elog"
	"github.com/ecoball/go-ecoball/core/types"
)

var log = elog.NewLogger("TxPool", elog.DebugLog)

type TxPool struct {
	PengdingTx *types.TxsList //Unpackaged list of legitimate transactions
}

//start transaction pool
func Start() (pool *TxPool, err error) {
	//transaction pool
	pool = &TxPool{PengdingTx: types.NewTxsList()}

	//transaction pool actor
	if _, err = NewTxPoolActor(pool); nil != err {
		pool = nil
	}

	return
}

func (p *TxPool) Put(tx *types.Transaction) {
	p.PengdingTx.Push(tx)
}
