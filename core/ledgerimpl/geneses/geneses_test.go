package geneses_test

import (
	"github.com/ecoball/go-ecoball/common"
	"github.com/ecoball/go-ecoball/common/config"
	"github.com/ecoball/go-ecoball/core/ledgerimpl"
	"github.com/ecoball/go-ecoball/core/types"
	"testing"
	"time"
	"fmt"
	"math/big"
)

func TestGenesesBlockInit(t *testing.T) {
	l, err := ledgerimpl.NewLedger("/tmp/geneses")
	if err != nil {
		t.Fatal(err)
	}

	timeStamp := time.Now().Unix()
	addr := common.NameToIndex("root")
	invoke, err := types.NewInvokeContract(addr, addr, "owner", types.VmNative, "new_account", []string{"pct", "01b1a6569a557eafcccc71e0d02461fd4b601aea"}, 0, timeStamp)
	invoke.SetSignature(&config.Root)
	if err := l.CheckTransaction(invoke); err != nil {
		t.Fatal(err)
	}
	txs := []*types.Transaction{invoke}
	con, err := types.InitConsensusData(timeStamp)
	if err != nil {
		t.Fatal(err)
	}
	block, err := l.NewTxBlock(txs, *con)
	if err != nil {
		t.Fatal(err)
	}
	block.SetSignature(&config.Root)
	if err := l.VerifyTxBlock(block); err != nil {
		t.Fatal(err)
	}
	if err := l.SaveTxBlock(block); err != nil {
		t.Fatal(err)
	}

	acc, err := l.AccountGet(common.NameToIndex("pct"))
	if err != nil {
		t.Fatal(err)
	}
	acc.Show()

	acc, err = l.AccountGet(common.NameToIndex("root"))
	if err != nil {
		t.Fatal(err)
	}
	acc.Show()

	transfer, err := types.NewTransfer(addr, common.NameToIndex("pct"), "owner", new(big.Int).SetUint64(1000), 100, timeStamp)
	transfer.SetSignature(&config.Root)
	if err := l.CheckTransaction(transfer); err != nil {
		t.Fatal(err)
	}

	txs2 := []*types.Transaction{transfer}
	block2, err := l.NewTxBlock(txs2, *con)
	if err != nil {
		t.Fatal(err)
	}
	block2.SetSignature(&config.Root)
	if err := l.VerifyTxBlock(block2); err != nil {
		t.Fatal(err)
	}
	if err := l.SaveTxBlock(block2); err != nil {
		t.Fatal(err)
	}

	acc, err = l.AccountGet(common.NameToIndex("root"))
	if err != nil {
		t.Fatal(err)
	}
	acc.Show()
	acc, err = l.AccountGet(common.NameToIndex("pct"))
	if err != nil {
		t.Fatal(err)
	}
	acc.Show()

	fmt.Println(common.ToHex(config.Root.PublicKey))
	fmt.Println(common.AddressFromPubKey(config.Root.PublicKey).HexString())
	fmt.Println(common.NewAddress(config.Root.PublicKey).HexString())

}
