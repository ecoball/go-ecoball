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
	"github.com/ecoball/go-ecoball/common"
	"bytes"
)

func xTestNewInvokeContract(t *testing.T) {
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
	code, err := wasmservice.ReadWasm("../test/add/aba_add.wasm")
	if err != nil {
		t.Fatal(err)
	}
	s, err := wasmservice.NewWasmService(nil, "main", code, nil)
	fmt.Println(s.Execute())
}

func TestLog(t *testing.T) {
	code, err := wasmservice.ReadWasm("../test/hello/hello.wasm")
	if err != nil {
		t.Fatal(err)
	}
	s, err := wasmservice.NewWasmService(nil, "main", code, nil)
	fmt.Println(s.Execute())
}

func xTestCreate(t *testing.T) {
	l, err := ledgerimpl.NewLedger("/tmp/quaker")
	if err != nil {
		t.Fatal(err)
	}
	server, err := smartcontract.NewContractService(l)
	if err != nil {
		t.Fatal(err)
	}
	code, err := wasmservice.ReadWasm("../test/token.wasm")
	if err != nil {
		t.Fatal(err)
	}
	strCode := common.ToHex(code)
	fmt.Println(strCode)
	fmt.Println()
	fmt.Println(con)
	if strCode != con2 {
		t.Fatal("string un equal")
	}
	if !bytes.Equal(common.FromHex(con2), code) {
		t.Fatal("un equal")
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

var con string = "0x0061736d0100000001efbfbdefbfbdefbfbdefbfbd00086000017f60037e7f7f017f60027f7f017f60017e017f60037f7f7f017f60017f017f60047f7f7f7e017f60027f7f017e02efbfbdefbfbdefbfbdefbfbd000603656e76144162614163636f756e7441646442616c616e6365000103656e76144162614163636f756e7447657442616c616e6365000203656e76144162614163636f756e7453756242616c616e6365000103656e76094162614c6f67496e74000303656e760b546f6b656e437265617465000403656e760e546f6b656e497345786973746564000503efbfbdefbfbdefbfbdefbfbd000304060704efbfbdefbfbdefbfbdefbfbd000170000005efbfbdefbfbdefbfbdefbfbd0001000106efbfbdefbfbdefbfbdefbfbd000007efbfbdefbfbdefbfbdefbfbd0004066d656d6f72790200066372656174650006087472616e7366657200070762616c616e636500080ae68080efbfbd0003efbfbdefbfbdefbfbdefbfbd0001017f417f21030240200110054101460d00417f410020002001200210041b21030b20030befbfbdefbfbdefbfbdefbfbd0001017f024020032000200210022204450d0020040f0b20032000200210000befbfbdefbfbdefbfbdefbfbd0001017e200120001001efbfbd220210031a20020b"
var con2 string = "0x0061736d0100000001b180808000086000017f60037e7f7f017f60027f7f017f60017e017f60037f7f7f017f60017f017f60047f7f7f7e017f60027f7f017e0289818080000603656e76144162614163636f756e7441646442616c616e6365000103656e76144162614163636f756e7447657442616c616e6365000203656e76144162614163636f756e7453756242616c616e6365000103656e76094162614c6f67496e74000303656e760b546f6b656e437265617465000403656e760e546f6b656e497345786973746564000503848080800003040607048480808000017000000583808080000100010681808080000007a88080800004066d656d6f72790200066372656174650006087472616e7366657200070762616c616e636500080ae68080800003a58080800001017f417f21030240200110054101460d00417f410020002001200210041b21030b20030b9f8080800001017f024020032000200210022204450d0020040f0b20032000200210000b928080800001017e200120001001ac220210031a20020b"
