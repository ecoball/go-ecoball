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

package wasmservice

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/ecoball/go-ecoball/common"
	"github.com/ecoball/go-ecoball/common/elog"
	"github.com/ecoball/go-ecoball/core/ledgerimpl/ledger"
	"github.com/ecoball/go-ecoball/core/types"
	"github.com/ecoball/go-ecoball/vm/wasmvm/exec"
	"github.com/ecoball/go-ecoball/vm/wasmvm/util"
	"github.com/ecoball/go-ecoball/vm/wasmvm/validate"
	"github.com/ecoball/go-ecoball/vm/wasmvm/wasm"
	"io/ioutil"
	"os"
)

var log = elog.NewLogger("wasm", elog.NoticeLog)

type WasmService struct {
	ledger ledger.Ledger
	tx     *types.Transaction
	Code   []byte
	Args   []uint64
	Method string
}

func NewWasmService(ledger ledger.Ledger, method string, code []byte, arg []uint64) (*WasmService, error) {
	if len(code) == 0 {
		return nil, errors.New("code is nil")
	}
	ws := &WasmService{
		ledger: ledger,
		Code:   code,
		Args:   arg,
		Method: method,
	}
	ws.RegisterApi()
	return ws, nil
}

func ReadWasm(file string) ([]byte, error) {
	raw, err := ioutil.ReadFile(file)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	return raw, nil
}

func (ws *WasmService) Execute() []byte {
	bf := bytes.NewBuffer(ws.Code)
	m, err := wasm.ReadModule(bf, importer)
	if err != nil {
		fmt.Printf("could not read module: %v", err)
	}

	if m.Export == nil {
		fmt.Printf("module has no export section")
	}

	vm, err := exec.NewVM(m)
	if err != nil {
		fmt.Printf("could not create VM: %v", err)
	}
	entry, ok := m.Export.Entries[ws.Method]

	if ok == false {
		fmt.Printf("method does not exist!")
	}
	index := int64(entry.Index)
	fIdx := m.Function.Types[int(index)]
	fType := m.Types.Entries[int(fIdx)]

	res, err := vm.ExecCode(index, ws.Args...)
	if err != nil {
		fmt.Printf("err=%v", err)
	}
	fmt.Printf("res:%[1]v (%[1]T)\n", res)
	switch fType.ReturnTypes[0] {
	case wasm.ValueTypeI32:
		return util.Int32ToBytes(res.(uint32))
	case wasm.ValueTypeI64:
		return util.Int64ToBytes(res.(uint64))
	case wasm.ValueTypeF32:
		return util.Float32ToBytes(res.(float32))
	case wasm.ValueTypeF64:
		return util.Float64ToBytes(res.(float64))
	default:
		return nil
	}
}

func importer(name string) (*wasm.Module, error) {
	f, err := os.Open(name + ".wasm")
	if err != nil {
		return nil, err
	}
	defer f.Close()
	m, err := wasm.ReadModule(f, nil)
	if err != nil {
		return nil, err
	}
	err = validate.VerifyModule(m)
	if err != nil {
		return nil, err
	}
	return m, nil
}

func (ws *WasmService) RegisterApi() {
	funs := wasm.InitNativeFuns()
	funs.Register("AbaAdd", ws.AbaAdd)
	funs.Register("AbaLog", ws.AbaLog)
	funs.Register("AbaLogString", ws.AbaLogString)
	funs.Register("AbaLogInt", ws.AbaLogInt)
	funs.Register("AbaGetCurrentHeight", ws.AbaGetCurrentHeight)
	funs.Register("AbaAccountGetBalance", ws.AbaAccountGetBalance)
	funs.Register("AbaAccountAddBalance", ws.AbaAccountAddBalance)
	funs.Register("AbaAccountSubBalance", ws.AbaAccountSubBalance)
	funs.Register("TokenIsExisted", ws.TokenIsExisted)
	funs.Register("TokenCreate", ws.TokenCreate)
}

func (ws *WasmService) AbaAdd(a int32, b int32) int32 {
	return a + b
}

func (ws *WasmService) AbaLogString(str string) int32 {
	fmt.Println(str)
	return 0
}

func (ws *WasmService) AbaLog(str string, msg interface{}) int32 {
	fmt.Println("AbaLog:---------")
	fmt.Printf(str, msg)
	return 0
}

func (ws *WasmService) AbaLogInt(value uint64) int32 {
	fmt.Println("value:", value)
	return 0
}

func (ws *WasmService) AbaGetCurrentHeight() uint64 {
	return ws.ledger.GetCurrentHeight()
}

func (ws *WasmService) AbaAccountGetBalance(token, addrHex string) uint64 {
	address := common.NewAddress(common.FromHex(addrHex))
	value, err := ws.ledger.AccountGetBalance(address, token)
	if err != nil {
		return 0
	}
	return value
}

func (ws *WasmService) AbaAccountAddBalance(value uint64, token, addrHex string) int32 {
	if err := ws.ledger.AccountAddBalance(common.NewAddress(common.FromHex(addrHex)), token, value); err != nil {
		log.Error(err)
		return -1
	}
	return 0
}

func (ws *WasmService) AbaAccountSubBalance(value uint64, token, addrHex string) int32 {
	if err := ws.ledger.AccountSubBalance(common.NewAddress(common.FromHex(addrHex)), token, value); err != nil {
		log.Error(err)
		return -1
	}
	return 0
}

func (ws *WasmService) TokenCreate(addrHex, token string, maximum uint64) int32 {
	if err := ws.ledger.TokenCreate(common.NewAddress(common.FromHex(addrHex)), token, maximum); err != nil {
		log.Error(err)
		return -1
	}
	return 0
}

func (ws *WasmService) TokenIsExisted(token string) int32 {
	ret := ws.ledger.TokenIsExisted(token)
	if ret {
		return 1
	} else {
		return 0
	}
}
