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
	"fmt"
	"github.com/ecoball/go-ecoball/common"
	"github.com/ecoball/go-ecoball/core/pb"
	"github.com/gogo/protobuf/proto"
	"math/big"
)

type Token struct {
	Index   common.AccountName `json:"index"`
	Balance *big.Int           `json:"balance"`
}

type Account struct {
	Index       common.AccountName    `json:"index"`
	Nonce       uint64                `json:"nonce"`
	Tokens      map[string]Token      `json:"token"`
	Permissions map[string]Permission `json:"permissions"`
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
		Tokens:      make(map[string]Token, 1),
		Permissions: make(map[string]Permission, 1),
	}
	Keys := make(map[string]address, 1)
	Keys[addr.HexString()] = address{Actor: addr, Weight: 1}
	Accounts := make(map[string]account, 1)
	Accounts[index.String()] = account{Actor: index, Weight: 1, Permission: "owner"}
	acc.Permissions["owner"] = Permission{
		PermName:  "owner",
		Parent:    "",
		Threshold: 1,
		Keys:      Keys,
		Accounts:  make(map[string]account, 0),
	}
	acc.Permissions["active"] = Permission{
		PermName:  "active",
		Parent:    "owner",
		Threshold: 1,
		Keys:      make(map[string]address),
		Accounts:  Accounts,
	}
	return &acc, nil
}

/**
 *  @brief create a new token in account
 *  @param index - the unique id of token name created by common.NameToIndex()
 */
func (a *Account) AddToken(index common.AccountName) error {
	ac := Token{Index: index, Balance: new(big.Int).SetUint64(0)}
	a.Tokens[index.String()] = ac
	return nil
}

/**
 *  @brief check the token for existence, return true if existed
 *  @param index - the unique id of token name created by common.NameToIndex()
 */
func (a *Account) TokenExisted(index common.AccountName) bool {
	_, ok := a.Tokens[common.IndexToName(index)]
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
func (a *Account) AddBalance(index common.AccountName, amount *big.Int) error {
	if amount.Sign() == 0 {
		return errors.New("amount is zero")
	}
	ac, ok := a.Tokens[common.IndexToName(index)]
	if !ok {
		if err := a.AddToken(index); err != nil {
			return err
		}
		ac, _ = a.Tokens[common.IndexToName(index)]
	}
	ac.SetBalance(new(big.Int).Add(ac.GetBalance(), amount))
	a.Nonce++
	a.Tokens[common.IndexToName(index)] = ac
	return nil
}

/**
 *  @brief sub balance into account
 *  @param index - the unique id of token name created by common.NameToIndex()
 *  @param amount - value of token
 */
func (a *Account) SubBalance(index common.AccountName, amount *big.Int) error {
	if amount.Sign() == 0 {
		return errors.New("amount is zero")
	}
	ac, ok := a.Tokens[common.IndexToName(index)]
	if !ok {
		return errors.New("not sufficient funds")
	}
	ac.SetBalance(new(big.Int).Sub(ac.GetBalance(), amount))
	a.Nonce++
	a.Tokens[common.IndexToName(index)] = ac
	return nil
}

/**
 *  @brief get the balance of account
 *  @param index - the unique id of token name created by common.NameToIndex()
 *  @return big.int - value of token
 */
func (a *Account) Balance(index common.AccountName) (*big.Int, error) {
	ac, ok := a.Tokens[common.IndexToName(index)]
	if !ok {
		return nil, errors.New("can't find token account")
	}
	return ac.GetBalance(), nil
}

/**
 *  @brief set balance of account
 *  @param amount - value of token
 */
func (t *Token) SetBalance(amount *big.Int) {
	//TODO:将变动记录存到日志文件
	t.setBalance(amount)
}

func (t *Token) setBalance(amount *big.Int) {
	t.Balance = amount
}

func (t *Token) GetBalance() *big.Int {
	return t.Balance
}

/**
 *  @brief converts a structure into a sequence of characters
 *  @return []byte - a sequence of characters
 */
func (a *Account) Serialize() ([]byte, error) {
	p, err := a.ProtoBuf()
	if err != nil {
		return nil, err
	}
	data, err := proto.Marshal(p)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (a *Account) ProtoBuf() (*pb.Account, error) {
	var tokens []*pb.Token
	for _, v := range a.Tokens {
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
	for _, perm := range a.Permissions {
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
		Index:       uint64(a.Index),
		Nonce:       a.Nonce,
		Tokens:      tokens,
		Permissions: perms,
	}

	return &pbAcc, nil
}

/**
 *  @brief converts a sequence of characters into a structure
 *  @param data - a sequence of characters
 */
func (a *Account) Deserialize(data []byte) error {
	if len(data) == 0 {
		return errors.New("input Token's length is zero")
	}
	var pbAcc pb.Account
	if err := proto.Unmarshal(data, &pbAcc); err != nil {
		return err
	}
	a.Index = common.AccountName(pbAcc.Index)
	a.Nonce = pbAcc.Nonce
	a.Tokens = make(map[string]Token)
	a.Permissions = make(map[string]Permission, 1)
	for _, v := range pbAcc.Tokens {
		ac := Token{
			Index:   common.AccountName(v.Index),
			Balance: new(big.Int),
		}
		if err := ac.Balance.GobDecode(v.Balance); err != nil {
			return err
		}
		a.Tokens[common.IndexToName(ac.Index)] = ac
	}
	for _, pbPerm := range pbAcc.Permissions {
		keys := make(map[string]address, 1)
		for _, pbKey := range pbPerm.Keys {
			key := address{Actor: common.NewAddress(pbKey.Actor), Weight: pbKey.Weight}
			keys[common.NewAddress(pbKey.Actor).HexString()] = key
		}
		accounts := make(map[string]account, 1)
		for _, pbAcc := range pbPerm.Accounts {
			acc := account{Actor: common.AccountName(pbAcc.Actor), Weight: pbAcc.Weight, Permission: string(pbAcc.Permission)}
			accounts[common.AccountName(pbAcc.Actor).String()] = acc
		}
		a.Permissions[string(pbPerm.PermName)] = Permission{
			PermName:  string(pbPerm.PermName),
			Parent:    string(pbPerm.Parent),
			Threshold: pbPerm.Threshold,
			Keys:      keys,
			Accounts:  accounts,
		}
	}

	return nil
}

func (a *Account) JsonString() string {
	data, err := json.Marshal(a)
	if err != nil {
		fmt.Println(err)
	}
	return string(data)
}

func (a *Account) Show() {
	fmt.Println(a.JsonString())
}
