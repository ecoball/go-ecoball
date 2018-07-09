package example

import (
	"fmt"
	"github.com/ecoball/go-ecoball/account"
	"github.com/ecoball/go-ecoball/common"
	"github.com/ecoball/go-ecoball/core/state"
	"github.com/ecoball/go-ecoball/core/types"
	"math/big"
	"time"
	"github.com/ecoball/go-ecoball/common/config"
)

func ExampleAddAccount(state *state.State) error {
	from := common.NewAddress(common.FromHex("01b1a6569a557eafcccc71e0d02461fd4b601aea"))
	addr := common.NewAddress(common.FromHex("01ca5cdd56d99a0023166b337ffc7fd0d2c42330"))
	indexFrom := common.NameToIndex("from")
	indexAddr := common.NameToIndex("addr")
	if err := state.AddAccount(indexFrom, from); err != nil {
		return nil
	}
	if err := state.AddAccount(indexAddr, addr); err != nil {
		return nil
	}
	return nil
}

func ExampleTestInvoke(method string) *types.Transaction {
	indexFrom := common.NameToIndex("from")
	indexAddr := common.NameToIndex("addr")
	invoke, err := types.NewInvokeContract(indexFrom, indexAddr, types.VmWasm, method, []string{"01b1a6569a557eafcccc71e0d02461fd4b601aea", "Token.Test", "20000"}, 0, time.Now().Unix())
	if err != nil {
		panic(err)
		return nil
	}
	acc := account.Account{PrivateKey: common.FromHex(config.RootPrivkey), PublicKey: common.FromHex(config.RootPubkey), Alg: 0}
	if err := invoke.SetSignature(&acc); err != nil {
		panic(err)
	}
	return invoke
}

func ExampleTestDeploy(code []byte) *types.Transaction {
	indexFrom := common.NameToIndex("from")
	indexAddr := common.NameToIndex("addr")
	deploy, err := types.NewDeployContract(indexFrom, indexAddr, types.VmWasm, "test deploy", code, 0, time.Now().Unix())
	if err != nil {
		panic(err)
		return nil
	}
	acc := account.Account{PrivateKey: common.FromHex(config.RootPrivkey), PublicKey: common.FromHex(config.RootPubkey), Alg: 0}
	if err := deploy.SetSignature(&acc); err != nil {
		panic(err)
	}
	return deploy
}

func ExampleTestTx() *types.Transaction {
	indexFrom := common.NameToIndex("from")
	indexAddr := common.NameToIndex("addr")
	value := big.NewInt(100)
	tx, err := types.NewTransfer(indexFrom, indexAddr, value, 0, time.Now().Unix())
	if err != nil {
		fmt.Println(err)
		return nil
	}
	fmt.Println(config.RootPrivkey)
	acc := account.Account{PrivateKey: common.FromHex(config.RootPrivkey), PublicKey: common.FromHex(config.RootPubkey), Alg: 0}
	if err := tx.SetSignature(&acc); err != nil {
		fmt.Println(err)
		return nil
	}
	tx.Show()
	return tx
}
