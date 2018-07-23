package ababft

import (
	"github.com/ecoball/go-ecoball/common"
	"github.com/ecoball/go-ecoball/common/elog"
	"testing"
	"github.com/ecoball/go-ecoball/core/types"
	"time"
	"github.com/ecoball/go-ecoball/core/ledgerimpl"
	"github.com/ecoball/go-ecoball/core/ledgerimpl/ledger"
	"github.com/ecoball/go-ecoball/common/config"
	"gitlab.quachain.net/aba/aba/consensus/ababft"
	"github.com/ecoball/go-ecoball/account"
	"bytes"
)

var log = elog.NewLogger("worker2", elog.InfoLog)

var root = common.NameToIndex("root")
var token = common.NameToIndex("token")
var worker1 = common.NameToIndex("worker1")
var worker2 = common.NameToIndex("worker2")
var worker3 = common.NameToIndex("worker3")

var accounts []account.Account

func TestABABFTPros(t *testing.T) {
	l, err := ledgerimpl.NewLedger("/tmp/geneses")
	if err != nil {
		t.Fatal(err)
	}

	ShowAccountInfo(l, t)
	// 1. set up parameters
	// 1.1 set the consensus algorithm
	config.ConsensusAlgorithm = "ABABFT"
	// 1.2 peers list
	ababft.Num_peers = 3
	ababft.Peers_list[0].PublicKey = config.Worker1.PublicKey
	ababft.Peers_list[1].PublicKey = config.Worker2.PublicKey
	ababft.Peers_list[2].PublicKey = config.Worker3.PublicKey
	// 1.3 accounts
	for i := 0; i < ababft.Num_peers; i++ {
		if ok := bytes.Equal(ababft.Peers_list[i].PublicKey,config.Worker1.PublicKey); ok== true {
			accounts[i].PublicKey = config.Worker1.PublicKey
			accounts[i].PrivateKey = config.Worker1.PrivateKey
		}
		if ok := bytes.Equal(ababft.Peers_list[i].PublicKey,config.Worker2.PublicKey); ok== true {
			accounts[i].PublicKey = config.Worker2.PublicKey
			accounts[i].PrivateKey = config.Worker2.PrivateKey
		}
		if ok := bytes.Equal(ababft.Peers_list[i].PublicKey,config.Worker3.PublicKey); ok== true {
			accounts[i].PublicKey = config.Worker3.PublicKey
			accounts[i].PrivateKey = config.Worker3.PrivateKey
		}
	}

	con, err := types.InitConsensusData(time.Now().Unix())



	//AddTokenAccount(l, con, t)
	//ContractStore(l, con, t)
	PledgeContract(l, con, t)
	ShowAccountInfo(l, t)
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

func PledgeContract(ledger ledger.Ledger, con *types.ConsensusData, t *testing.T) {
	var txs []*types.Transaction
	tokenContract, err := types.NewDeployContract(worker1, worker1, "active", types.VmNative, "system control", nil, 0, time.Now().Unix())
	if err != nil {
		t.Fatal(err)
	}
	tokenContract.SetSignature(&accounts[0])
	txs = append(txs, tokenContract)

	invoke, err := types.NewInvokeContract(worker1, worker1, "owner", "pledge",
		[]string{"worker1", "worker2", "10", "10"}, 0, time.Now().Unix())
	if err != nil {
		t.Fatal(err)
	}
	invoke.SetSignature(&accounts[0])
	txs = append(txs, invoke)

	block, err := ledger.NewTxBlock(txs, *con)
	if err != nil {
		t.Fatal(err)
	}
	block.SetSignature(&accounts[0])
	if err := ledger.VerifyTxBlock(block); err != nil {
		t.Fatal(err)
	}
	if err := ledger.SaveTxBlock(block); err != nil {
		t.Fatal(err)
	}
}