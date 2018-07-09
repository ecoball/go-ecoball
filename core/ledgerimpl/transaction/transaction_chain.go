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
	c = &ChainTx{CurrentHeader: &types.Header{}, ledger: ledger}
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

	f, err := c.RestoreBlock()
	if err != nil {
		return nil, err
	}
	if c.StateDB, err = state.NewState(path+config.StringState, c.CurrentHeader.StateHash); err != nil {
		return nil, err
	}
	if err := c.StateDB.AddAccount(state.IndexAbaRoot, common.Address{}); err != nil {
		return nil, err
	}
	if err := c.StateDB.AddBalance(state.IndexAbaRoot, state.IndexAbaToken, new(big.Int).SetUint64(2100000)); err != nil {
		return nil, err
	}
	if f == false {
		if err := c.GenesesBlockInit(); err != nil {
			return nil, err
		}
	}

	return c, nil
}

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
** If create a new block failed, then need to reset state DB
 */
func (c *ChainTx) ResetStateDB(hash common.Hash) error {
	return c.StateDB.Reset(hash)
}


func (c *ChainTx) SaveBlock(block *types.Block) error {
	if block == nil {
		return errors.New("block is nil")
	}
	block.Show()
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

func (c *ChainTx) GetTailBlockHash() common.Hash {
	return c.CurrentHeader.Hash
}

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

func (c *ChainTx) GenesesBlockInit() error {
	geneses, err := types.GenesesBlockInit()
	if err != nil {
		return err
	}
	c.CurrentHeader = geneses.Header
	if err := c.SaveBlock(geneses); err != nil {
		log.Error("Save geneses block error:", err)
		return err
	}
	c.CurrentHeader = geneses.Header
	return nil
}

func (c *ChainTx) RestoreBlock() (bool, error) {
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

func (c *ChainTx) CheckTransaction(tx *types.Transaction) (err error) {
	var data []byte
	switch tx.Type {
	case types.TxTransfer:
		if data, err = c.TxsStore.Get(tx.Hash.Bytes()); err != nil {
			return err
		} else {
			if data != nil {
				return errs.ErrDuplicatedTx
			}
		}
		if value, err := c.AccountGetBalance(tx.From, state.IndexAbaToken); err != nil {
			return err
		} else if value.Sign() <= 0 {
			return errs.ErrDoubleSpend
		}
	case types.TxDeploy:
		if data, err = c.TxsStore.Get(common.IndexToBytes(tx.Addr)); err != nil {
			return err
		} else if data != nil {
			return errs.ErrDuplicatedTx
		}
		//hash := c.StateDB.GetHashRoot()
		//c.HandleTransaction(c, tx)
	case types.TxInvoke:
		data, err = c.TxsStore.Get(tx.Hash.Bytes())
	default:
		return errors.New("check transaction unknown tx type")
	}

	return errs.ErrNoError
}

func (c *ChainTx) AccountGetBalance(indexAcc, indexToken uint64) (*big.Int, error) {
	return c.StateDB.GetBalance(indexAcc, indexToken)
}

func (c *ChainTx) AccountAddBalance(indexAcc, indexToken uint64, value uint64) error {
	return c.StateDB.AddBalance(indexAcc, indexToken, new(big.Int).SetUint64(value))
}

func (c *ChainTx) AccountSubBalance(indexAcc, indexToken uint64, value uint64) error {
	return c.StateDB.SubBalance(indexAcc, indexToken, new(big.Int).SetUint64(value))
}

func (c *ChainTx) HandleTransaction(ledger ledger.Ledger, tx *types.Transaction) ([]byte, error) {
	tx.Show()
	switch tx.Type {
	case types.TxTransfer:
		log.Info("Execute Transfer")
		payload, ok := tx.Payload.GetObject().(types.TransferInfo)
		if !ok {
			return nil, errors.New("transaction type error[transfer]")
		}
		if err := c.AccountSubBalance(tx.From, state.IndexAbaToken, payload.Value.Uint64()); err != nil {
			return nil, err
		}
		if err := c.AccountAddBalance(tx.Addr, state.IndexAbaToken, payload.Value.Uint64()); err != nil {
			return nil, err
		}
	case types.TxDeploy:
		payload, ok := tx.Payload.GetObject().(types.DeployInfo)
		if !ok {
			return nil, errors.New("transaction type error[deploy]")
		}
		log.Info("Deploy Execute:", common.ToHex(payload.Code))
	case types.TxInvoke:
		log.Info("InvokeInfo Execute()")
		payload, ok := tx.Payload.GetObject().(types.InvokeInfo)
		if !ok {
			return nil, errors.New("transaction type error[invoke]")
		}
		data, err := c.TxsStore.Get(common.IndexToBytes(tx.Addr))
		if err != nil {
			return nil, err
		}
		txDeploy := &types.Transaction{Payload: &types.DeployInfo{}}
		if err := txDeploy.Deserialize(data); err != nil {
			return nil, err
		}
		txDeploy.Show()
		deployInfo, ok := txDeploy.Payload.GetObject().(types.DeployInfo)
		if !ok {
			return nil, errors.New(fmt.Sprintf("can't find the deploy contract:%s", common.IndexToName(tx.Addr)))
		}
		fmt.Println("execute code:", common.ToHex(deployInfo.Code))
		fmt.Println("method:", string(payload.Method))
		fmt.Println("param:", payload.Param)
		service, err := smartcontract.NewContractService(ledger, tx)
		if err != nil {
			return nil, err
		}
		return service.ExecuteContract(payload.TypeVm, string(payload.Method), deployInfo.Code, payload.Param)
	default:
		return nil, errors.New("the transaction's type error")
	}

	return nil, nil
}

func (c *ChainTx) TokenExisted(indexToken uint64) bool {
	return c.StateDB.TokenExisted(indexToken)
}
