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
	"strings"
	"time"

	"github.com/ecoball/go-ecoball/core/types"

	"github.com/ecoball/go-ecoball/account"
	innerCommon "github.com/ecoball/go-ecoball/common"
	"github.com/ecoball/go-ecoball/common/event"
	"github.com/ecoball/go-ecoball/crypto/secp256k1"
	"github.com/ecoball/go-ecoball/http/common"
)

func SetContract(params []interface{}) *common.Response {
	if len(params) < 1 {
		return common.NewResponse(common.INVALID_PARAMS, nil)
	}

	switch {
	case len(params) == 5:
		if errCode, result := handleSetContract(params); errCode != common.SUCCESS {
			return common.NewResponse(errCode, nil)
		} else {
			return common.NewResponse(common.SUCCESS, result)
		}

	default:
		return common.NewResponse(common.INVALID_PARAMS, nil)
	}

	return common.NewResponse(common.SUCCESS, "")
}

func handleSetContract(params []interface{}) (common.Errcode, string) {

	//Get account address
	var (
		code         []byte
		contractName string
		description  string
		author       string
		email        string
		invalid      bool = false
	)

	if v, ok := params[0].(string); ok {
		code = innerCommon.FromHex(v)
	} else {
		invalid = true
	}

	if v, ok := params[1].(string); ok {
		contractName = v
	} else {
		invalid = true
	}

	if v, ok := params[2].(string); ok {
		description = v
	} else {
		invalid = true
	}

	if v, ok := params[3].(string); ok {
		author = v
	} else {
		invalid = true
	}

	if v, ok := params[4].(string); ok {
		email = v
	} else {
		invalid = true
	}

	if invalid {
		return common.INVALID_PARAMS, ""
	}

	//time
	time := time.Now().Unix()

	//generate key pair
	keyData, err := secp256k1.NewECDSAPrivateKey()
	if err != nil {
		return common.GENERATE_KEY_PAIR_FAILED, ""
	}

	public, err := secp256k1.FromECDSAPub(&keyData.PublicKey)
	if err != nil {
		return common.GENERATE_KEY_PAIR_FAILED, ""
	}

	//generate address
	address := account.AddressFromPubKey(public)

	//from address
	from := account.AddressFromPubKey(common.Account.PublicKey)

	transaction, err := types.NewDeployContract(from, address, types.VmWasm, author,
		contractName, email, description, code, 0, time)
	if nil != err {
		return common.INVALID_PARAMS, ""
	}

	/*err = transaction.SetSignature(&common.Account)
	if err != nil {
		return common.INVALID_ACCOUNT, ""
	}*/

	//send to txpool
	err = event.Send(event.ActorNil, event.ActorTxPool, transaction)
	if nil != err {
		return common.INTERNAL_ERROR, ""
	}

	return common.SUCCESS, address.HexString()
}

func InvokeContract(params []interface{}) *common.Response {
	if len(params) < 1 {
		return common.NewResponse(common.INVALID_PARAMS, nil)
	}

	switch {
	case len(params) == 3:
		if errCode := handleInvokeContract(params); errCode != common.SUCCESS {
			return common.NewResponse(errCode, nil)
		}

	default:
		return common.NewResponse(common.INVALID_PARAMS, nil)
	}

	return common.NewResponse(common.SUCCESS, "")
}

func handleInvokeContract(params []interface{}) common.Errcode {
	var (
		contractAddress string
		contractMethod  string
		contractParam   string
		parameters      []string
		invalid         bool = false
	)

	if v, ok := params[0].(string); ok {
		contractAddress = v
	} else {
		invalid = true
	}

	if v, ok := params[1].(string); ok {
		contractMethod = v
	} else {
		invalid = true
	}

	if v, ok := params[2].(string); ok {
		contractParam = v
	} else {
		invalid = true
	}

	if "" != contractParam {
		parameters = strings.Split(contractParam, " ")
	}

	if invalid {
		return common.INVALID_PARAMS
	}

	//from address
	from := account.AddressFromPubKey(common.Account.PublicKey)

	//contract address
	address := innerCommon.NewAddress(innerCommon.CopyBytes(innerCommon.FromHex(contractAddress)))

	//time
	time := time.Now().Unix()

	transaction, err := types.NewInvokeContract(from, address, types.VmWasm, contractMethod, parameters, 0, time)
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
