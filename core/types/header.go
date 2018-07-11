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
	"errors"
	"fmt"
	"github.com/ecoball/go-ecoball/account"
	"github.com/ecoball/go-ecoball/common"
	"github.com/ecoball/go-ecoball/common/elog"
	"github.com/ecoball/go-ecoball/core/bloom"
	"github.com/ecoball/go-ecoball/core/pb"
	"github.com/ecoball/go-ecoball/crypto/secp256k1"
	"encoding/json"
)

const VersionHeader = 1

type Header struct {
	Version       uint32
	TimeStamp     int64
	Height        uint64
	ConsensusData ConsensusData
	PrevHash      common.Hash
	MerkleHash    common.Hash
	StateHash     common.Hash
	Bloom         bloom.Bloom
	Signatures    []common.Signature

	Hash common.Hash
}

var log = elog.NewLogger("LedgerImpl", elog.DebugLog)

/**
* New a Header and compute it's hash
 */
func NewHeader(version uint32, height uint64, prevHash, merkleHash, stateHash common.Hash, conData ConsensusData, bloom bloom.Bloom, timeStamp int64) (*Header, error) {
	if version != VersionHeader {
		return nil, errors.New("version mismatch")
	}
	if conData.Payload == nil {
		return nil, errors.New("consensus' payload is nil")
	}
	header := Header{
		Version:       version,
		TimeStamp:     timeStamp,
		Height:        height,
		ConsensusData: conData,
		PrevHash:      prevHash,
		MerkleHash:    merkleHash,
		StateHash:     stateHash,
		Bloom:         bloom,
	}
	payload, err := header.unSignatureData()
	if err != nil {
		return nil, err
	}
	b, err := payload.Marshal()
	if err != nil {
		return nil, err
	}
	header.Hash, err = common.DoubleHash(b)
	fmt.Println("New Header Hash:", header.Hash.HexString())
	if err != nil {
		return nil, err
	}
	return &header, nil
}

func (h *Header) InitializeHash() error {
	if h.Version != VersionHeader {
		return errors.New("version mismatch")
	}
	if h.ConsensusData.Payload == nil {
		return errors.New("consensus' payload is nil")
	}
	payload, err := h.unSignatureData()
	if err != nil {
		return err
	}
	b, err := payload.Marshal()
	if err != nil {
		return err
	}
	h.Hash, err = common.DoubleHash(b)
	fmt.Println("New Header Hash:", h.Hash.HexString())
	if err != nil {
		return err
	}
	return nil
}

func (h *Header) SetSignature(account *account.Account) error {
	sigData, err := account.Sign(h.Hash.Bytes())
	if err != nil {
		return err
	}
	sig := common.Signature{}
	sig.SigData = common.CopyBytes(sigData)
	sig.PubKey = common.CopyBytes(account.PublicKey)
	h.Signatures = append(h.Signatures, sig)
	return nil
}

func (h *Header) VerifySignature() (bool, error) {
	h.Show()
	for _, v := range h.Signatures {
		b, err := secp256k1.Verify(h.Hash.Bytes(), v.SigData, v.PubKey)
		if err != nil || b != true {
			return false, err
		}
	}
	return true, nil
}

/**
** Used to compute hash
 */
func (h *Header) unSignatureData() (*pb.Header, error) {
	if h.TimeStamp == 0 {
		return nil, errors.New("this header struct is illegal")
	}
	pbCon, err := h.ConsensusData.ProtoBuf()
	if err != nil {
		return nil, err
	}
	return &pb.Header{
		Version:       h.Version,
		Timestamp:     h.TimeStamp,
		Height:        h.Height,
		ConsensusData: pbCon,
		PrevHash:      h.PrevHash.Bytes(),
		MerkleHash:    h.MerkleHash.Bytes(),
		StateHash:     h.StateHash.Bytes(),
		Bloom:         h.Bloom.Bytes(),
	}, nil
}

func (h *Header) protoBuf() (*pb.HeaderTx, error) {
	var sig []*pb.Signature
	for i := 0; i < len(h.Signatures); i++ {
		s := &pb.Signature{PubKey: h.Signatures[i].PubKey, SigData: h.Signatures[i].SigData}
		sig = append(sig, s)
	}
	pbCon, err := h.ConsensusData.ProtoBuf()
	if err != nil {
		return nil, err
	}
	return &pb.HeaderTx{
		Version:       h.Version,
		Timestamp:     h.TimeStamp,
		Height:        h.Height,
		ConsensusData: pbCon,
		PrevHash:      h.PrevHash.Bytes(),
		MerkleHash:    h.MerkleHash.Bytes(),
		Sign:          sig,
		StateHash:     h.StateHash.Bytes(),
		Bloom:         h.Bloom.Bytes(),
		BlockHash:     h.Hash.Bytes(),
	}, nil
}

/**
 *  @brief converts a structure into a sequence of characters
 *  @return []byte - a sequence of characters
 */
func (h *Header) Serialize() ([]byte, error) {
	p, err := h.protoBuf()
	if err != nil {
		return nil, err
	}
	data, err := p.Marshal()
	if err != nil {
		return nil, err
	}
	return data, nil
}

/**
 *  @brief converts a sequence of characters into a structure
 *  @param data - a sequence of characters
 */
func (h *Header) Deserialize(data []byte) error {
	if len(data) == 0 {
		return errors.New("input data's length is zero")
	}
	var pbHeader pb.HeaderTx
	if err := pbHeader.Unmarshal(data); err != nil {
		return err
	}

	h.Version = pbHeader.Version
	h.TimeStamp = pbHeader.Timestamp
	h.Height = pbHeader.Height
	h.PrevHash = common.NewHash(pbHeader.PrevHash)
	h.MerkleHash = common.NewHash(pbHeader.MerkleHash)
	for i := 0; i < len(pbHeader.Sign); i++ {
		sig := common.Signature{
			PubKey:  common.CopyBytes(pbHeader.Sign[i].PubKey),
			SigData: common.CopyBytes(pbHeader.Sign[i].SigData),
		}
		h.Signatures = append(h.Signatures, sig)
	}
	h.StateHash = common.NewHash(pbHeader.StateHash)
	h.Hash = common.NewHash(pbHeader.BlockHash)
	h.Bloom = bloom.NewBloom(pbHeader.Bloom)

	dataCon, err := pbHeader.ConsensusData.Marshal()
	if err != nil {
		return err
	}
	if err := h.ConsensusData.Deserialize(dataCon); err != nil {
		return err
	}

	return nil
}

func (h *Header) show() {
	fmt.Println("\t-----------Header------------")
	fmt.Println("\tHeight         :", h.Height)
	fmt.Println("\tTime           :", h.TimeStamp)
	fmt.Println("\tVersion        :", h.Version)
	fmt.Println("\tPrevHash       :", h.PrevHash.HexString())
	fmt.Println("\tMerkleHash     :", h.MerkleHash.HexString())
	fmt.Println("\tStateHash      :", h.StateHash.HexString())
	fmt.Println("\tHash           :", h.Hash.HexString())
	fmt.Println("\tSig Len        :", len(h.Signatures))
	for i := 0; i < len(h.Signatures); i++ {
		fmt.Println("\tPublicKey      :", common.ToHex(h.Signatures[i].PubKey))
		fmt.Println("\tSigData        :", common.ToHex(h.Signatures[i].SigData))
	}
}

func (h *Header) JsonString() string {
	data, err := json.Marshal(h)
	if err != nil {
		fmt.Println(err)
	}
	return string(data)
}

func (h *Header) Show() {
	log.Debug(h.JsonString())
}