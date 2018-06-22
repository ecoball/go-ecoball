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

package smartcontract_test

import (
	"fmt"
	"github.com/ecoball/go-ecoball/smartcontract/wasmservice"
	"testing"
)

func TestNewInvokeContract(t *testing.T) {
	code, err := wasmservice.ReadWasm("../test/contract_test.wasm")
	if err != nil {
		t.Fatal("read contract error\n")
	}
	method := "Invoke"
	arg := []uint64{13, 14}
	s, err := wasmservice.NewWasmService(nil, method, code, arg)
	ret := s.Execute()
	fmt.Printf("%v", ret)
}

func TestAdd(t *testing.T) {
	code, err := wasmservice.ReadWasm("../test/aba_add.wasm")
	if err != nil {
		t.Fatal(err)
	}
	s, err := wasmservice.NewWasmService(nil, "main", code, nil)
	fmt.Println(s.Execute())
}

func TestLog(t *testing.T) {
	code, err := wasmservice.ReadWasm("../test/aba_log.wasm")
	if err != nil {
		t.Fatal(err)
	}
	s, err := wasmservice.NewWasmService(nil, "main", code, nil)
	fmt.Println(s.Execute())
}