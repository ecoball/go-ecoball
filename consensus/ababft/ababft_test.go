package ababft

import (
	"github.com/ecoball/go-ecoball/common"
	"testing"
	"github.com/ecoball/go-ecoball/account"
	"github.com/ecoball/go-ecoball/core/ledgerimpl/ledger"
	"github.com/ecoball/go-ecoball/common/config"
	"bytes"
	"fmt"
	"github.com/ecoball/go-ecoball/core/types"
	"time"
	"github.com/ecoball/go-ecoball/core/state"
	"encoding/json"
	"github.com/ecoball/go-ecoball/smartcontract/wasmservice"
	"github.com/ecoball/go-ecoball/core/ledgerimpl"
)


var root = common.NameToIndex("root")
var token = common.NameToIndex("token")
var worker1 = common.NameToIndex("worker1")
var worker2 = common.NameToIndex("worker2")
var worker3 = common.NameToIndex("worker3")

var accounts []account.Account
func TestABABFTPros(t *testing.T) {
	log.Debug("start test for ababft")
	// l, err := ledgerimpl.NewLedger("/tmp/geneses")
	l, err := ledgerimpl.NewLedger("./")
	if err != nil {
		t.Fatal(err)
	}
	log.Debug("ledger build ok")
	fmt.Println("config:",config.ConsensusAlgorithm)

	// 1. set up parameters
	// 1.1 set the consensus algorithm
	config.ConsensusAlgorithm = "ABABFT"
	// 1.2 peers list
	Num_peers = 3
	var peer Peer_info
	peer.PublicKey = config.Worker1.PublicKey
	peer.Index = 1
	Peers_list = append(Peers_list,peer)
	peer.PublicKey = config.Worker2.PublicKey
	peer.Index = 2
	Peers_list = append(Peers_list,peer)
	peer.PublicKey = config.Worker3.PublicKey
	peer.Index = 3
	Peers_list = append(Peers_list,peer)

	// 1.3 accounts
	for i := 0; i < Num_peers; i++ {
		var account account.Account
		if ok := bytes.Equal(Peers_list[i].PublicKey,config.Worker1.PublicKey); ok== true {
			account.PublicKey = config.Worker1.PublicKey
			account.PrivateKey = config.Worker1.PrivateKey
		}
		if ok := bytes.Equal(Peers_list[i].PublicKey,config.Worker2.PublicKey); ok== true {
			account.PublicKey = config.Worker2.PublicKey
			account.PrivateKey = config.Worker2.PrivateKey
		}
		if ok := bytes.Equal(Peers_list[i].PublicKey,config.Worker3.PublicKey); ok== true {
			account.PublicKey = config.Worker3.PublicKey
			account.PrivateKey = config.Worker3.PrivateKey
		}
		accounts = append(accounts,account)
	}

	// 2. create the consensus data
	con, err := types.InitConsensusData(time.Now().Unix())

	// 3. genesis block, to create accounts and bind them with permissions
	CreateAccountBlock(l, con, t)

	/*






	ShowAccountInfo(l, t)
	//AddTokenAccount(l, con, t)
	//ContractStore(l, con, t)
	// PledgeContract(l, con, t)
	ShowAccountInfo(l, t)
	*/
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

func CreateAccountBlock(ledger ledger.Ledger, con *types.ConsensusData, t *testing.T) {
	//TODO
	var txs []*types.Transaction
	index := common.NameToIndex("root")
	if err := ledger.AccountAddBalance(index, state.AbaToken, 10000); err != nil {
		t.Fatal(err)
	}

	code, err := wasmservice.ReadWasm("../../test/root/root.wasm")
	if err != nil {
		t.Fatal(err)
	}
	log.Debug("load wasm ok")
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
/*


func PledgeContract(ledger ledger.Ledger, con *types.ConsensusData, t *testing.T) {
	var txs []*types.Transaction
	tokenContract, err := types.NewDeployContract(worker1, worker1, "active", types.VmNative, "system control", nil, 0, time.Now().Unix())
	if err != nil {
		t.Fatal(err)
	}
	tokenContract.SetSignature(&config.Worker1)
	txs = append(txs, tokenContract)

	invoke, err := types.NewInvokeContract(root, worker1, "owner", "pledge",
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

*/