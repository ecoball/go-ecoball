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
	"github.com/ecoball/go-ecoball/common/elog"
	"github.com/ecoball/go-ecoball/core/store"
	"math/big"
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
	return st, nil
}
/**
 *  @brief create a new account and store into mpt trie, meanwhile store the mapping of addr and index
 *  @param index - account's index
 *  @param addr - account's address convert from public key
 */
func (s *State) AddAccount(index common.AccountName, addr common.Address) (*Account, error) {
	key := common.IndexToBytes(index)
	acc, err := s.trie.TryGet(key)
	if err != nil {
		return nil, err
	}
	if acc != nil {
		return nil, errors.New("reduplicate name")
	}
	obj, err := NewAccount(s.path, index, addr)
	if err != nil {
		return nil, err
	}
	if err := s.CommitAccount(obj); err != nil {
		return nil, err
	}
	//save the mapping of addr and index
	if err := s.trie.TryUpdate(addr.Bytes(), common.IndexToBytes(obj.Index)); err != nil {
		return nil, err
	}
	return obj, nil
}
func (s *State) PledgeCpu(index common.AccountName, token string, value *big.Int) error {
	acc, err := s.GetAccountByName(index)
	if err != nil {
		return err
	}
	if err := acc.PledgeCpu(token, value); err != nil {
		return err
	}
	return s.CommitAccount(acc)
}
func (s *State) CancelPledgeCpu(index common.AccountName, token string, value *big.Int) error {
	acc, err := s.GetAccountByName(index)
	if err != nil {
		return err
	}
	if err := acc.CancelPledgeCpu(token, value); err != nil {
		return err
	}
	return s.CommitAccount(acc)
}
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
 *  @brief add a permission object into account, then update to mpt trie
 *  @param perm - the permission object
 */
func (s *State) AddPermission(index common.AccountName, perm Permission) error {
	acc, err := s.GetAccountByName(index)
	if err != nil {
		return err
	}
	acc.AddPermission(perm)
	return s.CommitAccount(acc)
}
/**
 *  @brief check the permission's validity, this method will not modified mpt trie
 *  @param index - the account index
 *  @param state - the world state tree
 *  @param name - the permission names
 *  @param signatures - the signatures list
 */
func (s *State) CheckPermission(index common.AccountName, name string, signatures []common.Signature) error {
	acc, err := s.GetAccountByName(index)
	if err != nil {
		return err
	}
	return acc.CheckPermission(s, name, signatures)
}
/**
 *  @brief search the permission by name, return json array string
 *  @param index - the account index
 *  @param name - the permission names
 */
func (s *State) FindPermission(index common.AccountName, name string) (string, error) {
	acc, err := s.GetAccountByName(index)
	if err != nil {
		return "", err
	}
	if str, err := acc.FindPermission(name); err != nil {
		return "", err
	} else {
		return "[" + str + "]", nil
	}
}
/**
 *  @brief search the account by name index
 *  @param index - the account index
 */
func (s *State) GetAccountByName(index common.AccountName) (*Account, error) {
	key := common.IndexToBytes(index)
	fData, err := s.trie.TryGet(key)
	if err != nil {
		return nil, err
	}
	if fData == nil {
		return nil, errors.New(fmt.Sprintf("no this account named:%s", common.IndexToName(index)))
	}
	acc := new(Account)
	if err := acc.Deserialize(fData); err != nil {
		return nil, err
	}
	return acc, nil
}
/**
 *  @brief search the account by address
 *  @param addr - the account address
 */
func (s *State) GetAccountByAddr(addr common.Address) (*Account, error) {
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
	d, err := acc.Serialize()
	if err != nil {
		return err
	}
	if err := s.trie.TryUpdate(common.IndexToBytes(acc.Index), d); err != nil {
		return err
	}
	return nil
}

func (s *State) GetBalance(index common.AccountName, token string) (*big.Int, error) {
	acc, err := s.GetAccountByName(index)
	if err != nil {
		return nil, err
	}

	return acc.Balance(token)
}
func (s *State) SubBalance(index common.AccountName, token string, value *big.Int) error {
	acc, err := s.GetAccountByName(index)
	if err != nil {
		return err
	}

	balance, err := acc.Balance(token)
	if err != nil {
		return err
	}
	if balance.Cmp(value) == -1 {
		return errors.New("no enough balance")
	}
	acc.SubBalance(token, value)
	if err := s.CommitAccount(acc); err != nil {
		return err
	}
	return nil
}
func (s *State) AddBalance(index common.AccountName, token string, value *big.Int) error {
	acc, err := s.GetAccountByName(index)
	if err != nil {
		return err
	}
	acc.AddBalance(token, value)
	if err := s.CommitAccount(acc); err != nil {
		return err
	}

	return nil
}
func (s *State) CreateToken(token string, value *big.Int) error {
	//add token into trie
	data, err := value.GobEncode()
	if err != nil {
		return err
	}
	if err := s.trie.TryUpdate([]byte(token), data); err != nil {
		return err
	}
	return nil
}
func (s *State) GetToken(token string) (*big.Int, error) {
	if data, err := s.trie.TryGet([]byte(token)); err != nil {
		return nil, err
	} else {
		value := new(big.Int)
		if err := value.GobDecode(data); err != nil {
			return nil, err
		}
		return value, nil
	}
}
func (s *State) TokenExisted(name string) bool {
	data, err := s.trie.TryGet([]byte(name))
	if err != nil {
		log.Error(err)
		return false
	}
	return string(data) == name
}

func (s *State) GetHashRoot() common.Hash {
	return common.NewHash(s.trie.Hash().Bytes())
}

func (s *State) CommitToMemory() error {
	root, err := s.trie.Commit(nil)
	if err != nil {
		return err
	}
	log.Debug("Commit State DB:", root.HexString())
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
