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
	"github.com/ecoball/go-ecoball/core/bloom"
	"github.com/ecoball/go-ecoball/core/pb"
	"github.com/ecoball/go-ecoball/core/trie"
	"time"
)

type Block struct {
	*Header
	CountTxs     uint32
	Transactions []*Transaction
}

func NewBlock(prevHeader *Header, stateHash common.Hash, consensusData ConsensusData, txs []*Transaction) (*Block, error) {
	if nil == prevHeader {
		return nil, errors.New("invalid parameter preHeader")
	}
	timeStamp := time.Now().Unix()
	var Bloom bloom.Bloom
	var hashes []common.Hash
	for _, t := range txs {
		hashes = append(hashes, t.Hash)
		Bloom.Add(t.Hash.Bytes())
		Bloom.Add(common.IndexToBytes(t.From))
		Bloom.Add(common.IndexToBytes(t.Addr))
	}
	merkleHash, err := trie.GetMerkleRoot(hashes)
	if err != nil {
		return nil, err
	}

	header, err := NewHeader(VersionHeader, prevHeader.Height+1, prevHeader.Hash, merkleHash, stateHash, consensusData, Bloom, timeStamp)
	if err != nil {
		return nil, err
	}
	block := Block{header, uint32(len(txs)), txs}
	return &block, nil
}

func (b *Block) SetSignature(account *account.Account) error {
	return b.Header.SetSignature(account)
}

func GenesesBlockInitConsensusData(timestamp int64) *ConsensusData {
	conData, err := InitConsensusData(timestamp)
	if err != nil {
		log.Debug(err)
		return nil
	}
	return conData
}

func GenesesBlockInit() (*Block, error) {
	tm, err := time.Parse("02/01/2006 15:04:05 PM", "21/02/1990 00:00:00 AM")
	if err != nil {
		return nil, err
	}
	timeStamp := tm.Unix()

	//TODO start
	SecondInMs := int64(1000)
	BlockIntervalInMs := int64(15000)
	timeStamp = int64((timeStamp*SecondInMs-SecondInMs)/BlockIntervalInMs) * BlockIntervalInMs
	timeStamp = timeStamp / SecondInMs
	//TODO end

	hash := common.NewHash([]byte("EcoBall Geneses Block"))
	conData := GenesesBlockInitConsensusData(timeStamp)
	header, err := NewHeader(VersionHeader, 1, hash, hash, hash, *conData, bloom.Bloom{}, timeStamp)
	if err != nil {
		return nil, err
	}
	block := Block{header, 0, nil}
	return &block, nil
}

func (b *Block) protoBuf() (*pb.BlockTx, error) {
	var block pb.BlockTx
	var err error
	block.Header, err = b.Header.protoBuf()
	if err != nil {
		return nil, err
	}
	var pbTxs []*pb.Transaction
	for _, tx := range b.Transactions {
		pbTx, err := tx.protoBuf()
		if err != nil {
			return nil, err
		}
		pbTxs = append(pbTxs, pbTx)
	}
	block.Transactions = append(block.Transactions, pbTxs...)
	return &block, nil
}

func (b *Block) Serialize() (data []byte, err error) {
	p, err := b.protoBuf()
	if err != nil {
		return nil, err
	}
	data, err = p.Marshal()
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (b *Block) Deserialize(data []byte) error {
	if len(data) == 0 {
		return errors.New("input data's length is zero")
	}
	var pbBlock pb.BlockTx
	if err := pbBlock.Unmarshal(data); err != nil {
		return err
	}
	dataHeader, err := pbBlock.Header.Marshal()
	if err != nil {
		return err
	}

	b.Header = new(Header)
	err = b.Header.Deserialize(dataHeader)
	if err != nil {
		return err
	}

	var txs []*Transaction
	for _, tx := range pbBlock.Transactions {
		b, err := tx.Marshal()
		if err != nil {
			return err
		}
		t := new(Transaction)
		if err := t.Deserialize(b); err != nil {
			return err
		}
		txs = append(txs, t)
	}

	b.CountTxs = uint32(len(txs))
	b.Transactions = txs

	return nil
}

func (b *Block) Show() {
	fmt.Println("\t-----------Block-------------")
	b.Header.Show()
	fmt.Println("\tTxs Number     :", b.CountTxs)
}
