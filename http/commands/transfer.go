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

package commands

import (
	"math/big"
	"time"

	"github.com/ecoball/go-ecoball/core/types"

	inner "github.com/ecoball/go-ecoball/common"
	"github.com/ecoball/go-ecoball/common/event"
	"github.com/ecoball/go-ecoball/http/common"
)

//transfer handle
func Transfer(params []interface{}) *common.Response {
	if len(params) < 1 {
		log.Error("invalid arguments")
		return common.NewResponse(common.INVALID_PARAMS, nil)
	}

	switch {
	case len(params) == 3:
		if errCode := handleTransfer(params); errCode != common.SUCCESS {
			log.Error(errCode.Info())
			return common.NewResponse(errCode, nil)
		}

	default:
		return common.NewResponse(common.INVALID_PARAMS, nil)
	}

	return common.NewResponse(common.SUCCESS, "")
}

func handleTransfer(params []interface{}) common.Errcode {
	var (
		from    string
		to      string
		value   *big.Int
		invalid bool = false
	)

	if v, ok := params[0].(string); ok {
		from = v
	} else {
		invalid = true
	}

	if v, ok := params[1].(string); ok {
		to = v
	} else {
		invalid = true
	}

	if v, ok := params[2].(float64); ok {
		value = big.NewInt(int64(v))
	} else {
		invalid = true
	}

	if invalid {
		return common.INVALID_PARAMS
	}

	//time
	time := time.Now().Unix()

	transaction, err := types.NewTransfer(inner.NameToIndex(from), inner.NameToIndex(to), "owner", value, 0, time)
	if nil != err {
		return common.INVALID_PARAMS
	}

	/*err = transaction.SetSignature(&common.Account)
	if err != nil {
		return common.INVALID_ACCOUNT
	}*/

	//send to txpool
	err = event.Send(event.ActorNil, event.ActorTxPool, transaction)
	if nil != err {
		return common.INTERNAL_ERROR
	}

	return common.SUCCESS
}
