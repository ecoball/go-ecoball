package geneses_test

import (
	"testing"
	"github.com/ecoball/go-ecoball/core/ledgerimpl"
	"github.com/ecoball/go-ecoball/core/types"
	"github.com/ecoball/go-ecoball/common"
	"time"
	"fmt"
	"github.com/ecoball/go-ecoball/common/config"
)

func TestGenesesBlockInit(t *testing.T) {
	privKey := config.RootPrivkey
	pubKey := config.RootPubkey
	fmt.Println(privKey)
	fmt.Println(pubKey)
	l, err := ledgerimpl.NewLedger("/tmp/geneses")
	if err != nil {
		t.Fatal(err)
	}
	header := l.GetCurrentHeader()
	header.Show()

	timeStamp := time.Now().Unix()
	addr := common.NameToIndex("root")
	invoke, err := types.NewInvokeContract(
		addr, addr, types.VmNative, "create",
		[]string{"01b1a6569a557eafcccc71e0d02461fd4b601aea", "pct"},
		0, timeStamp)
	invoke.SetSignature(&config.Root)
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
}
