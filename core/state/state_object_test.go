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
	indexAcc := common.NameToIndex("pct")
	acc1, _ := state.NewAccount("/tmp/state_object", indexAcc, addr)

	acc1.AddBalance(state.AbaToken, new(big.Int).SetUint64(100))
	value, err := acc1.Balance(state.AbaToken)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("Value:", value)
	data, err := acc1.Serialize()
	if err != nil {
		t.Fatal(err)
	}
	acc1.Show()

	acc2 := new(state.Account)
	if err := acc2.Deserialize(data); err != nil {
		t.Fatal(err)
	}

	value, err = acc1.Balance(state.AbaToken)
	if err != nil {
		t.Fatal(err)
	}
	if value.Uint64() != 100 {
		t.Fatal("balance error")
	}

	fmt.Println("Value:", value)
	acc2.Show()

	if acc1.JsonString() != acc2.JsonString() {
		t.Fatal("mismatch")
	}
}

func TestNewAccount(t *testing.T) {
	addr := common.NewAddress(common.FromHex("01ca5cdd56d99a0023166b337ffc7fd0d2c42330"))
	indexAcc := common.NameToIndex("pct")
	acc, err := state.NewAccount("/tmp/acc", indexAcc, addr)
	if err != nil {
		t.Fatal(err)
	}

	//for ; ;  {
		d, err := acc.Serialize()
		if err != nil {
			t.Fatal(err)
		}
		fmt.Println(d)
		if d[1] != 43 {
			t.Fatal("error")
		}
	//	time.Sleep(1 *time.Second)
	//}
}