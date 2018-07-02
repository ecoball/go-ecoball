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

package types

import (
	"errors"
	"github.com/ecoball/go-ecoball/core/pb"
	"gx/ipfs/QmcJukH2sAFjY3HdBKq35WDzWoL3UUu2gt9wdfqZTUyM74/go-libp2p-peer"

)

type BlkReqMsg struct {
	Peer      peer.ID
	ChainID   uint32
	BlkHeight uint64
}

type BlkAckMsg struct {
	Peer      peer.ID
	ChainID   uint32
	BlkHeight uint64
	BlkCount  uint64
	Data      []*Block
}

type BlkAck2Msg struct {
	ChainID  uint32
	BlkCount uint64
	Data     []*Block
}

func (blkReq *BlkReqMsg)Serialize() ([]byte, error) {
	p := &pb.PullBlocksRequest{
		PeerHash:   []byte(blkReq.Peer),
		ChainId:    blkReq.ChainID,
		Height:     blkReq.BlkHeight,
	}
	b, err := p.Marshal()
	if err != nil {
		return nil, err
	}
	return b, nil
}

func (blkReq *BlkReqMsg)Deserialize(data []byte) error {
	if len(data) == 0 {
		return errors.New("input data's length is zero")
	}
	var blkRequest pb.PullBlocksRequest
	if err := blkRequest.Unmarshal(data); err != nil {
		return err
	}
	blkReq.Peer = peer.ID(blkRequest.PeerHash)
	blkReq.ChainID = blkRequest.ChainId
	blkReq.BlkHeight = blkRequest.Height

	return nil
}

func (blkAck *BlkAckMsg)Serialize() ([]byte, error) {
	p := &pb.PullBlocksResponse{
		PeerHash:   []byte(blkAck.Peer),
		ChainId:    blkAck.ChainID,
		Height:     blkAck.BlkHeight,
		BlockCount: blkAck.BlkCount,
	}
	var pbBlks []*pb.BlockTx
	for _, blk := range blkAck.Data {
		pbBlk, err := blk.protoBuf()
		if err != nil {
			return nil, err
		}
		pbBlks = append(pbBlks, pbBlk)
	}
	p.Data = append(p.Data, pbBlks...)

	b, err := p.Marshal()
	if err != nil {
		return nil, err
	}
	return b, nil
}

func (blkAck *BlkAckMsg)Deserialize(data []byte) error {
	if len(data) == 0 {
		return errors.New("input data's length is zero")
	}
	var pbBlkAck pb.PullBlocksResponse
	if err := pbBlkAck.Unmarshal(data); err != nil {
		return err
	}

	blkAck.Peer = peer.ID(pbBlkAck.PeerHash)
	blkAck.ChainID = pbBlkAck.ChainId
	blkAck.BlkHeight = pbBlkAck.Height
	blkAck.BlkCount = pbBlkAck.BlockCount

	var blks []*Block
	for _, blk := range pbBlkAck.Data {
		data, err := blk.Marshal()
		if err != nil {
			return err
		}
		b := new(Block)
		if err := b.Deserialize(data); err != nil {
			return err
		}
		blks = append(blks, b)
	}
	blkAck.Data = blks
	return nil
}

func (blkAck2 *BlkAck2Msg)Serialize() ([]byte, error) {
	p := &pb.PushBlocks{
		ChainId:    blkAck2.ChainID,
		BlockCount: blkAck2.BlkCount,
	}
	var pbBlks []*pb.BlockTx
	for _, blk := range blkAck2.Data {
		pbBlk, err := blk.protoBuf()
		if err != nil {
			return nil, err
		}
		pbBlks = append(pbBlks, pbBlk)
	}
	p.Data = append(p.Data, pbBlks...)

	b, err := p.Marshal()
	if err != nil {
		return nil, err
	}
	return b, nil
}

func (blkAck2 *BlkAck2Msg)Deserialize(data []byte) error {
	if len(data) == 0 {
		return errors.New("input data's length is zero")
	}
	var pbPushBlks pb.PushBlocks
	if err := pbPushBlks.Unmarshal(data); err != nil {
		return err
	}

	blkAck2.ChainID = pbPushBlks.ChainId
	blkAck2.BlkCount = pbPushBlks.BlockCount

	var blks []*Block
	for _, blk := range pbPushBlks.Data {
		data, err := blk.Marshal()
		if err != nil {
			return err
		}
		b := new(Block)
		if err := b.Deserialize(data); err != nil {
			return err
		}
		blks = append(blks, b)
	}
	blkAck2.Data = blks
	return nil
}