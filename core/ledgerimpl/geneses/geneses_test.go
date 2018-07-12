package geneses_test

import (
	"github.com/ecoball/go-ecoball/common"
	"github.com/ecoball/go-ecoball/common/config"
	"github.com/ecoball/go-ecoball/core/ledgerimpl"
	"github.com/ecoball/go-ecoball/core/types"
	"testing"
	"time"
	"github.com/ecoball/go-ecoball/account"
)
var root = common.NameToIndex("root")
var pct = common.NameToIndex("pct")
var test = common.NameToIndex("test")
var token = common.NameToIndex("token")

var pctAccount = account.Account{PrivateKey:common.FromHex("0xc3e2cbed03aacc62d8f32045013364ea493f6d24e84f26bcef4edc2e9d260c0e"),
PublicKey:common.FromHex("0x04e0c1852b110d1586bf6202abf6e519cc4161d00c3780c04cfde80fd66748cc189b6b0e2771baeb28189ec42a363461357422bf76b1e0724fc63fc97daf52769f")}
var testAccount = account.Account{PrivateKey:common.FromHex("0x5238ede4f91f6c4f5f1f195cbf674e08cb6a18ae351e474b8927db82d3e5ecf5"),
PublicKey:common.FromHex("0x049e78e40b0dcca842b94cb2586d47ecc61888b52dce958b41aa38613c80f6607ee1de23eebb912431eccfe0fea81f8a38792ffecee38c490dde846c646ce1f0ee")}
var tokenAccount = account.Account{PrivateKey:common.FromHex("0x105cb8f936eec87d35e42fc0f656ab4b7fc9a007cbf4554f829c44e528df6ce4"),
PublicKey:common.FromHex("0x0481bce0ad10bd3d8cdfd089ac5534379149ca5c3cdab28b5063f707d20f3a4a51f192ef7933e91e3fd0a8ea21d8dd735407780937c3c71753b486956fd481349f")}

func TestGenesesBlockInit(t *testing.T) {
	l, err := ledgerimpl.NewLedger("/tmp/geneses")
	if err != nil {
		t.Fatal(err)
	}
	timeStamp := time.Now().Unix()
	invoke_pct, err := types.NewInvokeContract(root, root, "owner", types.VmNative, "new_account",
		[]string{"pct", common.AddressFromPubKey(common.FromHex("0xc3e2cbed03aacc62d8f32045013364ea493f6d24e84f26bcef4edc2e9d260c0e")).HexString()}, 0, timeStamp)
	invoke_pct.SetSignature(&config.Root)
	invoke_test, err := types.NewInvokeContract(root, root, "owner", types.VmNative, "new_account",
		[]string{"test", common.AddressFromPubKey(common.FromHex("0x5238ede4f91f6c4f5f1f195cbf674e08cb6a18ae351e474b8927db82d3e5ecf5")).HexString()}, 0, timeStamp)
	invoke_test.SetSignature(&config.Root)
	invoke_token, err := types.NewInvokeContract(root, root, "owner", types.VmNative, "new_account",
		[]string{"token", common.AddressFromPubKey(common.FromHex("0x105cb8f936eec87d35e42fc0f656ab4b7fc9a007cbf4554f829c44e528df6ce4")).HexString()}, 0, timeStamp)
	invoke_token.SetSignature(&config.Root)

	txs := []*types.Transaction{invoke_pct, invoke_test, invoke_token}
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

	acc_pct, err := l.AccountGet(pct)
	if err != nil {
		t.Fatal(err)
	}
	acc_pct.Show()
	acc_test, err := l.AccountGet(test)
	if err != nil {
		t.Fatal(err)
	}
	acc_test.Show()
	acc_token, err := l.AccountGet(token)
	if err != nil {
		t.Fatal(err)
	}
	acc_token.Show()

	param := `{"threshold" : 1, "keys" : [], "accounts" : [{"permission":{"actor":"pct","permission":"active"},"weight":1}, {"permission":{"actor":"test","permission":"active"},"weight":1}]}`
	invoke, err := types.NewInvokeContract(token, token, "owner", types.VmNative, "set_account",
		[]string{param}, 0, timeStamp)
	invoke.SetSignature(&tokenAccount)
	/*
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
    */
}
