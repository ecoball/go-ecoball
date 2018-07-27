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
	"github.com/ecoball/go-ecoball/common/elog"
	"github.com/ecoball/go-ecoball/core/store"
	"github.com/ecoball/go-ecoball/core/types"
)

var log = elog.NewLogger("state", elog.DebugLog)
var IndexAbaRoot = common.NameToIndex("root")
var AbaToken = "ABA"

type State struct {
	path   string
	trie   Trie
	db     Database
	diskDb *store.LevelDBStore

	Accounts map[string]Account
	Params   map[string]uint64
}

/**
 *  @brief create a new mpt trie and a levelDB
 *  @param path - the levelDB store path
 *  @param root - the root of mpt trie, this value decide the state of trie
 */
func NewState(path string, root common.Hash) (st *State, err error) {
	st = &State{path: path}
	st.diskDb, err = store.NewLevelDBStore(path, 0, 0)
	if err != nil {
		return nil, err
	}
	st.db = NewDatabase(st.diskDb)
	log.Notice("Open Trie Hash:", root.HexString())
	st.trie, err = st.db.OpenTrie(root)
	if err != nil {
		st.trie, _ = st.db.OpenTrie(common.Hash{})
	}
	st.Accounts = make(map[string]Account, 1)
	st.Params = make(map[string]uint64, 1)
	return st, nil
}
func (s *State) CopyState() (*State, error) {
	params := make(map[string]uint64, 1)
	accounts := make(map[string]Account, 1)

	if str, err := json.Marshal(s.Params); err != nil {
		return nil, err
	} else {
		if err := json.Unmarshal(str, &params); err != nil {
			return nil, err
		}
	}
	if str, err := json.Marshal(s.Accounts); err != nil {
		return nil, err
	} else {
		if err := json.Unmarshal(str, &accounts); err != nil {
			return nil, err
		}
	}
	return &State{
		path:     s.path,
		trie:     s.db.CopyTrie(s.trie),
		Accounts: accounts,
		Params:   params,
	}, nil
}

/**
 *  @brief create a new account and store into mpt trie, meanwhile store the mapping of addr and index
 *  @param index - account's index
 *  @param addr - account's address convert from public key
 */
func (s *State) AddAccount(index common.AccountName, addr common.Address) (*Account, error) {
	key := common.IndexToBytes(index)
	data, err := s.trie.TryGet(key)
	if err != nil {
		return nil, err
	}
	if data != nil {
		return nil, errors.New("reduplicate name")
	}
	acc, err := NewAccount(s.path, index, addr)
	if err != nil {
		return nil, err
	}
	if err := s.CommitAccount(acc); err != nil {
		return nil, err
	}
	//save the mapping of addr and index
	if err := s.trie.TryUpdate(addr.Bytes(), common.IndexToBytes(acc.Index)); err != nil {
		return nil, err
	}
	s.Accounts[common.IndexToName(index)] = *acc
	s.Params[addr.HexString()] = uint64(index)
	return acc, nil
}

/**
 *  @brief store the smart contract of account, every account only has one contract
 *  @param index - account's index
 *  @param t - the virtual machine type
 *  @param des - the description of contract
 *  @param code - the code of contract
 */
func (s *State) SetContract(index common.AccountName, t types.VmType, des, code []byte) error {
	acc, err := s.GetAccountByName(index)
	if err != nil {
		return err
	}
	if err := acc.SetContract(t, des, code); err != nil {
		return err
	}
	return s.CommitAccount(acc)
}

/**
 *  @brief get the code of account
 *  @param index - account's index
 */
func (s *State) GetContract(index common.AccountName) (*types.DeployInfo, error) {
	acc, err := s.GetAccountByName(index)
	if err != nil {
		return nil, err
	}
	return acc.GetContract()
}
func (s *State) StoreSet(index common.AccountName, key, value []byte) (err error) {
	acc, err := s.GetAccountByName(index)
	if err != nil {
		return err
	}
	if err := acc.StoreSet(s.path, key, value); err != nil {
		return err
	}
	return s.CommitAccount(acc)
}
func (s *State) StoreGet(index common.AccountName, key []byte) (value []byte, err error) {
	acc, err := s.GetAccountByName(index)
	if err != nil {
		return nil, err
	}
	return acc.StoreGet(s.path, key)
}

/**
 *  @brief search the account by name index
 *  @param index - the account index
 */
func (s *State) GetAccountByName(index common.AccountName) (*Account, error) {
	acc, ok := s.Accounts[common.IndexToName(index)]
	if ok {
		return &acc, nil
	}
	key := common.IndexToBytes(index)
	fData, err := s.trie.TryGet(key)
	if err != nil {
		return nil, err
	}
	if fData == nil {
		return nil, errors.New(fmt.Sprintf("no this account named:%s", common.IndexToName(index)))
	}
	acc = Account{}
	if err := acc.Deserialize(fData); err != nil {
		return nil, err
	}
	return &acc, nil
}

/**
 *  @brief search the account by address
 *  @param addr - the account address
 */
func (s *State) GetAccountByAddr(addr common.Address) (*Account, error) {
	index, ok := s.Params[addr.HexString()]
	if ok {
		return s.GetAccountByName(common.AccountName(index))
	}
	if fData, err := s.trie.TryGet(addr.Bytes()); err != nil {
		return nil, err
	} else {
		if fData == nil {
			return nil, errors.New(fmt.Sprintf("can't find this account by address:%s", addr.HexString()))
		} else {
			acc, err := s.GetAccountByName(common.IndexSetBytes(fData))
			if err != nil {
				return nil, err
			}
			return acc, nil
		}
	}
}

/**
 *  @brief update the account's information into trie
 *  @param acc - account object
 */
func (s *State) CommitAccount(acc *Account) error {
	if acc == nil {
		return errors.New("param acc is nil")
	}
	d, err := acc.Serialize()
	if err != nil {
		return err
	}
	if err := s.trie.TryUpdate(common.IndexToBytes(acc.Index), d); err != nil {
		return err
	}
	s.RecoverResources(acc)
	s.Accounts[common.IndexToName(acc.Index)] = *acc
	return nil
}
func (s *State) CommitParam(key string, value uint64) error {
	if err := s.trie.TryUpdate([]byte(key), common.Uint64ToBytes(value)); err != nil {
		return err
	}
	s.Params[key] = value
	return nil
}
func (s *State) GetParam(key string) (uint64, error) {
	value, ok := s.Params[key]
	if ok {
		return value, nil
	}
	data, err := s.trie.TryGet([]byte(key))
	log.Warn(data, err)
	if err != nil {
		return 0, err
	}
	if len(data) == 0 {
		return 0, nil
	}
	value = common.Uint64SetBytes(data)
	s.Params[key] = value
	return value, nil
}

func (s *State) GetHashRoot() common.Hash {
	return common.NewHash(s.trie.Hash().Bytes())
}

func (s *State) CommitToMemory() error {
	root, err := s.trie.Commit(nil)
	if err != nil {
		return err
	}
	log.Debug("commit state db to memory:", root.HexString())
	return nil
}

/**
 *  @brief save the information of mpt trie into levelDB
 */
func (s *State) CommitToDB() error {
	if err := s.CommitToMemory(); err != nil {
		return err
	}
	return s.db.TrieDB().Commit(s.trie.Hash(), false)
}

/**
 *  @brief reset the mpt state by root hash
 *  @param hash - the hash of mpt witch state will be reset
 */
func (s *State) Reset(hash common.Hash) error {
	s.Close()
	diskDb, err := store.NewLevelDBStore(s.path, 0, 0)
	if err != nil {
		return err
	}
	s.db = NewDatabase(diskDb)
	log.Notice("Open Trie Hash:", hash.HexString())
	s.trie, err = s.db.OpenTrie(hash)
	if err != nil {
		s.trie, err = s.db.OpenTrie(common.Hash{})
		if err != nil {
			return err
		}
	}
	return nil
}
func (s *State) Close() {
	s.diskDb.Close()
}
func (s *State) Trie() Trie {
	return s.trie
}
func (s *State) DataBase() Database {
	return s.db
}
