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
	"errors"
	"fmt"
	"github.com/ecoball/go-ecoball/account"
	"github.com/ecoball/go-ecoball/common"
	"github.com/ecoball/go-ecoball/core/pb"
	"time"
)

type InvokeInfo struct {
	TypeVm VmType
	Method []byte
	Param  []string
}

func NewInvokeContract(from, addr common.Address, vm VmType, method string, param []string, nonce uint64, time int64) (*Transaction, error) {
	invoke := &InvokeInfo{TypeVm: vm, Method: []byte(method), Param: param}
	trans, err := NewTransaction(TxInvoke, from, addr, invoke, nonce, time)
	if err != nil {
		return nil, err
	}
	return trans, nil
}

func (i InvokeInfo) GetObject() interface{} {
	return i
}

func (i *InvokeInfo) Show() {
	fmt.Println("\tTypeVm        :", i.TypeVm)
	fmt.Println("\tMethod        :", string(i.Method))
	fmt.Println("\tParam Num     :", len(i.Param))
	for _, v := range i.Param {
		fmt.Println("\tParam         :", v)
	}
	fmt.Println("\t---------------------------")
}

func (i *InvokeInfo) Serialize() ([]byte, error) {
	var param []*pb.ParamData
	for _, v := range i.Param {
		p := pb.ParamData{Param:[]byte(v)}
		param = append(param, &p)
	}

	p := &pb.InvokeInfo{
		TypeVm: uint32(i.TypeVm),
		Method: i.Method,
		Param:  param,
	}
	b, err := p.Marshal()
	if err != nil {
		return nil, err
	}
	return b, nil
}

func (i *InvokeInfo) Deserialize(data []byte) error {
	if len(data) == 0 {
		return errors.New("input data's length is zero")
	}
	var invoke pb.InvokeInfo
	if err := invoke.Unmarshal(data); err != nil {
		return err
	}
	i.TypeVm = VmType(invoke.TypeVm)
	i.Method = common.CopyBytes(invoke.Method)
	for _, v := range invoke.Param {
		p := string(v.Param)
		i.Param = append(i.Param, p)
	}

	return nil
}

func NewTestInvoke() *Transaction {
	from := common.NewAddress(common.FromHex("01b1a6569a557eafcccc71e0d02461fd4b601aea"))
	addr := common.NewAddress(common.FromHex("01ca5cdd56d99a0023166b337ffc7fd0d2c42330"))
	invoke, err := NewInvokeContract(from, addr, VmWasm, "main", []string{"-run"}, 0, time.Now().Unix())
	if err != nil {
		panic(err)
		return nil
	}
	acc, err := account.NewAccount(0)
	if err != nil {
		panic(err)
	}
	if err := invoke.SetSignature(&acc); err != nil {
		panic(err)
	}
	return invoke
}
