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

package types

import (
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/ecoball/go-ecoball/common"
	"github.com/ecoball/go-ecoball/core/pb"
)

type ConType uint32

const (
	ConDBFT ConType = 0x01
	CondPos ConType = 0x02
	ConSolo ConType = 0x03
	ConABFT ConType = 0x04
)

type ConsensusPayload interface {
	Serialize() ([]byte, error)
	Deserialize(data []byte) error
	GetObject() interface{}
	Show()
}

type ConsensusData struct {
	Type    ConType
	Payload ConsensusPayload
}

func NewConsensusPayload(Type ConType, payload ConsensusPayload) *ConsensusData {
	return &ConsensusData{Type: Type, Payload: payload}
}

func (c *ConsensusData) ProtoBuf() (*pb.ConsensusData, error) {
	data, err := c.Payload.Serialize()
	if err != nil {
		return nil, err
	}
	return &pb.ConsensusData{
		Type: uint32(c.Type),
		Data: common.CopyBytes(data),
	}, nil
}

func (c *ConsensusData) Serialize() ([]byte, error) {
	data, err := c.Payload.Serialize()
	if err != nil {
		return nil, err
	}
	pbCon := pb.ConsensusData{
		Type: uint32(c.Type),
		Data: common.CopyBytes(data),
	}
	b, err := pbCon.Marshal()
	if err != nil {
		return nil, err
	}
	return b, nil
}

func (c *ConsensusData) Deserialize(data []byte) error {
	var pbCon pb.ConsensusData
	if err := pbCon.Unmarshal(data); err != nil {
		return err
	}
	c.Type = ConType(pbCon.Type)
	switch c.Type {
	case CondPos:
		c.Payload = new(DPosData)
	case ConDBFT:
		c.Payload = new(DBFTData)
	case ConSolo:
		c.Payload = new(SoloData)
	case ConABFT:
		c.Payload = new(AbaBftData)
	default:
		return errors.New("unknown consensus type")
	}
	return c.Payload.Deserialize(pbCon.Data)
}

///////////////////////////////////////dPos/////////////////////////////////////////
type DPosData struct {
	proposer common.Hash
}

func (d *DPosData) Serialize() ([]byte, error) {
	return d.proposer.Bytes(), nil
}
func (d *DPosData) Deserialize(data []byte) error {
	d.proposer = common.NewHash(data)
	return nil
}
func (d DPosData) GetObject() interface{} {
	return d
}
func (d *DPosData) Show() {
	fmt.Println("Proposer:", d.proposer)
}

/////////////////////////////////////////dBft///////////////////////////////////////
type DBFTData struct {
	data uint64
}

func (d *DBFTData) Serialize() ([]byte, error) {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, d.data)
	return b, nil
}
func (d *DBFTData) Deserialize(data []byte) error {
	d.data = binary.BigEndian.Uint64(data)
	return nil
}
func (d DBFTData) GetObject() interface{} {
	return d
}
func (d *DBFTData) Show() {
	fmt.Println("Data:", d.data)
}

///////////////////////////////////////////Solo/////////////////////////////////////
type SoloData struct{}

func (s *SoloData) Serialize() ([]byte, error) {
	return nil, nil
}
func (s *SoloData) Deserialize(data []byte) error {
	return nil
}
func (s SoloData) GetObject() interface{} {
	return s
}
func (s *SoloData) Show() {
	fmt.Println("Solo Module Data")
}

///////////////////////////////////////////aBft/////////////////////////////////////
type AbaBftData struct {
	NumberRound        uint32
	PerBlockSignatures []common.Signature
}

func (a *AbaBftData) Serialize() ([]byte, error) {
	var sig []*pb.Signature
	for i := 0; i < len(a.PerBlockSignatures); i++ {
		s := &pb.Signature{PubKey: a.PerBlockSignatures[i].PubKey, SigData: a.PerBlockSignatures[i].SigData}
		sig = append(sig, s)
	}
	pbData := pb.AbaBftData{
		NumberRound: a.NumberRound,
		Sign:        sig,
	}
	data, err := pbData.Marshal()
	if err != nil {
		return nil, err
	}
	return data, nil
}
func (a *AbaBftData) Deserialize(data []byte) error {
	if len(data) == 0 {
		return errors.New("AbaBftData is nil")
	}
	var pbData pb.AbaBftData
	if err := pbData.Unmarshal(data); err != nil {
		return err
	}
	a.NumberRound = pbData.NumberRound
	for i := 0; i < len(pbData.Sign); i++ {
		sig := common.Signature{
			PubKey:  common.CopyBytes(pbData.Sign[i].PubKey),
			SigData: common.CopyBytes(pbData.Sign[i].SigData),
		}
		a.PerBlockSignatures = append(a.PerBlockSignatures, sig)
	}
	return nil
}
func (a AbaBftData) GetObject() interface{} {
	return a
}
func (a *AbaBftData) Show() {
	fmt.Println("\t-----------AbaBft------------")
	fmt.Println("\tNumberRound    :", a.NumberRound)
	fmt.Println("\tSig Len        :", len(a.PerBlockSignatures))
	for i := 0; i < len(a.PerBlockSignatures); i++ {
		fmt.Println("\tPublicKey      :", common.ToHex(a.PerBlockSignatures[i].PubKey))
		fmt.Println("\tSigData        :", common.ToHex(a.PerBlockSignatures[i].SigData))
	}
}
