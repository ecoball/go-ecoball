// Copyright 2018 The go-ecoball Authors
// This file is part of the go-ecoball library.
//
// The go-ecoball library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ecoball library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ecoball library. If not, see <http://www.gnu.org/licenses/>.

package transaction

import (
	"errors"
	"fmt"
	"github.com/ecoball/go-ecoball/common"
	"github.com/ecoball/go-ecoball/common/config"
	"github.com/ecoball/go-ecoball/common/elog"
	errs "github.com/ecoball/go-ecoball/common/errors"
	"github.com/ecoball/go-ecoball/common/event"
	"github.com/ecoball/go-ecoball/core/ledgerimpl/geneses"
	"github.com/ecoball/go-ecoball/core/ledgerimpl/ledger"
	"github.com/ecoball/go-ecoball/core/state"
	"github.com/ecoball/go-ecoball/core/store"
	"github.com/ecoball/go-ecoball/core/types"
	"github.com/ecoball/go-ecoball/smartcontract"
	"math/big"
)

var log = elog.NewLogger("Chain Tx", elog.NoticeLog)

type ChainTx struct {
	BlockStore     store.Storage
	HeaderStore    store.Storage
	TxsStore       store.Storage
	ConsensusStore store.Storage

	CurrentHeader *types.Header
	StateDB       *state.State
	ledger        ledger.Ledger
}

func NewTransactionChain(path string, ledger ledger.Ledger) (c *ChainTx, err error) {
	c = &ChainTx{ledger: ledger}
	c.BlockStore, err = store.NewLevelDBStore(path+config.StringBlock, 0, 0)
	if err != nil {
		return nil, err
	}
	c.HeaderStore, err = store.NewLevelDBStore(path+config.StringHeader, 0, 0)
	if err != nil {
		return nil, err
	}
	c.TxsStore, err = store.NewLevelDBStore(path+config.StringTxs, 0, 0)
	if err != nil {
		return nil, err
	}

	existed, err := c.RestoreCurrentHeader()
	if err != nil {
		return nil, err
	}
	if existed {
		if c.StateDB, err = state.NewState(path+config.StringState, c.CurrentHeader.StateHash); err != nil {
			return nil, err
		}
	} else {
		if c.StateDB, err = state.NewState(path+config.StringState, common.Hash{}); err != nil {
			return nil, err
		}
	}

	return c, nil
}
/**
*  @brief  create a new block, this function will execute the transaction to rebuild mpt trie
*  @param  consensusData - the data of consensus module set
*/
func (c *ChainTx) NewBlock(ledger ledger.Ledger, txs []*types.Transaction, consensusData types.ConsensusData) (*types.Block, error) {
	for i := 0; i < len(txs); i++ {
		if _, err := c.HandleTransaction(ledger, txs[i]); err != nil {
			log.Error("Handle Transaction Error:", err)
			return nil, err
		}
		//event.Send(event.ActorLedger, event.ActorP2P, txs[i]) //send result to p2p actor
	}
	return types.NewBlock(c.CurrentHeader, c.StateDB.GetHashRoot(), consensusData, txs)
}
/**
*  @brief  if create a new block failed, then need to reset state DB
*  @param  hash - the root hash of mpt trie which need to reset
*/
func (c *ChainTx) ResetStateDB(hash common.Hash) error {
	return c.StateDB.Reset(hash)
}
/**
*  @brief  check block's signature and all transactions
*  @param  block - the block need to verify
*/
func (c *ChainTx) VerifyTxBlock(block *types.Block) error {
	result, err := block.VerifySignature()
	if err != nil {
		log.Error("Block VerifySignature Failed")
		return err
	}
	if result == false {
		return errors.New("block verify signature failed")
	}
	for _, v := range block.Transactions {
		if err := c.CheckTransaction(v); err != nil {
			log.Error("Transactions VerifySignature Failed")
			return err
		}
	}
	return nil
}
/**
*  @brief  save a block into levelDB, then push this block to p2p and tx pool module, and commit mpt trie into levelDB
*  @param  block - the block need to save
*/
func (c *ChainTx) SaveBlock(block *types.Block) error {
	if block == nil {
		return errors.New("block is nil")
	}
	//block.Show()
	if err := event.Publish(event.ActorLedger, block, event.ActorTxPool, event.ActorP2P); err != nil {
		log.Warn(err)
	}
	for _, t := range block.Transactions {
		payload, _ := t.Serialize()
		if t.Type == types.TxDeploy {
			c.TxsStore.BatchPut(common.IndexToBytes(t.Addr), payload)
		} else {
			c.TxsStore.BatchPut(t.Hash.Bytes(), payload)
		}
	}
	if err := c.TxsStore.BatchCommit(); err != nil {
		return err
	}

	payload, err := block.Header.Serialize()
	if err != nil {
		return err
	}
	if err := c.HeaderStore.Put(block.Header.Hash.Bytes(), payload); err != nil {
		return err
	}
	payload, _ = block.Serialize()
	c.BlockStore.BatchPut(block.Hash.Bytes(), payload)
	if err := c.BlockStore.BatchCommit(); err != nil {
		return err
	}
	c.StateDB.CommitToDB()
	c.CurrentHeader = block.Header
	return nil
}
/**
*  @brief  return the highest block's hash
*/
func (c *ChainTx) GetTailBlockHash() common.Hash {
	return c.CurrentHeader.Hash
}
/**
*  @brief  get a block by a hash value
*  @param  hash - the block's hash need to return
*/
func (c *ChainTx) GetBlock(hash common.Hash) (*types.Block, error) {
	dataBlock, err := c.BlockStore.Get(hash.Bytes())
	if err != nil {
		log.Error(err)
		return nil, err
	}
	block := new(types.Block)
	if err := block.Deserialize(dataBlock); err != nil {
		return nil, err
	}
	return block, nil
}
/**
*  @brief  get a block by a height value
*  @param  height - the block's height need to return
*/
func (c *ChainTx) GetBlockByHeight(height uint64) (*types.Block, error) {
	headers, err := c.HeaderStore.SearchAll()
	if err != nil {
		return nil, err
	}
	if len(headers) == 0 {
		return nil, nil
	}
	log.Info("The geneses block is existed:", len(headers))
	var hash common.Hash
	for _, v := range headers {
		header := new(types.Header)
		if err := header.Deserialize([]byte(v)); err != nil {
			return nil, err
		}
		if header.Height == height {
			hash = header.Hash
			break
		}
	}
	if hash.Equals(&common.Hash{}) {
		return nil, errors.New(fmt.Sprintf("can't find the block by height:%d", height))
	}
	return c.GetBlock(hash)
}
/**
*  @brief  create a genesis block with built-in account and contract, then save this block into block chain
*/
func (c *ChainTx) GenesesBlockInit() error {
	if c.CurrentHeader != nil {
		c.CurrentHeader.Show()
		return nil
	}
	//geneses, err := types.GenesesBlockInit()
	block, err := geneses.GenesisBlockInit(c.ledger)
	if err != nil {
		return err
	}
	if err := c.VerifyTxBlock(block); err != nil {
		return err
	}
	c.CurrentHeader = block.Header
	if err := c.SaveBlock(block); err != nil {
		log.Error("Save geneses block error:", err)
		return err
	}
	for i := 0; i < len(block.Transactions); i++ {
		if _, err := c.HandleTransaction(c.ledger, block.Transactions[i]); err != nil {
			log.Error("Handle Transaction Error:", err)
			return err
		}
		//event.Send(event.ActorLedger, event.ActorP2P, txs[i]) //send result to p2p actor
	}
	c.CurrentHeader = block.Header
	return nil
}
/**
*  @brief  restore the highest block's header from levelDB
*  @return bool - if can't find block in levelDB, return false, otherwise return true
*/
func (c *ChainTx) RestoreCurrentHeader() (bool, error) {
	headers, err := c.HeaderStore.SearchAll()
	if err != nil {
		return false, err
	}
	if len(headers) == 0 {
		return false, nil
	}
	log.Info("The geneses block is existed:", len(headers))
	var h uint64 = 0
	for _, v := range headers {
		header := new(types.Header)
		if err := header.Deserialize([]byte(v)); err != nil {
			return false, err
		}
		if header.Height > h {
			h = header.Height
			c.CurrentHeader = header
		}
	}
	log.Info("the block height is:", h, "hash:", c.CurrentHeader.Hash.HexString())
	return true, nil
}
/**
*  @brief  get a transaction from levelDB by a hash
*  @param  key - the hash of transaction
*/
func (c *ChainTx) GetTransaction(key []byte) (*types.Transaction, error) {
	data, err := c.TxsStore.Get(key)
	if err != nil {
		return nil, err
	}
	tx := new(types.Transaction)
	if err := tx.Deserialize(data); err != nil {
		return nil, err
	}
	return tx, nil
}
/**
*  @brief  validity check of transaction, include signature verify, duplicate check and balance check
*  @param  tx - a transaction
*/
func (c *ChainTx) CheckTransaction(tx *types.Transaction) (err error) {
	result, err := tx.VerifySignature()
	if err != nil {
		return err
	} else if result == false {
		return errors.New("tx verify signature failed")
	}
	if err := c.StateDB.CheckPermission(tx.From, tx.Permission, tx.Signatures); err != nil {
		return err
	}

	switch tx.Type {
	case types.TxTransfer:
		if data, _ := c.TxsStore.Get(tx.Hash.Bytes()); data != nil {
			return errs.ErrDuplicatedTx
		}
		if value, err := c.AccountGetBalance(tx.From, state.AbaToken); err != nil {
			return err
		} else if value.Sign() <= 0 {
			log.Error(err)
			return errs.ErrDoubleSpend
		}
	case types.TxDeploy:
		if data, _ := c.TxsStore.Get(common.IndexToBytes(tx.Addr)); data != nil {
			return errs.ErrDuplicatedTx
		}
		//hash := c.StateDB.GetHashRoot()
		//c.HandleTransaction(c, tx)
	case types.TxInvoke:
		if data, _ := c.TxsStore.Get(tx.Hash.Bytes()); data != nil {
			return errs.ErrDuplicatedTx
		}
	default:
		return errors.New("check transaction unknown tx type")
	}

	return nil
}
func (c *ChainTx) CheckPermission(index common.AccountName, name string, sig []common.Signature) error {
	return c.StateDB.CheckPermission(index, name, sig)
}

/**
*  @brief  create a new account in mpt tree
*  @param  index - the uuid of account
*  @param  addr - the public key of account
*/
func (c *ChainTx) AccountAdd(index common.AccountName, addr common.Address) (*state.Account, error) {
	return c.StateDB.AddAccount(index, addr)
}
func (c *ChainTx) StoreSet(index common.AccountName, key, value []byte) (err error) {
	return c.StateDB.StoreSet(index, key, value)
}
func (c *ChainTx) StoreGet(index common.AccountName, key []byte) (value []byte, err error) {
	return c.StateDB.StoreGet(index, key)
}
func (c *ChainTx) SetContract(index common.AccountName, t types.VmType, des, code []byte) error {
	return c.StateDB.SetContract(index, t, des, code)
}
func (c *ChainTx) GetContract(index common.AccountName) (*types.DeployInfo, error) {
	return c.StateDB.GetContract(index)
}
func (c *ChainTx) AddPermission(index common.AccountName, perm state.Permission) error {
	return c.StateDB.AddPermission(index, perm)
}
func (c *ChainTx) FindPermission(index common.AccountName, name string) (string, error) {
	return c.StateDB.FindPermission(index, name)
}
/**
*  @brief  get a account's balance
*  @param  indexAcc - the uuid of account
*  @param  indexToken - the uuid of token
*/
func (c *ChainTx) AccountGetBalance(index common.AccountName, token string) (*big.Int, error) {
	return c.StateDB.GetBalance(index, token)
}
/**
*  @brief  add a account's balance
*  @param  indexAcc - the uuid of account
*  @param  indexToken - the uuid of token
*/
func (c *ChainTx) AccountAddBalance(index common.AccountName, token string, value uint64) error {
	return c.StateDB.AddBalance(index, token, new(big.Int).SetUint64(value))
}
/**
*  @brief  sub a account's balance
*  @param  indexAcc - the uuid of account
*  @param  indexToken - the uuid of token
*/
func (c *ChainTx) AccountSubBalance(index common.AccountName, token string, value uint64) error {
	return c.StateDB.SubBalance(index, token, new(big.Int).SetUint64(value))
}
/**
*  @brief  handle transaction with transaction's type
*  @param  ledger - the interface of ledger impl
*  @param  tx - a transaction
*/
func (c *ChainTx) HandleTransaction(ledger ledger.Ledger, tx *types.Transaction) ([]byte, error) {
	switch tx.Type {
	case types.TxTransfer:
		log.Info("Transfer Execute")
		payload, ok := tx.Payload.GetObject().(types.TransferInfo)
		if !ok {
			return nil, errors.New("transaction type error[transfer]")
		}
		if err := c.AccountSubBalance(tx.From, state.AbaToken, payload.Value.Uint64()); err != nil {
			return nil, err
		}
		if err := c.AccountAddBalance(tx.Addr, state.AbaToken, payload.Value.Uint64()); err != nil {
			return nil, err
		}
	case types.TxDeploy:
		if err := c.CheckPermission(tx.From, state.Active, tx.Signatures); err != nil {
			return nil, err
		}
		payload, ok := tx.Payload.GetObject().(types.DeployInfo)
		if !ok {
			return nil, errors.New("transaction type error[deploy]")
		}
		log.Info("Deploy Execute:", common.ToHex(payload.Code))
		if err := ledger.SetContract(tx.From, payload.TypeVm, payload.Describe, payload.Code); err != nil {
			return nil, err
		}
	case types.TxInvoke:
		log.Info("Invoke Execute")
		service, err := smartcontract.NewContractService(ledger, tx)
		if err != nil {
			return nil, err
		}
		return service.Execute()
	default:
		return nil, errors.New("the transaction's type error")
	}

	return nil, nil
}

func (c *ChainTx) TokenExisted(token string) bool {
	return c.StateDB.TokenExisted(token)
}

func (c *ChainTx) TokenAllocation() error {
	if err := c.StateDB.AddBalance(state.IndexAbaRoot, state.AbaToken, new(big.Int).SetUint64(2100000)); err != nil {
		return err
	}
	return nil
}
