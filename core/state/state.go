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
	"bytes"
	"errors"
	"github.com/ecoball/go-ecoball/common"
	"github.com/ecoball/go-ecoball/common/elog"
	"github.com/ecoball/go-ecoball/core/store"
	"math/big"
)

var log = elog.NewLogger("state", elog.DebugLog)
var AbaToken = "Aba"

type State struct {
	path     string
	diskDb   *store.LevelDBStore
	hashRoot common.Hash
	trie     Trie
	db       Database
}

func NewState(path string, root common.Hash) (st *State, err error) {
	st = new(State)
	diskDb, err := store.NewLevelDBStore(path, 0, 0)
	if err != nil {
		return nil, err
	}
	st.db = NewDatabase(diskDb)
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

func (s *State) GetStateObject(addr common.Address) (*StateObject, error) {
	fHash := common.SingleHash(addr.Bytes())
	fData, err := s.trie.TryGet(fHash.Bytes())
	if err != nil {
		return nil, err
	}
	fObj := new(StateObject)
	if fData == nil {
		if fObj, err = NewStateObject(addr); err != nil {
			return nil, err
		}
	} else {
		if err := fObj.Deserialize(fData); err != nil {
			return nil, err
		}
	}
	return fObj, nil
}

func (s *State) SubBalance(addr common.Address, name string, value *big.Int) error {
	fObj, err := s.GetStateObject(addr)
	if err != nil {
		return err
	}
	balance, err := fObj.Balance(name)
	if err != nil {
		return errors.New("no enough balance")
	}
	if balance.Cmp(value) == -1 {
		return errors.New("no enough balance")
	}
	fObj.SubBalance(name, value)
	d, err := fObj.Serialize()
	if err != nil {
		return err
	}
	if err := s.trie.TryUpdate(fObj.addrHash.Bytes(), d); err != nil {
		return err
	}
	return nil
}

func (s *State) AddBalance(addr common.Address, name string, value *big.Int) error {
	tObj, err := s.GetStateObject(addr)
	if err != nil {
		return err
	}
	tObj.AddBalance(name, value)
	d, err := tObj.Serialize()
	if err != nil {
		return err
	}
	if err := s.trie.TryUpdate(tObj.addrHash.Bytes(), d); err != nil {
		return err
	}
	//add token into trie
	hash := common.SingleHash([]byte(name))
	if err := s.trie.TryUpdate(hash.Bytes(), []byte(name)); err != nil {
		return err
	}
	return nil
}

func (s *State) TokenExisted(name string) bool {
	hash := common.SingleHash([]byte(name))
	data, err := s.trie.TryGet(hash.Bytes())
	if err != nil {
		log.Error(err)
		return false
	}
	return bytes.Equal(data, []byte(name))
}

func (s *State) GetHashRoot() common.Hash {
	s.hashRoot = common.NewHash(s.trie.Hash().Bytes())
	return s.hashRoot
}

func (s *State) CommitToMemory() error {
	root, err := s.trie.Commit(nil)
	if err != nil {
		return err
	}
	s.hashRoot = common.BytesToHash(root.Bytes())
	return nil
}

func (s *State) CommitToDB() error {
	if err := s.CommitToMemory(); err != nil {
		return err
	}
	return s.db.TrieDB().Commit(s.trie.Hash(), false)
}

func (s *State) GetBalance(addr common.Address, token string) (*big.Int, error) {
	val := new(big.Int).SetUint64(0)
	key := common.SingleHash(addr.Bytes())
	data, err := s.trie.TryGet(key.Bytes())
	if err != nil {
		log.Error(err)
		return val, err
	}
	if data == nil {
		return nil, errors.New("can't find account in MPT tree")
	}
	obj := new(StateObject)
	err = obj.Deserialize(data)
	if err != nil {
		log.Error(err)
		return val, err
	}

	return obj.Balance(token)
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
		s.trie, _ = s.db.OpenTrie(common.Hash{})
	}
	return nil
}
