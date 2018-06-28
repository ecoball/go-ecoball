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
	"strconv"

	"github.com/ecoball/go-ecoball/http/common"
)

//query
func Query(params []interface{}) *common.Response {
	if len(params) < 1 {
		return common.NewResponse(common.INVALID_PARAMS, nil)
	}

	switch params[0].(type) {
	case string:
		if params[0].(string) == string("balance") {
			address := params[1].(string)

			//query balance
			balance, errCode := QueryBalance(address)
			if errCode != common.SUCCESS {
				log.Error(errCode.Info())
				return common.NewResponse(errCode, nil)
			}

			//result
			return common.NewResponse(common.SUCCESS, address+": "+strconv.FormatInt(balance, 10))
		}
	default:
		return common.NewResponse(common.INVALID_PARAMS, nil)
	}

	return common.NewResponse(common.SUCCESS, "")
}

func QueryBalance(address string) (int64, common.Errcode) {
	return 100, common.SUCCESS
}
