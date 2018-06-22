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

package common

import (
	"encoding/json"

	"github.com/ecoball/go-ecoball/account"
)

var Account account.Account

type Response struct {
	errCode Errcode
	desc    string
	result  interface{}
}

func NewResponse(code Errcode, info interface{}) *Response {
	resp := Response{
		errCode: code,
		desc:    code.info(),
		result:  info,
	}

	return &resp
}

func (this *Response) Serialize() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"errorCode": int64(this.errCode),
		"desc":      this.desc,
		"result":    this.result,
	})
}
