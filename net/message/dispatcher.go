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

package message

import (
	"github.com/ecoball/go-ecoball/core/types"
	eactor "github.com/ecoball/go-ecoball/common/event"
	"github.com/ecoball/go-ecoball/consensus/ababft"
)

func HdTransactionMsg(data []byte) error {
	tx := new(types.Transaction)
	err := tx.Deserialize(data)
	if err != nil {
		return err
	}
	log.Debug("dispatch tx msg")
	eactor.Send(0, eactor.ActorTxPool, tx)
	return  nil
}

func HdBlkMsg(data []byte) error {
	blk := new(types.Block)
	err := blk.Deserialize(data)
	if err != nil {
		return err
	}
	log.Debug("dispatch blk msg")
	eactor.Send(0, eactor.ActorLedger, blk)
	return nil
}

func HdGossipBlkReqMsg(data []byte) error {
	blkReq := new(types.BlkReqMsg)
	blkReq.Deserialize(data)
	eactor.Send(0, eactor.ActorP2P, blkReq)
	return nil
}

func HdGossipBlkAckMsg(data []byte) error {
	blkAck := new(types.BlkAckMsg)
	blkAck.Deserialize(data)
	eactor.Send(0, eactor.ActorP2P, blkAck)
	return nil
}

func HdGossipBlkAck2Msg(data []byte) error {
	blkAck2 := new(types.BlkAck2Msg)
	blkAck2.Deserialize(data)
	eactor.Send(0, eactor.ActorP2P, blkAck2)
	return nil
}

func HdSignPreMsg(data []byte) error {
	signpre_receive := new(ababft.Signature_Preblock)
	err := signpre_receive.Deserialize(data)
	if err != nil {
		return err
	}
	log.Debug("dispatch signpre msg")
	eactor.Send(0, eactor.ActorConsensus, signpre_receive)
	return nil
}

func HdBlkFMsg(data []byte) error {
	block_firstround := new(ababft.Block_FirstRound)
	err := block_firstround.Blockfirst.Deserialize(data)
	if err != nil {
		return err
	}
	log.Debug("dispatch first round block msg")
	eactor.Send(0, eactor.ActorConsensus, block_firstround)
	return nil
}

func HdReqSynMsg(data []byte) error {
	reqsyn := new(ababft.REQSyn)
	err := reqsyn.Deserialize(data)
	if err != nil {
		return err
	}
	log.Debug("dispatch synchronization request msg")
	eactor.Send(0, eactor.ActorConsensus, reqsyn)
	return nil
}

func HdToutMsg(data []byte) error {
	toutmsg := new(ababft.TimeoutMsg)
	err := toutmsg.Deserialize(data)
	if err != nil {
		return err
	}
	log.Debug("dispatch synchronization request msg")
	eactor.Send(0, eactor.ActorConsensus, toutmsg)
	return nil
}

func HdSignBlkFMsg(data []byte) error {
	signblkf_receive := new(ababft.Signature_BlkF)
	err := signblkf_receive.Deserialize(data)
	if err != nil {
		return err
	}
	log.Debug("dispatch the signature of first-round block msg")
	eactor.Send(0, eactor.ActorConsensus, signblkf_receive)
	return nil
}

/*






func HdBlkSMsg(data []byte) error {
	block_secondround := new(ababft.Block_SecondRound)
	err := block_secondround.Blocksecond.Deserialize(data)
	if err != nil {
		return err
	}
	log.Debug("dispatch second-round(final) block msg")
	eactor.Send(0, eactor.ActorConsensus, block_secondround)
	return nil
}



func HdBlkSynMsg(data []byte) error {
	blksyn := new(ababft.Block_Syn)
	err := blksyn.Blksyn.Deserialize(data)
	if err != nil {
		return err
	}
	log.Debug("dispatch the block according to the synchronization request")
	eactor.Send(0, eactor.ActorConsensus, blksyn)
	return nil
}


 */

// MakeHandlers generates a map of MsgTypes to their corresponding handler functions
func MakeHandlers() map[uint32]HandlerFunc {
	return map[uint32]HandlerFunc{
		APP_MSG_TRN:     HdTransactionMsg,
		APP_MSG_BLK:     HdBlkMsg,
		APP_MSG_GOSSIP_PULL_BLK_REQ: HdGossipBlkReqMsg,
		APP_MSG_GOSSIP_PULL_BLK_ACK: HdGossipBlkAckMsg,
		APP_MSG_GOSSIP_PUSH_BLKS:    HdGossipBlkAck2Msg,
		APP_MSG_SIGNPRE:   HdSignPreMsg,
		APP_MSG_BLKF:      HdBlkFMsg,
		APP_MSG_REQSYN:    HdReqSynMsg,
		/*
		APP_MSG_SIGNBLKF:  HdSignBlkFMsg,
		APP_MSG_BLKS:      HdBlkSMsg,

		APP_MSG_BLKSYN:    HdBlkSynMsg,
		APP_MSG_TIMEOUT:   HdToutMsg,
		*/
		//TODO add new msg handler at here
	}
}