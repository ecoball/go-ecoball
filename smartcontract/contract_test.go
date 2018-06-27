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
	"github.com/ecoball/go-ecoball/core/ledgerimpl"
	"github.com/ecoball/go-ecoball/core/types"
	"github.com/ecoball/go-ecoball/smartcontract"
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

func TestCreate(t *testing.T) {
	l, err := ledgerimpl.NewLedger("/tmp/quaker")
	if err != nil {
		t.Fatal(err)
	}
	server, err := smartcontract.NewContractService(l, nil)
	if err != nil {
		t.Fatal(err)
	}
	code, err := wasmservice.ReadWasm("../test/token.wasm")
	if err != nil {
		t.Fatal(err)
	}
	addr := "01b1a6569a557eafcccc71e0d02461fd4b601aea"
	name := "TokenTest"
	var arg = []string{addr, name, "100001"}
	s, err := server.ExecuteContract(types.VmWasm, "create", code, arg)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(s)

	balance := []string{addr, name}
	s, err = server.ExecuteContract(types.VmWasm, "balance", code, balance)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(s)
}
