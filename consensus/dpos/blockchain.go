// Copyright 2018 The go-ecoball Authors
// This file is part of the go-ecoball.
//
// The go-ecoball is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ecoball is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ecoball. If not, see <http://www.gnu.org/licenses/>.


package dpos

import (
	"github.com/hashicorp/golang-lru"

	"time"
	"github.com/ecoball/go-ecoball/common"

)

type Blockchain struct {

	genesisBlock *DposBlock
	tailBlock *DposBlock
	//TODO, consensusHandler not nessary?
	consensusHandler *DposService

	cachedBlocks *lru.Cache
	blockPool *BlockForest

	//TODO, use Cache to sore this Date structure will cause problem? (Lost some data)
	detachedTailBlocks *lru.Cache

	lib *DposBlock


	quitCh chan int
	chainTx ChainInterface



}

const (

	Tail = "tail"

)

func NewBlockChain(chainTx ChainInterface) (*Blockchain, error)  {

	blockPool, err := NewBlockForest(128)
	if err != nil {
		return nil, err
	}

	var bc = &Blockchain{
		genesisBlock:       nil,
		tailBlock:          nil,
		consensusHandler:   nil,
		cachedBlocks:       nil,
		blockPool:          blockPool,
		detachedTailBlocks: nil,
		lib:                nil,
		quitCh:             nil,
		chainTx:            chainTx,
	}

	//TODO, think the situation when instances blow out the cache
	bc.cachedBlocks, err = lru.New(128)
	if err != nil {
		return nil, err
	}

	bc.detachedTailBlocks, err = lru.New(128)
	if err != nil {
		return nil, err
	}

	return bc, nil
}


//TODO, loop maybe?
func (bc *Blockchain) Start() {
	log.Info("Starting BlockChain...")
}

func (bc *Blockchain) Setup(dpos *DposService) error  {
	//TODO
	var err error
	bc.consensusHandler = dpos

	if err != nil {
		log.Info("NewTransactionChain error")
		return err
	}

	bc.tailBlock = bc.GetBlock(bc.chainTx.GetTailBlockHash())

	bc.blockPool.Setup(bc)

	return nil
}

func (bc *Blockchain) TailBlock() *DposBlock {
	return bc.tailBlock
}

func (bc *Blockchain) BlockPool() *BlockForest {
	return bc.blockPool
}

func (bc *Blockchain) ConsensusHandler() ForkConsensus {
	return bc.consensusHandler
}

func (bc *Blockchain) putVerifiedNewBlocks(parent *DposBlock, allBlocks, tailBlocks []*DposBlock) error {
	for _, v := range allBlocks {
		bc.cachedBlocks.Add(v.Hash.HexString(), v)
		if err := bc.StoreBlockToStorage(v); err != nil {
			log.Debug("Failed to store the verified block.")
			return err
		}
	}

	log.Info("Accepted the new block on chain")

	for _, v := range tailBlocks {
		bc.detachedTailBlocks.Add(v.Hash.HexString(), v)
	}
	bc.detachedTailBlocks.Remove(parent.Hash.HexString())

	return nil
}

func (bc *Blockchain) SetTailBlock(newTail *DposBlock) error {
	if newTail == nil {
		return ErrNilArgument
	}
	oldTail := bc.tailBlock
	ancestor, err := bc.FindLowestCommonAncestorWithTail(newTail)
	if err != nil {
		log.Debug("Failed to find common ancestor with tail")
		return err
	}

	if err := bc.revertBlocks(ancestor, oldTail); err != nil {
		log.Debug("Failed to revert blocks")
		return err
	}

	if err := bc.buildIndexByBlockHeight(ancestor, newTail); err != nil {
		log.Debug("Failed to build index by block height.")
		return err
	}

	if err := bc.StoreTailHashToStorage(newTail); err != nil {
		return err
	}
	bc.tailBlock = newTail

	log.Info("Succeed to update new tail.")

	return nil
}

func (bc *Blockchain) buildIndexByBlockHeight(from *DposBlock, to *DposBlock) error {
	blocks := []*DposBlock{}
	for !to.Hash.Equals(&from.Hash) {
		//TODO
		/*err := bc.storage.Put(byteutils.FromUint64(to.height), to.Hash())
		if err != nil {
			return err
		}*/
		blocks = append(blocks, to)
		go bc.dropTxsInBlockFromTxPool(to)
		if to == nil {
			return ErrMissingParentBlock
		}

		return nil
	}
	return nil
}

//TODO
func (bc *Blockchain) dropTxsInBlockFromTxPool(block *DposBlock)  {
	/*for _, tx := range block.Transactions {
		bc.txPool.Del(tx)
	}*/
}


/*
return the block's lowest common ancestor with current tail
param block is the new tail, has bigger height
   */

func (bc *Blockchain) FindLowestCommonAncestorWithTail(block *DposBlock) (*DposBlock, error) {
	if block == nil {
		return nil, ErrNilArgument
	}
	target := bc.GetBlock(block.Hash)
	if target == nil {
		return nil, ErrMissingParentBlock
	}

	tail := bc.tailBlock

	for tail.Header.Height < target.Header.Height {
		target = bc.GetBlock(target.PrevHash)
		if target == nil {
			return nil, ErrMissingParentBlock
		}
	}

	for !tail.Hash.Equals(&target.Hash) {
		tail = bc.GetBlock(tail.PrevHash)
		target = bc.GetBlock(target.PrevHash)
		if tail == nil || target == nil {
			return nil, ErrMissingParentBlock
		}
	}

	return target, nil
}

func (bc *Blockchain) GetBlock(hash common.Hash) *DposBlock {
	v, _ := bc.cachedBlocks.Get(hash.HexString())
	if v == nil {
		block, err := LoadBlockFromStorage(hash, bc)
		if err != nil {
			return nil
		}
		return block
	}

	block := v.(*DposBlock)
	return block
}

//TODO, asychronized call for save function, not good
func (bc *Blockchain) StoreBlockToStorage(block *DposBlock) error {
	//event.Send(0, event.ActorLedger, block)
	bc.chainTx.SaveBlock(block.Block)
	bc.chainTx.SaveConsensusState(block)
	return nil
}


func (bc *Blockchain) revertBlocks(from *DposBlock, to *DposBlock) error {
	reverted := to
	var revertTimes int64
	blocks := []string{}
	for revertTimes = 0; !reverted.Hash.Equals(&from.Hash); {
		if reverted.Hash.Equals(&bc.lib.Hash) {
			return ErrCannotRevertLIB
		}

		reverted.ReturnTransactions()

		log.Warn("A block is reverted.")
		revertTimes++
		blocks = append(blocks, reverted.String())

		reverted = bc.GetBlock(reverted.Header.PrevHash)
		if reverted == nil {
			return ErrMissingParentBlock
		}
	}

	go bc.triggerRevertBlockEvent(blocks)

	return nil
}

func (bc *Blockchain) triggerRevertBlockEvent(blocks []string) {

}

func (bc *Blockchain) StoreTailHashToStorage(block *DposBlock) error {
	//TODO
	/*
	return bc.storage.Put([]byte(Tail), block.Hash())
	 */
	 return nil
}

// LIB return the latest irrversible block
func (bc *Blockchain) LIB() *DposBlock {
	return bc.lib
}

// SetLIB update the latest irrversible block
func (bc *Blockchain) SetLIB(lib *DposBlock) {
	bc.lib = lib
}

func (bc *Blockchain) loop()  {
	log.Info("Started BlockChain.")
	timerChan := time.NewTicker(15 * time.Second).C
	for {
		select {
		case <- bc.quitCh:
			log.Info("Stopped BlockChain")
		return
		case <-timerChan:
			bc.ConsensusHandler().UpdateLIB()
		}
	}
}

//TODO
func (bc *Blockchain) StartActiveSync() bool {
	return true
}

func (bc *Blockchain) DetachedTailBlocks() []*DposBlock {
	ret := make([]*DposBlock, 0)
	for _, k := range bc.detachedTailBlocks.Keys() {
		v, _ := bc.detachedTailBlocks.Get(k)
		if v != nil {
			block := v.(*DposBlock)
			ret = append(ret, block)
		}
	}
	return ret
}


func (bc *Blockchain) SaveBlock(block *DposBlock) error {
	err := bc.chainTx.SaveBlock(block.Block)
	if err != nil {
		return err
	}
	err = bc.chainTx.SaveConsensusState(block)
	if err != nil {
		return err
	}
	return nil
}


