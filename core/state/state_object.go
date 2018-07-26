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
	"github.com/ecoball/go-ecoball/core/store"
	"github.com/ecoball/go-ecoball/core/types"
	"github.com/gogo/protobuf/proto"
	"math/big"
	"sort"
	"time"
)

type Account struct {
	Index       common.AccountName    `json:"index"`
	TimeStamp   int64                 `json:"timestamp"`
	Tokens      map[string]Token      `json:"token"`
	Permissions map[string]Permission `json:"permissions"`
	Contract    types.DeployInfo      `json:"contract"`
	Delegates   []Delegate            `json:"delegate"`
	Resource    `json:"resource"`

	Hash   common.Hash `json:"hash"`
	trie   Trie
	db     Database
	diskDb *store.LevelDBStore
}

/**
 *  @brief create a new account, binding a char name with a address
 *  @param index - the unique id of account name created by common.NameToIndex()
 *  @param address - the account's public key
 */
func NewAccount(path string, index common.AccountName, addr common.Address) (acc *Account, err error) {
	log.Info("add a new account:", index)
	fmt.Printf("index:%d\n", index)
	acc = &Account{
		Index:       index,
		TimeStamp:   time.Now().Unix(),
		Tokens:      make(map[string]Token, 1),
		Permissions: make(map[string]Permission, 1),
	}
	perm := NewPermission(Owner, "", 1, []KeyFactor{{Actor: addr, Weight: 1}}, []AccFactor{})
	acc.AddPermission(perm)
	perm = NewPermission(Active, Owner, 1, []KeyFactor{{Actor: addr, Weight: 1}}, []AccFactor{})
	acc.AddPermission(perm)

	if err := acc.NewStoreTrie(path); err != nil {
		return nil, err
	}
	acc.diskDb.Close()
	return acc, nil
}

func (a *Account) NewStoreTrie(path string) error {
	diskDb, err := store.NewLevelDBStore(path+"/"+common.IndexToName(a.Index), 0, 0)
	if err != nil {
		return err
	}
	a.diskDb = diskDb
	a.db = NewDatabase(diskDb)
	a.trie, err = a.db.OpenTrie(a.Hash)
	if err != nil {
		a.trie, err = a.db.OpenTrie(common.Hash{})
		if err != nil {
			return err
		}
	}
	return nil
}

/**
 *  @brief add a smart contract into a account data
 *  @param t - the type of virtual machine
 *  @param des - the description of smart contract
 *  @param code - the code of smart contract
 */
func (a *Account) SetContract(t types.VmType, des, code []byte) error {
	a.Contract.TypeVm = t
	a.Contract.Describe = common.CopyBytes(des)
	a.Contract.Code = common.CopyBytes(code)
	return nil
}

/**
 *  @brief get a smart contract from a account data
 */
func (a *Account) GetContract() (*types.DeployInfo, error) {
	if a.Contract.TypeVm == 0 {
		return nil, errors.New("this account is not set contract")
	}
	return &a.Contract, nil
}

func (a *Account) StoreSet(path string, key, value []byte) (err error) {
	if err := a.NewStoreTrie(path); err != nil {
		return err
	}
	defer a.diskDb.Close()
	log.Debug("StoreSet key:", string(key), "value:", string(value))
	if err := a.trie.TryUpdate(key, value); err != nil {
		return err
	}
	if _, err := a.trie.Commit(nil); err != nil {
		return err
	}
	if err := a.db.TrieDB().Commit(a.trie.Hash(), false); err != nil {
		return err
	}
	a.Hash = a.trie.Hash()
	return nil
}
func (a *Account) StoreGet(path string, key []byte) (value []byte, err error) {
	if err := a.NewStoreTrie(path); err != nil {
		return nil, err
	}
	defer a.diskDb.Close()
	value, err = a.trie.TryGet(key)
	if err != nil {
		return nil, err
	}
	log.Debug("StoreGet key:", string(key), "value:", string(value))
	return value, nil
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
	var keysToken []string
	for _, v := range a.Tokens {
		keysToken = append(keysToken, v.Name)
	}
	sort.Strings(keysToken)
	for _, k := range keysToken {
		v := a.Tokens[k]
		balance, err := v.Balance.GobEncode()
		if err != nil {
			return nil, err
		}
		ac := pb.Token{
			Name:    v.Name,
			Balance: balance,
		}
		tokens = append(tokens, &ac)
	}

	var perms []*pb.Permission
	var keysPerm []string
	for _, perm := range a.Permissions {
		keysPerm = append(keysPerm, perm.PermName)
	}
	sort.Strings(keysPerm)
	for _, k := range keysPerm {
		perm := a.Permissions[k]
		var pbKeys []*pb.KeyWeight
		var pbAccounts []*pb.AccountWeight
		var keysKeys []string
		var keysAccount []string
		for _, key := range perm.Keys {
			keysKeys = append(keysKeys, key.Actor.HexString())
		}
		sort.Strings(keysKeys)
		for _, k := range keysKeys {
			key := perm.Keys[k]
			pbKey := &pb.KeyWeight{Actor: key.Actor.Bytes(), Weight: key.Weight}
			pbKeys = append(pbKeys, pbKey)
		}

		for _, acc := range perm.Accounts {
			keysAccount = append(keysAccount, acc.Permission)
		}
		sort.Strings(keysAccount)
		for _, k := range keysAccount {
			acc := perm.Accounts[k]
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
	var delegates []*pb.Delegate
	for _, v := range a.Delegates {
		d := pb.Delegate{Index: uint64(v.Index), Cpu: v.CpuStaked, Net: v.NetStaked}
		delegates = append(delegates, &d)
	}
	pbAcc := pb.Account{
		Index:       uint64(a.Index),
		TimeStamp:   a.TimeStamp,
		Tokens:      tokens,
		Permissions: perms,
		Contract: &pb.DeployInfo{
			TypeVm:   uint32(a.Contract.TypeVm),
			Describe: common.CopyBytes(a.Contract.Describe),
			Code:     common.CopyBytes(a.Contract.Code),
		},
		Delegates: delegates,
		Ram: &pb.Ram{
			Quota: a.Ram.Quota,
			Used:  a.Ram.Used,
		},
		Cpu: &pb.Res{
			Staked:    a.Cpu.Staked,
			Delegated: a.Cpu.Delegated,
			Used:      a.Cpu.Used,
			Available: a.Cpu.Available,
			Limit:     a.Cpu.Limit,
		},
		Net: &pb.Res{
			Staked:    a.Net.Staked,
			Delegated: a.Net.Delegated,
			Used:      a.Net.Used,
			Available: a.Net.Available,
			Limit:     a.Net.Limit,
		},
		Hash: a.Hash.Bytes(),
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
	a.TimeStamp = pbAcc.TimeStamp

	a.Ram.Quota = pbAcc.Ram.Quota
	a.Ram.Used = pbAcc.Ram.Used
	a.Cpu.Staked = pbAcc.Cpu.Staked
	a.Cpu.Delegated = pbAcc.Cpu.Delegated
	a.Cpu.Used = pbAcc.Cpu.Used
	a.Cpu.Available = pbAcc.Cpu.Available
	a.Cpu.Limit = pbAcc.Cpu.Limit
	a.Net.Staked = pbAcc.Net.Staked
	a.Net.Delegated = pbAcc.Net.Delegated
	a.Net.Used = pbAcc.Net.Used
	a.Net.Available = pbAcc.Net.Available
	a.Net.Limit = pbAcc.Net.Limit

	a.Hash = common.NewHash(pbAcc.Hash)
	a.Tokens = make(map[string]Token)
	a.Contract = types.DeployInfo{
		TypeVm:   types.VmType(pbAcc.Contract.TypeVm),
		Describe: common.CopyBytes(pbAcc.Contract.Describe),
		Code:     common.CopyBytes(pbAcc.Contract.Code),
	}
	a.Permissions = make(map[string]Permission, 1)
	for _, v := range pbAcc.Tokens {
		ac := Token{
			Name:    string(v.Name),
			Balance: new(big.Int),
		}
		if err := ac.Balance.GobDecode(v.Balance); err != nil {
			return err
		}
		a.Tokens[ac.Name] = ac
	}
	for _, v := range pbAcc.Delegates {
		a.Delegates = append(a.Delegates, Delegate{Index: common.AccountName(v.Index), CpuStaked: v.Cpu, NetStaked: v.Net})
	}
	for _, pbPerm := range pbAcc.Permissions {
		keys := make(map[string]KeyFactor, 1)
		for _, pbKey := range pbPerm.Keys {
			key := KeyFactor{Actor: common.NewAddress(pbKey.Actor), Weight: pbKey.Weight}
			keys[common.NewAddress(pbKey.Actor).HexString()] = key
		}
		accounts := make(map[string]AccFactor, 1)
		for _, pbAcc := range pbPerm.Accounts {
			acc := AccFactor{Actor: common.AccountName(pbAcc.Actor), Weight: pbAcc.Weight, Permission: string(pbAcc.Permission)}
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
	fmt.Println("----------------" + common.IndexToName(a.Index) + ":")
	fmt.Println(a.JsonString())
}
