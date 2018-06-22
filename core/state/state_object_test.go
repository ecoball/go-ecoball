package state_test

import (
	"fmt"
	"github.com/ecoball/go-ecoball/common"
	"github.com/ecoball/go-ecoball/core/state"
	"math/big"
	"testing"
)

func TestStateObject(t *testing.T) {
	addr := common.NewAddress(common.FromHex("01ca5cdd56d99a0023166b337ffc7fd0d2c42330"))
	token := common.NewAddress(common.FromHex("01b1a6569a557eafcccc71e0d02461fd4b601aea"))
	obj, _ := state.NewStateObject(addr)
	obj.AddBalance(token, []byte("aba"), new(big.Int).SetUint64(100))
	value, err := obj.Balance(token, []byte("aba"))
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("Balance:", value)
	data, err := obj.Serialize()
	if err != nil {
		t.Fatal(err)
	}
	obj.Show()

	obj2 := new(state.StateObject)
	if err := obj2.Deserialize(data); err != nil {
		t.Fatal(err)
	}
	if !obj.Address.Equals(&obj2.Address) {
		t.Fatal("mismatch")
	}
	obj2.Show()
}
