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
)

var log = elog.NewLogger("state", elog.DebugLog)
var IndexAbaRoot = common.NameToIndex("root")
var IndexAbaToken = common.NameToIndex("aba")

type State struct {
	path   string
	trie   Trie
	db     Database
	diskDb *store.LevelDBStore

	accounts map[common.AccountName]common.Address
}

func NewState(path string, root common.Hash) (st *State, err error) {
	st = &State{path: path}
	st.diskDb, err = store.NewLevelDBStore(path, 0, 0)
	if err != nil {
		return nil, err
	}
	st.accounts = make(map[common.AccountName]common.Address, 1)
	st.db = NewDatabase(st.diskDb)
	log.Notice("Open Trie Hash:", root.HexString())
	st.trie, err = st.db.OpenTrie(root)
	if err != nil {
		st.trie, _ = st.db.OpenTrie(common.Hash{})
	}
	return st, nil
}

func (s *State) Close() {
	s.diskDb.Close()
}

func (s *State) AddAccount(index common.AccountName, addr common.Address) error {
	key := common.IndexToBytes(index)
	acc, err := s.trie.TryGet(key)
	if err != nil {
		return err
	}
	if acc != nil {
		return errors.New("reduplicate name")
	}
	obj, err := NewAccount(index, addr)
	if err != nil {
		return err
	}
	d, err := obj.Serialize()
	if err != nil {
		return err
	}
	if err := s.trie.TryUpdate(common.IndexToBytes(obj.Index), d); err != nil {
		return err
	}
	if err := s.trie.TryUpdate(addr.Bytes(), common.IndexToBytes(obj.Index)); err != nil {
		return err
	}
	return nil
}

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

func (s *State) SubBalance(indexAcc, indexToken common.AccountName, value *big.Int) error {
	acc, err := s.GetAccountByName(indexAcc)
	if err != nil {
		return err
	}

	balance, err := acc.Balance(indexToken)
	if err != nil {
		return err
	}
	if balance.Cmp(value) == -1 {
		return errors.New("no enough balance")
	}
	acc.SubBalance(indexToken, value)
	d, err := acc.Serialize()
	if err != nil {
		return err
	}
	if err := s.trie.TryUpdate(common.IndexToBytes(indexAcc), d); err != nil {
		return err
	}
	return nil
}

func (s *State) AddBalance(indexAcc, indexToken common.AccountName, value *big.Int) error {
	acc, err := s.GetAccountByName(indexAcc)
	if err != nil {
		return err
	}
	acc.AddBalance(indexToken, value)
	d, err := acc.Serialize()
	if err != nil {
		return err
	}
	if err := s.trie.TryUpdate(common.IndexToBytes(indexAcc), d); err != nil {
		return err
	}
	//add token into trie
	if err := s.trie.TryUpdate(common.IndexToBytes(indexToken), common.IndexToBytes(indexToken)); err != nil {
		return err
	}
	return nil
}

func (s *State) TokenExisted(indexToken common.AccountName) bool {
	data, err := s.trie.TryGet(common.IndexToBytes(indexToken))
	if err != nil {
		log.Error(err)
		return false
	}
	return common.IndexSetBytes(data) == indexToken
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

func (s *State) CommitToDB() error {
	if err := s.CommitToMemory(); err != nil {
		return err
	}
	return s.db.TrieDB().Commit(s.trie.Hash(), false)
}

func (s *State) GetBalance(indexAcc, indexToken common.AccountName) (*big.Int, error) {
	acc, err := s.GetAccountByName(indexAcc)
	if err != nil {
		return nil, err
	}

	return acc.Balance(indexToken)
}

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
