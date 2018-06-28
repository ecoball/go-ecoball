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
	"github.com/AsynkronIT/protoactor-go/actor"
	actorTypes "github.com/ecoball/go-ecoball/consensus/dpos/actor"
	"fmt"
	"time"
	"errors"
	"github.com/ecoball/go-ecoball/common"
	"github.com/ecoball/go-ecoball/common/event"
	"github.com/ecoball/go-ecoball/common/message"
	"github.com/ecoball/go-ecoball/core/types"
	"github.com/ecoball/go-ecoball/account"
	"reflect"
	"github.com/ecoball/go-ecoball/core/ledgerimpl/ledger"
)


var (
	ErrInvalidLeader              = errors.New("invalid leader")
	ErrCannotForgeWhenPending   = errors.New("cannot forge block now, waiting for cancel pending again")
	ErrCannotForgeWhenDisable   = errors.New("cannot forge block now, waiting for enable it again")
	ErrWaitingBlockInLastTimePoint     = errors.New("cannot forge block now, waiting for last block")
	ErrBlockForgedInNextTimePoint      = errors.New("cannot forge block now, there is a block produced in current TimePoint")
	ErrGenerateNextConsensusState = errors.New("Failed to generate next consensus state")
	ErrAppendNewBlockFailed       = errors.New("failed to append new block to real chain")

	ErrMissingParentBlock = errors.New("cannot find the block's parent block in storage")
	ErrSyncParent         = errors.New("floating block received, sync its parent from others")
	ErrDuplicatedBlock    = errors.New("DuplicatedBlock")
	ErrCannotRevertLIB    = errors.New("Cannot revert LIB")
	)

type DposService struct {

	exitChan chan bool

	chain *Blockchain

	pid *actor.PID

	bookkeeper common.Hash
	//TODO, init account
    account *account.Account

	enable bool
	pending bool

	ledger ledger.Ledger
}

func NewDposService() (*DposService, error)  {
	service := &DposService{
		exitChan: make(chan bool, 5),
		enable: false,
		pending: false,
	}

	props := actor.FromProducer(func() actor.Actor {
		return service
	})

	pid, err := actor.SpawnNamed(props, "consensus-dpos")
	service.pid = pid

	if err != nil {
		return nil, err
	}

	return service, nil
}

//TODO
func (dpos *DposService) Setup(bc *Blockchain, ledger ledger.Ledger)  {
	dpos.chain = bc
	addr1 := common.Address{1,2,3,4,5,6,7,8,9,10,11,12,13,1,2,3,4,5,6,7}
	s1 := addr1.ToBase58()
	hash := common.NewHash(common.AddressFromBase58(s1).Bytes())
	dpos.bookkeeper = hash
	//TODO
	acc, _ := account.NewAccount(0)
	dpos.account = &acc

	dpos.enable = true

	dpos.ledger = ledger
}

func (dpos *DposService) Start()  {
	log.Info("Starting Dpos Accouting")
	go dpos.working()
}

func (dpos *DposService) Stop()  {
	log.Info("Stopping Dpos Accouting...")
	dpos.DisableAccouting()
	dpos.exitChan <- true
}

//TODO
func (dpos *DposService) DisableAccouting() error {
	dpos.enable = false
	return nil
}

func (dpos *DposService) working()  {
	log.Info("Start Dpos Accouting")
	chronicChan := time.NewTicker(time.Second).C
	for {
		select {
		case now := <-chronicChan:
			dpos.forgeBlock(now.Unix())
		case <- dpos.exitChan:
			log.Info("Stop Dpos Accouting")
			return
		}
	}
}

//TODO
func (dpos *DposService) UpdateLIB()  {

}

func (dpos *DposService) forgeBlock(now int64) error {

	log.Debug("In forge block")

	nowInMs := now * Second

	if !dpos.enable {
		return ErrCannotForgeWhenDisable
	}

	if dpos.pending {
		return ErrCannotForgeWhenPending
	}

	tail := dpos.chain.TailBlock()

	deadlineInMs, err := dpos.calculateDeadline(tail, nowInMs)
	if err != nil {
		return err
	}

	consensusState, err := dpos.isLeader(tail, nowInMs)
	if err != nil {
		log.Error(err)
		return err
	}

	bookkeeper := "nil"
	bookkeeper = dpos.bookkeeper.HexString()
	log.Info("My turn to forge block ", bookkeeper)

	block, err := dpos.newBlock(tail, consensusState, deadlineInMs)
	if err != nil {
		return err
	}

	if err := dpos.pushAndBroadcast(tail, block); err != nil {
		go block.ReturnTransactions()
		return err
	}

	return nil
}

func (dpos *DposService) pushAndBroadcast(tail *DposBlock, block *DposBlock) error  {
	if err := dpos.chain.BlockPool().PushAndBroadcast(block); err != nil {
		log.Error("Failed to push new block into block pool")
		return err
	}

	//TODO, under what situation this will happen?
	if !dpos.chain.TailBlock().Hash.Equals(&block.Hash) {
		return ErrAppendNewBlockFailed
	}

	log.Info("Broadcasted new block")
	return nil
}

func (dpos *DposService) isLeader(tail *DposBlock, nowInMs int64) (ConsensusState, error) {
	pointInMs := nextChance(nowInMs)
	elapsedInMs := pointInMs - tail.TimeStamp *Second
	consensusState, err := tail.state.NextConsensusState(elapsedInMs / Second)

	if err != nil {
		log.Debug("Failed to generate next consensus state", err)
		return nil, ErrGenerateNextConsensusState
	}
	//TODO, check nil
	log.Debug("Partial Success")

	leader := consensusState.Leader()
	if !(&leader).Equals(&dpos.bookkeeper) {

		log.Debug("No my turn, waiting, I'm %s, but actual leader is %s", dpos.bookkeeper, leader)
		return nil, ErrInvalidLeader
	}
	return consensusState, nil
}

func (dpos *DposService) calculateDeadline(tail *DposBlock, nowInMs int64) (int64, error){
	lastPoint := lastChance(nowInMs)
	nextPoint := nextChance(nowInMs)

	if tail.Timestamp() *Second >= nextPoint {
		return 0, ErrBlockForgedInNextTimePoint
	}
	if tail.Timestamp() *Second == lastPoint {
		return deadline(nowInMs), nil
	}
	if nextPoint - nowInMs <= MinProduceDuration {
		return deadline(nowInMs), nil
	}
	return 0, ErrWaitingBlockInLastTimePoint

}

func deadline(nowInMs int64) int64 {
	nextTimePointInMs := nextChance(nowInMs)
	remainInMs := nextTimePointInMs - nowInMs
	if MaxProduceDuration > remainInMs {
		return nextTimePointInMs
	}
	return nowInMs + MaxProduceDuration
}


func lastChance(nowInMs int64) int64 {
	return int64((nowInMs-Second)/BlockInterval) * BlockInterval
}

func nextChance(nowInMs int64) int64 {
	return int64((nowInMs+BlockInterval-Second)/BlockInterval) * BlockInterval
}

func (dpos *DposService) Receive(context actor.Context) {
	switch msg := context.Message().(type) {
	case *actorTypes.StartConsensus:
		fmt.Printf("Hello %v\n", msg)
	}
}

func (dpos *DposService) newBlock(tail *DposBlock, consensusState ConsensusState, deadlineInMs int64) (*DposBlock, error) {
	startAt := time.Now().Unix()

	secondInMs := int64(1000)
	elapseInMs := deadlineInMs - time.Now().Unix() * secondInMs
	log.Info("Time to pack txs.", elapseInMs)

	if elapseInMs <= 0 {
		return nil, ErrTimeNegative
	}
	deadlineTimer := time.NewTimer(time.Duration(elapseInMs) * time.Millisecond)
	<-deadlineTimer.C

	value, err := event.SendSync(event.ActorTxPool, message.GetTxs{}, time.Microsecond * 1000)
	log.Debug("value type = ", reflect.TypeOf(value))
	txList, ok := value.(*types.TxsList)

	if !ok {
		log.Debug("type error")
		return nil, ErrTypeWrong
	}
	var txs []*types.Transaction
	for _, v := range txList.Txs {
		log.Debug(v.Hash.HexString())
		txs = append(txs, v)
	}

	conData := types.ConsensusData{Type:types.ConSolo, Payload:&types.SoloData{}}
	block, err := dpos.chain.chainTx.NewBlock(dpos.ledger ,txs, conData)
	if err != nil {
		log.Error("Failed to create new block")
		return nil, err
	}

	dposBlock := &DposBlock {
		block,
		consensusState,
	}

	//TODO, remote sign


	if err = dposBlock.Pack(); err != nil {
		log.Error("Failed to seal new block")
		go dposBlock.ReturnTransactions()
		return nil, err
	}

	//TODO, make sure it's right
	err = dposBlock.SetSignature(dpos.account)

	if err != nil {
		log.Error("Failed to sign new block")
		go dposBlock.ReturnTransactions()
		return nil, err
	}
	endAt := time.Now().Unix()

	log.Debug("Packed txs.", startAt, endAt)

	return dposBlock, nil

}

func (dpos DposService) VerifyBlock(block *DposBlock) error {
	//TODO
	return nil
}

func verifyBlockSign(bookkeeper *common.Address, block *DposBlock)  error {
	//TODO
	return nil
}

func less(a *DposBlock, b *DposBlock) bool {
	if a.Height != b.Height {
		return a.Height < b.Height
	}
	return a.Hash.HexString() < b.Hash.HexString()
}

func (dpos *DposService) DealWithFork() error  {
	bc := dpos.chain
	tailBlock := bc.TailBlock()
	detachedTailBlocks := bc.DetachedTailBlocks()

	// Find the max height
	newTailBlock := tailBlock

	for _, v := range detachedTailBlocks {
		if less(newTailBlock, v) {
			newTailBlock = v
		}
	}

	if newTailBlock.Hash.Equals(&tailBlock.Hash) {
		log.Debug("Current tail is best, no need to change")
		return nil
	}

	err := bc.SetTailBlock(newTailBlock)
	if err != nil {
		log.Debug("Failed to set new tail block")
		return err
	}

	log.Info("change to new tail")
	return nil
}