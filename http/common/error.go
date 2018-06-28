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

type Errcode int64

const (
	SUCCESS Errcode = iota
	INVALID_ACCOUNT
	INVALID_PARAMS
	GENERATE_KEY_PAIR_FAILED
	INTERNAL_ERROR
	SAMEDATA
)

var ErrorCodeInfo = map[Errcode]string{
	SUCCESS:                  "success",
	INVALID_ACCOUNT:          "invalid account",
	INVALID_PARAMS:           "invalid arguments",
	GENERATE_KEY_PAIR_FAILED: "generate key pair failed",
	INTERNAL_ERROR:           "internal error",
	SAMEDATA:                 "duplicated data",
}

func (this *Errcode) Info() string {
	desc, exist := ErrorCodeInfo[*this]
	if exist {
		return desc
	} else {
		return "No error description"
	}
}
