package geneses_test

import (
	"encoding/json"
	"github.com/ecoball/go-ecoball/account"
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
	"fmt"
)

var log = elog.NewLogger("test", elog.InfoLog)

var root = common.NameToIndex("root")
var pct = common.NameToIndex("pct")
var test = common.NameToIndex("test")
var token = common.NameToIndex("token")

var pctAccount = account.Account{PrivateKey: common.FromHex("0xc3e2cbed03aacc62d8f32045013364ea493f6d24e84f26bcef4edc2e9d260c0e"),
	PublicKey: common.FromHex("0x04e0c1852b110d1586bf6202abf6e519cc4161d00c3780c04cfde80fd66748cc189b6b0e2771baeb28189ec42a363461357422bf76b1e0724fc63fc97daf52769f")}
var testAccount = account.Account{PrivateKey: common.FromHex("0x5238ede4f91f6c4f5f1f195cbf674e08cb6a18ae351e474b8927db82d3e5ecf5"),
	PublicKey: common.FromHex("0x049e78e40b0dcca842b94cb2586d47ecc61888b52dce958b41aa38613c80f6607ee1de23eebb912431eccfe0fea81f8a38792ffecee38c490dde846c646ce1f0ee")}
var tokenAccount = account.Account{PrivateKey: common.FromHex("0x105cb8f936eec87d35e42fc0f656ab4b7fc9a007cbf4554f829c44e528df6ce4"),
	PublicKey: common.FromHex("0x0481bce0ad10bd3d8cdfd089ac5534379149ca5c3cdab28b5063f707d20f3a4a51f192ef7933e91e3fd0a8ea21d8dd735407780937c3c71753b486956fd481349f")}

func TestGenesesBlockInit(t *testing.T) {
	log.Warn(common.AddressFromPubKey(tokenAccount.PublicKey).HexString())
	l, err := ledgerimpl.NewLedger("/tmp/geneses")
	if err != nil {
		t.Fatal(err)
	}
	con, err := types.InitConsensusData(time.Now().Unix())
	if err != nil {
		t.Fatal(err)
	}

	CreateAccountBlock(l, con, t)
	acc, err := l.AccountGet(pct)
	if err != nil {
		t.Fatal(err)
	}
	acc.Show()
	acc, err = l.AccountGet(test)
	if err != nil {
		t.Fatal(err)
	}
	acc.Show()
	acc, err = l.AccountGet(token)
	if err != nil {
		t.Fatal(err)
	}
	acc.Show()

	SetTokenAccountBlock(l, con, t)
	acc, err = l.AccountGet(common.NameToIndex("token"))
	if err != nil {
		t.Fatal(err)
	}
	acc.Show()

	TokenAccountTransferBlock(l, con, t)
	acc, err = l.AccountGet(common.NameToIndex("token"))
	if err != nil {
		t.Fatal(err)
	}
	acc.Show()
	acc, err = l.AccountGet(pct)
	if err != nil {
		t.Fatal(err)
	}
	acc.Show()
	permStr, err := l.FindPermission(token, "active")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(permStr)
}

func CreateAccountBlock(ledger ledger.Ledger, con *types.ConsensusData, t *testing.T) {
	timeStamp := time.Now().Unix()
	var txs []*types.Transaction

	invoke, err := types.NewInvokeContract(root, root, "owner", types.VmNative, "new_account",
		[]string{"pct", common.AddressFromPubKey(common.FromHex("0x04e0c1852b110d1586bf6202abf6e519cc4161d00c3780c04cfde80fd66748cc189b6b0e2771baeb28189ec42a363461357422bf76b1e0724fc63fc97daf52769f")).HexString()}, 0, timeStamp)
	invoke.SetSignature(&config.Root)
	txs = append(txs, invoke)

	invoke, err = types.NewInvokeContract(root, root, "owner", types.VmNative, "new_account",
		[]string{"test", common.AddressFromPubKey(common.FromHex("0x049e78e40b0dcca842b94cb2586d47ecc61888b52dce958b41aa38613c80f6607ee1de23eebb912431eccfe0fea81f8a38792ffecee38c490dde846c646ce1f0ee")).HexString()}, 0, timeStamp)
	invoke.SetSignature(&config.Root)
	txs = append(txs, invoke)

	invoke, err = types.NewInvokeContract(root, root, "owner", types.VmNative, "new_account",
		[]string{"token", common.AddressFromPubKey(common.FromHex("0x0481bce0ad10bd3d8cdfd089ac5534379149ca5c3cdab28b5063f707d20f3a4a51f192ef7933e91e3fd0a8ea21d8dd735407780937c3c71753b486956fd481349f")).HexString()}, 0, timeStamp)
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
	perm := state.NewPermission("active", "owner", 1, []state.KeyFactor{}, []state.AccFactor{{Actor: pct, Weight: 1, Permission: "active"}, {Actor: test, Weight: 1, Permission: "active"}})
	param, err := json.Marshal(perm)
	if err != nil {
		t.Fatal(err)
	}
	invoke, err := types.NewInvokeContract(token, root, "owner", types.VmNative, "set_account", []string{"token", string(param)}, 0, time.Now().Unix())
	invoke.SetSignature(&tokenAccount)
	transfer, err := types.NewTransfer(root, token, "owner", new(big.Int).SetUint64(1000), 100, time.Now().Unix())
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
	transfer, err := types.NewTransfer(token, pct, "active", new(big.Int).SetUint64(100), 101, time.Now().Unix())
	if err != nil {
		t.Fatal(err)
	}
	if err := transfer.SetSignature(&pctAccount); err != nil {
		t.Fatal(err)
	}
	if err := transfer.SetSignature(&testAccount); err != nil {
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
