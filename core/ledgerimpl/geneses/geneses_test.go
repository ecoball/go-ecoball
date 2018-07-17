package geneses_test

import (
	"encoding/json"
	"github.com/ecoball/go-ecoball/common"
	"github.com/ecoball/go-ecoball/common/config"
	"github.com/ecoball/go-ecoball/common/elog"
	"github.com/ecoball/go-ecoball/core/ledgerimpl"
	"github.com/ecoball/go-ecoball/core/ledgerimpl/ledger"
	"github.com/ecoball/go-ecoball/core/state"
	"github.com/ecoball/go-ecoball/core/types"
	"math/big"
	"testing"
	"time"
)

var log = elog.NewLogger("worker2", elog.InfoLog)

var root = common.NameToIndex("root")
var worker1 = common.NameToIndex("worker1")
var worker2 = common.NameToIndex("worker2")
var worker3 = common.NameToIndex("worker3")

func TestGenesesBlockInit(t *testing.T) {
	l, err := ledgerimpl.NewLedger("/tmp/geneses")
	if err != nil {
		t.Fatal(err)
	}

	ShowAccountInfo(l, t)
}

func CreateAccountBlock(ledger ledger.Ledger, con *types.ConsensusData, t *testing.T) {
	timeStamp := time.Now().Unix()
	var txs []*types.Transaction

	invoke, err := types.NewInvokeContract(root, root, "owner", types.VmNative, "new_account",
		[]string{"worker1", common.AddressFromPubKey(common.FromHex("0x04e0c1852b110d1586bf6202abf6e519cc4161d00c3780c04cfde80fd66748cc189b6b0e2771baeb28189ec42a363461357422bf76b1e0724fc63fc97daf52769f")).HexString()}, 0, timeStamp)
	invoke.SetSignature(&config.Root)
	txs = append(txs, invoke)

	invoke, err = types.NewInvokeContract(root, root, "owner", types.VmNative, "new_account",
		[]string{"worker2", common.AddressFromPubKey(common.FromHex("0x049e78e40b0dcca842b94cb2586d47ecc61888b52dce958b41aa38613c80f6607ee1de23eebb912431eccfe0fea81f8a38792ffecee38c490dde846c646ce1f0ee")).HexString()}, 0, timeStamp)
	invoke.SetSignature(&config.Root)
	txs = append(txs, invoke)

	invoke, err = types.NewInvokeContract(root, root, "owner", types.VmNative, "new_account",
		[]string{"worker3", common.AddressFromPubKey(common.FromHex("0x0481bce0ad10bd3d8cdfd089ac5534379149ca5c3cdab28b5063f707d20f3a4a51f192ef7933e91e3fd0a8ea21d8dd735407780937c3c71753b486956fd481349f")).HexString()}, 0, timeStamp)
	invoke.SetSignature(&config.Root)
	txs = append(txs, invoke)

	block, err := ledger.NewTxBlock(txs, *con)
	if err != nil {
		t.Fatal(err)
	}
	block.SetSignature(&config.Root)
	if err := ledger.VerifyTxBlock(block); err != nil {
		t.Fatal(err)
	}
	if err := ledger.SaveTxBlock(block); err != nil {
		t.Fatal(err)
	}
}

func SetTokenAccountBlock(ledger ledger.Ledger, con *types.ConsensusData, t *testing.T) {
	perm := state.NewPermission("active", "owner", 2, []state.KeyFactor{}, []state.AccFactor{{Actor: worker1, Weight: 1, Permission: "active"}, {Actor: worker2, Weight: 1, Permission: "active"}})
	param, err := json.Marshal(perm)
	if err != nil {
		t.Fatal(err)
	}
	invoke, err := types.NewInvokeContract(worker3, root, "owner", types.VmNative, "set_account", []string{"root", string(param)}, 0, time.Now().Unix())
	invoke.SetSignature(&config.Worker3)
	transfer, err := types.NewTransfer(root, worker3, "owner", new(big.Int).SetUint64(1000), 100, time.Now().Unix())
	transfer.SetSignature(&config.Root)

	txs := []*types.Transaction{invoke, transfer}
	block, err := ledger.NewTxBlock(txs, *con)
	if err != nil {
		t.Fatal(err)
	}
	block.SetSignature(&config.Root)
	if err := ledger.VerifyTxBlock(block); err != nil {
		t.Fatal(err)
	}
	if err := ledger.SaveTxBlock(block); err != nil {
		t.Fatal(err)
	}
}

func TokenAccountTransferBlock(ledger ledger.Ledger, con *types.ConsensusData, t *testing.T) {
	transfer, err := types.NewTransfer(worker3, worker1, "active", new(big.Int).SetUint64(100), 101, time.Now().Unix())
	if err != nil {
		t.Fatal(err)
	}
	if err := transfer.SetSignature(&config.Worker2); err != nil {
		t.Fatal(err)
	}
	if err := transfer.SetSignature(&config.Worker3); err != nil {
		t.Fatal(err)
	}
	txs := []*types.Transaction{transfer}
	block, err := ledger.NewTxBlock(txs, *con)
	if err != nil {
		t.Fatal(err)
	}
	block.SetSignature(&config.Root)
	if err := ledger.VerifyTxBlock(block); err != nil {
		t.Fatal(err)
	}
	if err := ledger.SaveTxBlock(block); err != nil {
		t.Fatal(err)
	}
}

func ShowAccountInfo(l ledger.Ledger, t *testing.T) {
	acc, err := l.AccountGet(root)
	if err != nil {
		t.Fatal(err)
	}
	acc.Show()

	acc, err = l.AccountGet(worker1)
	if err != nil {
		t.Fatal(err)
	}
	acc.Show()

	acc, err = l.AccountGet(worker2)
	if err != nil {
		t.Fatal(err)
	}
	acc.Show()

	acc, err = l.AccountGet(worker3)
	if err != nil {
		t.Fatal(err)
	}
	acc.Show()
}