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
//
// The following is the ababft consensus algorithm.
// Author: Xu Wang, 2018.07.16

package ababft

import (
	"github.com/ecoball/go-ecoball/core/types"
	"github.com/ecoball/go-ecoball/core/pb"
	"github.com/syndtr/goleveldb/leveldb/errors"
)

type Block_FirstRound struct {
	Blockfirst types.Block
}

type Block_SecondRound struct {
	Blocksecond *types.Block
}

type Signature_Preblock struct {
	Signature_preblock pb.SignaturePreblock
}

func (sign *Signature_Preblock) Serialize() ([]byte, error) {
	b, err := sign.Signature_preblock.Marshal()
	if err != nil {
		return nil, err
	}
	return b, nil
}

func (sign *Signature_Preblock) Deserialize(data []byte) error {
	if len(data) == 0 {
		return errors.New("input data's length is zero")
	}

	if err := sign.Signature_preblock.Unmarshal(data); err != nil {
		return err
	}
	return nil
}

type REQSyn struct {
	Reqsyn *pb.RequestSyn
}

func (reqsyn *REQSyn) Serialize() ([]byte, error) {
	b, err := reqsyn.Reqsyn.Marshal()
	if err != nil {
		return nil, err
	}
	return b, nil
}

func (reqsyn *REQSyn) Deserialize(data []byte) error {
	if len(data) == 0 {
		return errors.New("input data's length is zero")
	}

	if err := reqsyn.Reqsyn.Unmarshal(data); err != nil {
		return err
	}
	return nil
}

type TimeoutMsg struct {
	Toutmsg *pb.ToutMsg
}

func (toutmsg *TimeoutMsg) Serialize() ([]byte, error) {
	b, err := toutmsg.Toutmsg.Marshal()
	if err != nil {
		return nil, err
	}
	return b, nil
}

func (toutmsg *TimeoutMsg) Deserialize(data []byte) error {
	if len(data) == 0 {
		return errors.New("input data's length is zero")
	}

	if err := toutmsg.Toutmsg.Unmarshal(data); err != nil {
		return err
	}
	return nil
}

type Signature_BlkF struct {
	Signature_blkf pb.Signature
}

func (sign *Signature_BlkF) Serialize() ([]byte, error) {
	b, err := sign.Signature_blkf.Marshal()
	if err != nil {
		return nil, err
	}
	return b, nil
}

func (sign *Signature_BlkF) Deserialize(data []byte) error {
	if len(data) == 0 {
		return errors.New("input data's length is zero")
	}

	if err := sign.Signature_blkf.Unmarshal(data); err != nil {
		return err
	}
	return nil
}

type Block_Syn struct {
	Blksyn *pb.BlockSyn
}

func (bls *Block_Syn) Serialize() ([]byte, error) {
	b, err := bls.Blksyn.Marshal()
	if err != nil {
		return nil, err
	}
	return b, nil
}

func (bls *Block_Syn) Deserialize(data []byte) error {
	if len(data) == 0 {
		return errors.New("input data's length is zero")
	}

	if err := bls.Blksyn.Unmarshal(data); err != nil {
		return err
	}
	return nil
}