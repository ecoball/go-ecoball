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

package state

import (
	"errors"
	"fmt"
	"github.com/ecoball/go-ecoball/common"
	"github.com/ecoball/go-ecoball/core/pb"
	"github.com/gogo/protobuf/proto"
	"math/big"
)

type Permission struct {
	parent    uint64
	threshold int
	addr      uint64
	weight    int
}

type Account struct {
	Index   uint64
	Nonce   uint64         //Token Random Number
	Address common.Address //User Address
	Tokens  map[uint64]Token
}

type Token struct {
	Index   uint64   //Token Name UUID
	Balance *big.Int //Value
}

/**
 *  @brief create a new account, binding a char name with a address
 *  @param index - the unique id of account name created by common.NameToIndex()
 *  @param address - the account's public key
 */
func NewAccount(index uint64, address common.Address) (*Account, error) {
	state := Account{Index: index, Nonce: 0, Address: address, Tokens: make(map[uint64]Token, 1)}
	return &state, nil
}

/**
 *  @brief create a new token in account
 *  @param index - the unique id of token name created by common.NameToIndex()
 */
func (s *Account) AddToken(index uint64) error {
	ac := Token{Index: index, Balance: new(big.Int).SetUint64(0)}
	s.Tokens[index] = ac
	return nil
}

/**
 *  @brief check the token for existence, return true if existed
 *  @param index - the unique id of token name created by common.NameToIndex()
 */
func (s *Account) TokenExisted(index uint64) bool {
	_, ok := s.Tokens[index]
	if ok {
		return true
	}
	return false
}

func (s *Account) AddBalance(index uint64, amount *big.Int) error {
	if amount.Sign() == 0 {
		return errors.New("amount is zero")
	}
	ac, ok := s.Tokens[index]
	if !ok {
		if err := s.AddToken(index); err != nil {
			return err
		}
		ac, _ = s.Tokens[index]
	}
	ac.SetBalance(new(big.Int).Add(ac.GetBalance(), amount))
	s.Nonce++
	s.Tokens[index] = ac
	return nil
}

func (s *Account) SubBalance(index uint64, amount *big.Int) error {
	if amount.Sign() == 0 {
		return errors.New("amount is zero")
	}
	ac, ok := s.Tokens[index]
	if !ok {
		return errors.New("not sufficient funds")
	}
	ac.SetBalance(new(big.Int).Sub(ac.GetBalance(), amount))
	s.Nonce++
	s.Tokens[index] = ac
	return nil
}

func (s *Account) Balance(index uint64) (*big.Int, error) {
	ac, ok := s.Tokens[index]
	if !ok {
		return nil, errors.New("can't find token account")
	}
	return ac.GetBalance(), nil
}

func (a *Token) SetBalance(amount *big.Int) {
	//TODO:将变动记录存到日志文件
	a.setBalance(amount)
}

func (a *Token) setBalance(amount *big.Int) {
	a.Balance = amount
}

func (a *Token) GetBalance() *big.Int {
	return a.Balance
}

func (s *Account) Serialize() ([]byte, error) {
	p, err := s.ProtoBuf()
	if err != nil {
		return nil, err
	}
	data, err := proto.Marshal(p)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (s *Account) ProtoBuf() (*pb.StateObject, error) {
	var tokens []*pb.Token
	for _, v := range s.Tokens {
		balance, err := v.Balance.GobEncode()
		if err != nil {
			return nil, err
		}
		ac := pb.Token{
			Index:   v.Index,
			Balance: balance,
		}
		tokens = append(tokens, &ac)
	}
	pbState := pb.StateObject{
		Index:   s.Index,
		Nonce:   s.Nonce,
		Address: s.Address.Bytes(),
		Tokens:  tokens,
	}

	return &pbState, nil
}

func (s *Account) Deserialize(data []byte) error {
	if len(data) == 0 {
		return errors.New("input Token's length is zero")
	}
	var pbObject pb.StateObject
	if err := proto.Unmarshal(data, &pbObject); err != nil {
		return err
	}
	s.Index = pbObject.Index
	s.Nonce = pbObject.Nonce
	s.Address = common.NewAddress(pbObject.Address)
	s.Tokens = make(map[uint64]Token)
	for _, v := range pbObject.Tokens {
		ac := Token{
			Index:   v.Index,
			Balance: new(big.Int),
		}
		if err := ac.Balance.GobDecode(v.Balance); err != nil {
			return err
		}
		s.Tokens[ac.Index] = ac
	}

	return nil
}

func (s *Account) Show() {
	fmt.Println("\t-----------Tokens------------")
	fmt.Println("\tIndex          :", s.Index)
	fmt.Println("\tName           :", common.IndexToName(s.Index))
	fmt.Println("\tNonce          :", s.Nonce)
	fmt.Println("\tAddress        :", s.Address.HexString())
	fmt.Println("\tTokens Len     :", len(s.Tokens))
	for _, v := range s.Tokens {
		fmt.Println("\tName           :", common.IndexToName(v.Index))
		fmt.Println("\tBalance        :", v.Balance)
	}
}
