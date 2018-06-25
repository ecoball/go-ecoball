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
	"github.com/ecoball/go-ecoball/common"
	"github.com/ecoball/go-ecoball/common/event"
)

type BlockForest struct {
	size int
	bc *Blockchain
	cache *lru.Cache
}

type linkedBlock struct {
	parentBlock *linkedBlock
	childBlocks map[string]*linkedBlock
	block *DposBlock
	chain *Blockchain
	hash common.Hash
	parentHash common.Hash
}

func NewBlockForest(size int) (*BlockForest, error)  {
	bp := &BlockForest{
		size: size,
	}
	var err error
	bp.cache, err = lru.NewWithEvict(size, func(key interface{}, value interface{}) {
		lb := value.(*linkedBlock)
		if lb != nil {
			lb.Dispose()
		}
	})

	if err != nil {
		return nil, err
	}
	return bp, nil
}

func (bp *BlockForest) Setup(bc *Blockchain)  {
	bp.bc = bc
}

func (lb *linkedBlock) LinkParent(parentBlock *linkedBlock)  {
	lb.parentBlock = parentBlock
	parentBlock.childBlocks[lb.hash.HexString()] = lb
}


func (lb *linkedBlock) travelAndLinkBlockTree(parentBlock *DposBlock) ([]*DposBlock, []*DposBlock, error) {
	if err := lb.block.LinkParentBlock(lb.chain, parentBlock); err != nil {
		log.Error("Failed to link the block with its parent")
		return nil, nil, err
	}

	// TODO, Verify execution

	log.Info("Block Verified.")

	allBlocks := []*DposBlock{lb.block}
	tailBlocks := []*DposBlock{}

	if (len(lb.childBlocks) == 0) {
		tailBlocks = append(tailBlocks, lb.block)
	}

	for _, clb := range lb.childBlocks {
		a, b, err  := clb.travelAndLinkBlockTree(lb.block)
		if err == nil {
			allBlocks = append(allBlocks, a...)
			tailBlocks = append(tailBlocks, b...)
		}
	}

	return allBlocks, tailBlocks, nil
}

const (
	NoSender = ""
)

func (forest *BlockForest) PushAndBroadcast(block *DposBlock) error {

	//TODO, broadcast

	//Push
	return forest.push(NoSender, block)
}

func (forest *BlockForest)  push(sender string, block *DposBlock) error {
	//verify non-dup block
	if forest.cache.Contains(block.Hash.HexString()) ||
		forest.bc.GetBlock(block.Hash) != nil {
		log.Debug("Found duplicated block.")
		return ErrDuplicatedBlock
	}

	//verify block integrity
	if err := block.VerifyIntegrity(); err != nil {
		log.Debug("Failed to check block integrity")
		return err
	}

	bc := forest.bc
	cache := forest.cache

	var plb *linkedBlock
	lb := newLinkedBlock(block, forest.bc)
	cache.Add(lb.hash.HexString(), lb)

	// Find child blocks in forest cache
	for _, k := range cache.Keys() {
		v, _ := cache.Get(k)
		c := v.(*linkedBlock)
		if c.parentHash.Equals(&lb.hash) {
			c.LinkParent(lb)
		}
	}

	unsyncCnt := 0
	v, _ := cache.Get(lb.parentHash.HexString())
	if v != nil {
		plb = v.(*linkedBlock)
		lb.LinkParent(plb)
		lb = plb
		unsyncCnt++

		for lb.parentBlock != nil {
			lb = lb.parentBlock
			unsyncCnt++
		}

		log.Warn("Found unlinked ancestor.")

		//TODO
		if sender == NoSender {
			return ErrMissingParentBlock
		}
	}

	// find parent in Chain
	var parentBlock *DposBlock
	if parentBlock = bc.GetBlock(lb.parentHash); parentBlock == nil {
		// still not found, wait to parent block from network
		if sender == NoSender {
			return ErrMissingParentBlock
		}

		// do sync if too many in cache
		//TODO

		if (unsyncCnt > 16) {
			if bc.StartActiveSync() {
				log.Warn("unsync too much, pending accouting and sync from others")
			}
			return ErrSyncParent
		}

		if err := forest.getParent(sender, lb.block); err != nil {
			return err
		}
		return ErrSyncParent
	}

	//TODO
	if sender != NoSender {
		event.Send(0, event.ActorP2P, block)
	}

	allBlocks, tailBlocks, err := lb.travelAndLinkBlockTree(parentBlock)
	//TODO, Not sure why remove only cache tree root
	if err != nil {
		cache.Remove(lb.hash.HexString())
		return err
	}

	if err := bc.putVerifiedNewBlocks(parentBlock, allBlocks, tailBlocks); err != nil {
		cache.Remove(lb.hash.HexString())
		return err
	}

	for _, v := range allBlocks {
		cache.Remove(v.Hash.HexString())
	}

	return forest.bc.ConsensusHandler().DealWithFork()

}

//TODO
func (forest *BlockForest) getParent(sender string, block *DposBlock) error {
	return nil
}

func newLinkedBlock(block *DposBlock, chain *Blockchain) *linkedBlock {
	return &linkedBlock{
		block: block,
		chain: chain,
		hash: block.Hash,
		parentHash: block.Header.PrevHash,
		parentBlock: nil,
		childBlocks: make(map[string]*linkedBlock),
	}
}

func (lb *linkedBlock) Dispose() {
	lb.block = nil
	lb.chain = nil

	for _, v := range lb.childBlocks {
		v.parentBlock = nil
	}
	lb.childBlocks = nil

	if lb.parentBlock != nil  {
		delete(lb.parentBlock.childBlocks, lb.hash.HexString())
		lb.parentBlock = nil
	}

}