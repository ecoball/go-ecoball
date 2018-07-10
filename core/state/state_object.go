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
	"encoding/json"
	"errors"
	"github.com/ecoball/go-ecoball/common"
	"github.com/ecoball/go-ecoball/core/pb"
	"github.com/gogo/protobuf/proto"
	"math/big"
)

type account struct {
	Actor      common.AccountName `json:"actor"`
	Weight     uint32             `json:"weight"`
	Permission string             `json:"permission"`
}

type address struct {
	Actor  common.Address `json:"actor"`
	Weight uint32         `json:"weight"`
}

type Permission struct {
	PermName  string    `json:"perm_name"`
	Parent    string    `json:"parent"`
	Threshold uint32    `json:"threshold"`
	Keys      []address `json:"keys"`
	Accounts  []account `json:"accounts"`
}

type Token struct {
	Index   common.AccountName `json:"index"`
	Balance *big.Int           `json:"balance"`
}

type Account struct {
	Index       common.AccountName           `json:"index"`
	Nonce       uint64                       `json:"nonce"`
	Tokens      map[common.AccountName]Token `json:"token"`
	Permissions map[string]Permission        `json:"permissions"`
}

/**
 *  @brief create a new account, binding a char name with a address
 *  @param index - the unique id of account name created by common.NameToIndex()
 *  @param address - the account's public key
 */
func NewAccount(index common.AccountName, addr common.Address) (*Account, error) {
	acc := Account{
		Index:       index,
		Nonce:       0,
		Tokens:      make(map[common.AccountName]Token, 1),
		Permissions: make(map[string]Permission, 1),
	}
	acc.Permissions["owner"] = Permission{
		PermName:  "owner",
		Parent:    "",
		Threshold: 1,
		Keys:      []address{{Actor: addr, Weight: 1}},
	}
	acc.Permissions["active"] = Permission{
		PermName:  "active",
		Parent:    "owner",
		Threshold: 1,
		Accounts:  []account{{Actor: index, Weight: 1, Permission: "owner"}},
	}
	return &acc, nil
}

/**
 *  @brief create a new token in account
 *  @param index - the unique id of token name created by common.NameToIndex()
 */
func (s *Account) AddToken(index common.AccountName) error {
	ac := Token{Index: index, Balance: new(big.Int).SetUint64(0)}
	s.Tokens[index] = ac
	return nil
}

/**
 *  @brief check the token for existence, return true if existed
 *  @param index - the unique id of token name created by common.NameToIndex()
 */
func (s *Account) TokenExisted(index common.AccountName) bool {
	_, ok := s.Tokens[index]
	if ok {
		return true
	}
	return false
}

/**
 *  @brief add balance into account
 *  @param index - the unique id of token name created by common.NameToIndex()
 *  @param amount - value of token
 */
func (s *Account) AddBalance(index common.AccountName, amount *big.Int) error {
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

/**
 *  @brief sub balance into account
 *  @param index - the unique id of token name created by common.NameToIndex()
 *  @param amount - value of token
 */
func (s *Account) SubBalance(index common.AccountName, amount *big.Int) error {
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

/**
 *  @brief get the balance of account
 *  @param index - the unique id of token name created by common.NameToIndex()
 *  @return big.int - value of token
 */
func (s *Account) Balance(index common.AccountName) (*big.Int, error) {
	ac, ok := s.Tokens[index]
	if !ok {
		return nil, errors.New("can't find token account")
	}
	return ac.GetBalance(), nil
}

/**
 *  @brief set balance of account
 *  @param amount - value of token
 */
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

/**
 *  @brief converts a structure into a sequence of characters
 *  @return []byte - a sequence of characters
 */
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

func (s *Account) ProtoBuf() (*pb.Account, error) {
	var tokens []*pb.Token
	for _, v := range s.Tokens {
		balance, err := v.Balance.GobEncode()
		if err != nil {
			return nil, err
		}
		ac := pb.Token{
			Index:   uint64(v.Index),
			Balance: balance,
		}
		tokens = append(tokens, &ac)
	}
	var perms []*pb.Permission
	for _, perm := range s.Permissions {
		var pbKeys []*pb.KeyWeight
		var pbAccounts []*pb.AccountWeight
		for _, key := range perm.Keys {
			pbKey := &pb.KeyWeight{Actor: key.Actor.Bytes(), Weight: key.Weight}
			pbKeys = append(pbKeys, pbKey)
		}
		for _, acc := range perm.Accounts {
			pbAccount := &pb.AccountWeight{Actor: uint64(acc.Actor), Weight: acc.Weight, Permission: []byte(acc.Permission)}
			pbAccounts = append(pbAccounts, pbAccount)
		}
		pbPerm := &pb.Permission{
			PermName:  []byte(perm.PermName),
			Parent:    []byte(perm.Parent),
			Threshold: perm.Threshold,
			Keys:      pbKeys,
			Accounts:  pbAccounts,
		}
		perms = append(perms, pbPerm)
	}
	pbAcc := pb.Account{
		Index:       uint64(s.Index),
		Nonce:       s.Nonce,
		Tokens:      tokens,
		Permissions: perms,
	}

	return &pbAcc, nil
}

/**
 *  @brief converts a sequence of characters into a structure
 *  @param data - a sequence of characters
 */
func (s *Account) Deserialize(data []byte) error {
	if len(data) == 0 {
		return errors.New("input Token's length is zero")
	}
	var pbAcc pb.Account
	if err := proto.Unmarshal(data, &pbAcc); err != nil {
		return err
	}
	s.Index = common.AccountName(pbAcc.Index)
	s.Nonce = pbAcc.Nonce
	s.Tokens = make(map[common.AccountName]Token)
	s.Permissions = make(map[string]Permission, 1)
	for _, v := range pbAcc.Tokens {
		ac := Token{
			Index:   common.AccountName(v.Index),
			Balance: new(big.Int),
		}
		if err := ac.Balance.GobDecode(v.Balance); err != nil {
			return err
		}
		s.Tokens[ac.Index] = ac
	}
	for _, pbPerm := range pbAcc.Permissions {
		var keys []address
		for _, pbKey := range pbPerm.Keys {
			key := address{Actor: common.NewAddress(pbKey.Actor), Weight: pbKey.Weight}
			keys = append(keys, key)
		}
		var accounts []account
		for _, pbAcc := range pbPerm.Accounts {
			acc := account{Actor: common.AccountName(pbAcc.Actor), Weight: pbAcc.Weight, Permission: string(pbAcc.Permission)}
			accounts = append(accounts, acc)
		}
		s.Permissions[string(pbPerm.PermName)] = Permission{
			PermName:  string(pbPerm.PermName),
			Parent:    string(pbPerm.Parent),
			Threshold: pbPerm.Threshold,
			Keys:      keys,
			Accounts:  accounts,
		}
	}

	return nil
}

func (s *Account) JsonString() string {
	data, _ := json.Marshal(s)
	return string(data)
}
