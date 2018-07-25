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
	"github.com/ecoball/go-ecoball/smartcontract/wasmservice"
	"math/big"
	"testing"
	"time"
)

var log = elog.NewLogger("worker2", elog.InfoLog)

var root = common.NameToIndex("root")
var token = common.NameToIndex("token")
var worker1 = common.NameToIndex("worker1")
var worker2 = common.NameToIndex("worker2")
var worker3 = common.NameToIndex("worker3")
var delegate = common.NameToIndex("delegate")

func TestGenesesBlockInit(t *testing.T) {
	l, err := ledgerimpl.NewLedger("/tmp/geneses")
	if err != nil {
		t.Fatal(err)
	}
	con, err := types.InitConsensusData(time.Now().Unix())
	CreateAccountBlock(l, con, t)
	ShowAccountInfo(l, t)
	//AddTokenAccount(l, con, t)
	//ContractStore(l, con, t)
	PledgeContract(l, con, t)
	ShowAccountInfo(l, t)
	CancelPledgeContract(l, con, t)
	ShowAccountInfo(l, t)
}

func CreateAccountBlock(ledger ledger.Ledger, con *types.ConsensusData, t *testing.T) {
	//TODO
	var txs []*types.Transaction
	index := common.NameToIndex("root")
	//if err := ledger.AccountAddBalance(index, state.AbaToken, 10000); err != nil {
	//	t.Fatal(err)
	//}
	code, err := wasmservice.ReadWasm("../../../test/root/root.wasm")
	if err != nil {
		t.Fatal(err)
	}
	tokenContract, err := types.NewDeployContract(index, index, state.Active, types.VmWasm, "system control", code, 0, time.Now().Unix())
	if err != nil {
		t.Fatal(err)
	}
	if err := tokenContract.SetSignature(&config.Root); err != nil {
		t.Fatal(err)
	}
	txs = append(txs, tokenContract)

	invoke, err := types.NewInvokeContract(index, index, state.Owner,"new_account",
		[]string{"worker1", common.AddressFromPubKey(config.Worker1.PublicKey).HexString()}, 0, time.Now().Unix())
	invoke.SetSignature(&config.Root)
	txs = append(txs, invoke)

	invoke, err = types.NewInvokeContract(index, index, state.Owner, "new_account",
		[]string{"worker2", common.AddressFromPubKey(config.Worker2.PublicKey).HexString()}, 1, time.Now().Unix())
	invoke.SetSignature(&config.Root)
	txs = append(txs, invoke)

	invoke, err = types.NewInvokeContract(index, index, state.Owner, "new_account",
		[]string{"worker3", common.AddressFromPubKey(config.Worker3.PublicKey).HexString()}, 2, time.Now().Unix())
	invoke.SetSignature(&config.Root)
	txs = append(txs, invoke)

	perm := state.NewPermission(state.Active, state.Owner, 2, []state.KeyFactor{}, []state.AccFactor{{Actor: common.NameToIndex("worker1"), Weight: 1, Permission: "active"}, {Actor: common.NameToIndex("worker2"), Weight: 1, Permission: "active"}, {Actor: common.NameToIndex("worker3"), Weight: 1, Permission: "active"}})
	param, err := json.Marshal(perm)
	if err != nil {
		t.Fatal(err)
	}
	invoke, err = types.NewInvokeContract(index, index, state.Active, "set_account", []string{"root", string(param)}, 0, time.Now().Unix())
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
	invoke, err := types.NewInvokeContract(worker3, root, "owner", "set_account", []string{"root", string(param)}, 0, time.Now().Unix())
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

	acc, err = l.AccountGet(delegate)
	if err != nil {
		t.Fatal(err)
	}
	acc.Show()
}

func AddTokenAccount(ledger ledger.Ledger, con *types.ConsensusData, t *testing.T) {
	var txs []*types.Transaction
	invoke, err := types.NewInvokeContract(root, root, "owner", "new_account",
		[]string{"token", common.AddressFromPubKey(config.Worker1.PublicKey).HexString()}, 0, time.Now().Unix())
	if err != nil {
		t.Fatal(err)
	}
	invoke.SetSignature(&config.Root)
	txs = append(txs, invoke)

	code, err := wasmservice.ReadWasm("../../../test/token/token.wasm")
	if err != nil {
		t.Fatal(err)
	}
	tokenContract, err := types.NewDeployContract(token, token, "active", types.VmWasm, "system control", code, 0, time.Now().Unix())
	if err != nil {
		t.Fatal(err)
	}
	tokenContract.SetSignature(&config.Worker1)
	txs = append(txs, tokenContract)

	invoke, err = types.NewInvokeContract(token, token, "owner", "create",
		[]string{"token", "aba", "10000"}, 0, time.Now().Unix())
	if err != nil {
		t.Fatal(err)
	}
	invoke.SetSignature(&config.Worker1)
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

func ContractStore(ledger ledger.Ledger, con *types.ConsensusData, t *testing.T) {
	var txs []*types.Transaction
	code, err := wasmservice.ReadWasm("../../../test/store/store.wasm")
	if err != nil {
		t.Fatal(err)
	}
	tokenContract, err := types.NewDeployContract(worker3, worker3, "active", types.VmWasm, "system control", code, 0, time.Now().Unix())
	if err != nil {
		t.Fatal(err)
	}
	tokenContract.SetSignature(&config.Worker3)
	txs = append(txs, tokenContract)

	invoke, err := types.NewInvokeContract(worker3, worker3, "owner", "StoreSet",
		[]string{"pct", "panchangtao"}, 0, time.Now().Unix())
	if err != nil {
		t.Fatal(err)
	}
	invoke.SetSignature(&config.Worker3)
	txs = append(txs, invoke)

	invoke, err = types.NewInvokeContract(worker3, worker3, "owner", "StoreGet",
		[]string{"pct"}, 0, time.Now().Unix())
	if err != nil {
		t.Fatal(err)
	}
	invoke.SetSignature(&config.Worker3)
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

func PledgeContract(ledger ledger.Ledger, con *types.ConsensusData, t *testing.T) {
	var txs []*types.Transaction
	tokenContract, err := types.NewDeployContract(delegate, delegate, "active", types.VmNative, "system control", nil, 0, time.Now().Unix())
	if err != nil {
		t.Fatal(err)
	}
	tokenContract.SetSignature(&config.Delegate)
	txs = append(txs, tokenContract)

	invoke, err := types.NewInvokeContract(root, delegate, "owner", "pledge", []string{"root", "worker2", "10", "10"}, 0, time.Now().Unix())
	if err != nil {
		t.Fatal(err)
	}
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
func CancelPledgeContract(ledger ledger.Ledger, con *types.ConsensusData, t *testing.T) {
	var txs []*types.Transaction
	invoke, err := types.NewInvokeContract(root, delegate, "owner", "cancel_pledge",
		[]string{"root", "worker2", "10", "10"}, 0, time.Now().Unix())
	if err != nil {
		t.Fatal(err)
	}
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

