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
	"math/big"
	"github.com/ecoball/go-ecoball/common/event"
	"os"
	"github.com/ecoball/go-ecoball/txpool"
)


var root = common.NameToIndex("root")
var token = common.NameToIndex("token")
var worker1 = common.NameToIndex("worker1")
var worker2 = common.NameToIndex("worker2")
var worker3 = common.NameToIndex("worker3")
var delegate = common.NameToIndex("delegate")

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
	fmt.Println("config.delegate:",config.Delegate)
	//start transaction pool
	if _, err := txpool.Start(); err != nil {
		log.Fatal("start txpool error, ", err.Error())
		os.Exit(1)
	}
	fmt.Println("start txpool ok")

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
	PledgeContract(l, con, t)

	ShowAccountInfo(l, t)

	// 4.create ababft service and start it
	abas,err := Service_ababft_gen(l, &accounts[2])
	abas.Start()

	// 5. test ABABFTStart in actor
	event.Send(event.ActorConsensus,event.ActorConsensus,ABABFTStart{})
	// 5a. create a tx for the block generation later
	// add 1000ABA to worker1 from worker2
	transfer_t, err := types.NewTransfer(worker2, worker1, "owner", new(big.Int).SetUint64(800), 400, time.Now().Unix())
	transfer_t.SetSignature(&config.Worker2)
	event.Send(event.ActorNil, event.ActorTxPool,transfer_t)
	log.Debug("create one tx for tx pool",transfer_t)

	// 6. test Signature_Preblock
	// generate the signature for previous block
	curheader := l.GetCurrentHeader()
	log.Debug("current height:", curheader.Height)
	hash_t := curheader.Hash
	for i:=0;i<Num_peers;i++{
		var signaturepre_send Signature_Preblock
		signaturepre_send.Signature_preblock.PubKey = accounts[i].PublicKey
		signaturepre_send.Signature_preblock.SigData,_ = accounts[i].Sign(hash_t.Bytes())
		signaturepre_send.Signature_preblock.Round = uint32(0)
		signaturepre_send.Signature_preblock.Height = uint32(curheader.Height)
		// fmt.Println("Signature_preblock:",signaturepre_send)
		// broadcast
		event.Send(event.ActorNil, event.ActorConsensus, signaturepre_send)
	}



	// AddTokenAccount(l, con, t)








	ShowAccountInfo(l, t)


	time.Sleep(time.Second * 100)
	/*







	//ContractStore(l, con, t)
	//
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

	// add 1000ABA to worker1 from root
	transfer1, err := types.NewTransfer(root, worker1, "owner", new(big.Int).SetUint64(1000), 100, time.Now().Unix())
	transfer1.SetSignature(&config.Root)
	txs = append(txs,transfer1)
	// add 2000ABA to worker2 from root
	transfer2, err := types.NewTransfer(root, worker2, "owner", new(big.Int).SetUint64(2000), 200, time.Now().Unix())
	transfer2.SetSignature(&config.Root)
	txs = append(txs,transfer2)
	// add 3000ABA to worker3 from root
	transfer3, err := types.NewTransfer(root, worker3, "owner", new(big.Int).SetUint64(3000), 300, time.Now().Unix())
	transfer3.SetSignature(&config.Root)
	txs = append(txs,transfer3)

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

	/*
	var txs1 []*types.Transaction
		tokenContract1, err := types.NewDeployContract(delegate, delegate, "active", types.VmNative, "system control", nil, 0, time.Now().Unix())
	if err != nil {
		t.Fatal(err)
	}
	tokenContract.SetSignature(&config.Delegate)
	txs1 = append(txs1, tokenContract1)
	fmt.Println("aa")
	// add cpu and net resource for worker1
	invoke, err = types.NewInvokeContract(root, delegate, "owner", "pledge", []string{"root", "worker1", "100", "100"}, 10, time.Now().Unix())
	if err != nil {
		t.Fatal(err)
	}
	invoke.SetSignature(&config.Root)
	txs1 = append(txs1, invoke)
	// add cpu and net resource for worker2
	invoke, err = types.NewInvokeContract(root, delegate, "owner", "pledge", []string{"root", "worker2", "100", "100"}, 20, time.Now().Unix())
	if err != nil {
		t.Fatal(err)
	}
	invoke.SetSignature(&config.Root)
	txs1 = append(txs1, invoke)
	// add cpu and net resource for worker3
	invoke, err = types.NewInvokeContract(root, delegate, "owner", "pledge", []string{"root", "worker3", "100", "100"}, 30, time.Now().Unix())
	if err != nil {
		t.Fatal(err)
	}
	invoke.SetSignature(&config.Root)
	txs1 = append(txs1, invoke)

	block, err = ledger.NewTxBlock(txs1, *con)
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
	 */

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

	code, err := wasmservice.ReadWasm("../../test/token/token.wasm")
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
func PledgeContract(ledger ledger.Ledger, con *types.ConsensusData, t *testing.T) {
	var txs []*types.Transaction
	tokenContract, err := types.NewDeployContract(delegate, delegate, "active", types.VmNative, "system control", nil, 0, time.Now().Unix())
	if err != nil {
		t.Fatal(err)
	}
	log.Debug("delegate account",config.Delegate)
	tokenContract.SetSignature(&config.Delegate)
	txs = append(txs, tokenContract)

	invoke, err := types.NewInvokeContract(root, delegate, "owner", "pledge", []string{"root", "worker1", "100", "100"}, 10, time.Now().Unix())
	if err != nil {
		t.Fatal(err)
	}
	invoke.SetSignature(&config.Root)
	txs = append(txs, invoke)
	invoke, err = types.NewInvokeContract(root, delegate, "owner", "pledge", []string{"root", "worker2", "200", "200"}, 20, time.Now().Unix())
	if err != nil {
		t.Fatal(err)
	}
	invoke.SetSignature(&config.Root)
	txs = append(txs, invoke)
	invoke, err = types.NewInvokeContract(root, delegate, "owner", "pledge", []string{"root", "worker3", "300", "300"}, 30, time.Now().Unix())
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
