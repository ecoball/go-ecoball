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
	"github.com/gogo/protobuf/proto"
	"github.com/ecoball/go-ecoball/common"
	"github.com/ecoball/go-ecoball/core/pb"
	"math/big"
)

type StateObject struct {
	Address  common.Address //User Address
	addrHash common.Hash    //Address Hash, used to key
	Account  map[common.Address]Account
}

type Account struct {
	Address common.Address //Token Address
	Name    []byte         //Token Name
	Nonce   uint64         //Account Random Number
	Balance *big.Int       //Balance
}

func NewStateObject(address common.Address) (*StateObject, error) {
	state := StateObject{Address: address, Account: make(map[common.Address]Account, 1)}
	state.addrHash = common.SingleHash(address.Bytes())

	return &state, nil
}

func (s *StateObject) AddAccount(token common.Address, name []byte) error {
	ac := Account{Address: token, Name: name, Nonce: 1, Balance: new(big.Int).SetUint64(0)}
	s.Account[token] = ac
	return nil
}

func (s *StateObject) AddBalance(token common.Address, name []byte, amount *big.Int) error {
	if amount.Sign() == 0 {
		return errors.New("amount is zero")
	}
	ac, ok := s.Account[token]
	if !ok {
		if err := s.AddAccount(token, name); err != nil {
			return err
		}
		ac, _ = s.Account[token]
	}
	ac.SetBalance(new(big.Int).Add(ac.GetBalance(), amount))
	ac.Nonce++
	s.Account[token] = ac
	return nil
}

func (s *StateObject) SubBalance(token common.Address, name []byte, amount *big.Int) error {
	if amount.Sign() == 0 {
		return errors.New("amount is zero")
	}
	ac, ok := s.Account[token]
	if !ok {
		return errors.New("not sufficient funds")
	}
	ac.SetBalance(new(big.Int).Sub(ac.GetBalance(), amount))
	ac.Nonce++
	s.Account[token] = ac
	return nil
}

func (s *StateObject) Balance(token common.Address, name []byte) (*big.Int, error) {
	ac, ok := s.Account[token]
	if !ok {
		return nil, errors.New("can't find token account")
	}
	return ac.GetBalance(), nil
}

func (a *Account) SetBalance(amount *big.Int) {
	//TODO:将变动记录存到日志文件
	a.setBalance(amount)
}

func (a *Account) setBalance(amount *big.Int) {
	a.Balance = amount
}

func (a *Account) GetBalance() *big.Int {
	return a.Balance
}

func (s *StateObject) Serialize() ([]byte, error) {
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

func (s *StateObject) ProtoBuf() (*pb.StateObject, error) {
	pbState := pb.StateObject{}
	pbState.Address = s.Address.Bytes()
	pbState.AddrHash = s.addrHash.Bytes()
	var acs []*pb.Account
	for _, v := range s.Account {
		balance, err := v.Balance.GobEncode()
		if err != nil {
			return nil, err
		}
		ac := pb.Account{
			Address: v.Address.Bytes(),
			Name:    common.CopyBytes(v.Name),
			Nonce:   v.Nonce,
			Balance: balance,
		}
		acs = append(acs, &ac)
	}
	pbState.Account = acs

	return &pbState, nil
}

func (s *StateObject) Deserialize(data []byte) error {
	if len(data) == 0 {
		return errors.New("input Account's length is zero")
	}
	var pbObject pb.StateObject
	if err := proto.Unmarshal(data, &pbObject); err != nil {
		return err
	}
	s.Address = common.NewAddress(pbObject.Address)
	s.addrHash = common.NewHash(pbObject.AddrHash)
	s.Account = make(map[common.Address]Account)
	for _, v := range pbObject.Account {
		ac := Account{
			Address: common.NewAddress(v.Address),
			Name:    common.CopyBytes(v.Name),
			Nonce:   v.Nonce,
			Balance: new(big.Int),
		}
		if err := ac.Balance.GobDecode(v.Balance); err != nil {
			return err
		}
		s.Account[ac.Address] = ac
	}

	return nil
}

func (s *StateObject) Show() {
	fmt.Println("\t-----------StateObject------------")
	fmt.Println("\tAddress        :", s.Address.HexString())
	fmt.Println("\tAccount Len    :", len(s.Account))
	for _, v := range s.Account {
		fmt.Println("\tToken          :", v.Address.HexString())
		fmt.Println("\tName           :", string(v.Name))
		fmt.Println("\tBalance        :", v.Balance)
	}
}
