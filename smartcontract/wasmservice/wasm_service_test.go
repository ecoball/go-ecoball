package wasmservice_test

import (
	"fmt"
	"github.com/ecoball/go-ecoball/common"
	"github.com/ecoball/go-ecoball/smartcontract/wasmservice"
	"testing"
	"github.com/ecoball/go-ecoball/test/example"
)

func TestLog(t *testing.T) {
	code, err := wasmservice.ReadWasm("../../test/hello/hello.wasm")
	if err != nil {
		t.Fatal(err)
	}
	ws := &wasmservice.WasmService{
		Code:   code,
		Method: "main",
	}
	ws.RegisterApi()
	fmt.Println(ws.Execute())
}

func TestAbaLog(t *testing.T) {
	code, err := wasmservice.ReadWasm("../../test/aba_log/aba_log.wasm")
	if err != nil {
		t.Fatal(err)
	}
	arg, err := common.StringToPointer("Hello World!")
	if err != nil {
		t.Fatal(err)
	}
	ws := &wasmservice.WasmService{
		Code:   code,
		Args:   []uint64{arg},
		Method: "AbaLog",
	}
	ws.RegisterApi()
	fmt.Println(ws.Execute())
}

func TestNewAccount(t *testing.T) {
	code, err := wasmservice.ReadWasm("../../test/root/new_account.wasm")
	if err != nil {
		t.Fatal(err)
	}
	arg1, err := common.StringToPointer("worker")
	arg2, err := common.StringToPointer("0x011d09d7f87741494bdc69b67b8b2dc4ecd49c7e")
	if err != nil {
		t.Fatal(err)
	}
	ws := &wasmservice.WasmService{
		Code:   code,
		Args:   []uint64{arg1, arg2},
		Method: "new_account",
	}
	ws.RegisterApi()
	fmt.Println(ws.Execute())
}

