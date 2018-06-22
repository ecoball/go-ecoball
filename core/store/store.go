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

package store

import (
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/errors"
	"github.com/syndtr/goleveldb/leveldb/filter"
	"github.com/syndtr/goleveldb/leveldb/iterator"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"github.com/syndtr/goleveldb/leveldb/util"
)

const PathBlock = "DataBase/block"

type Storage interface {
	Put(key, value []byte) error
	Get(key []byte) ([]byte, error)
	Has(key []byte) (bool, error)
	Delete(key []byte) error
	BatchPut(key, value []byte)
	BatchCommit() error
	SearchAll() (result map[string]string, err error)
	DeleteAll() error
	NewIterator() iterator.Iterator
}

type LevelDBStore struct {
	db    *leveldb.DB
	batch *leveldb.Batch
}

func NewBlockStore(dirPath string) (Storage, error) {
	return NewLevelDBStore(dirPath, 0, 0)
}

func NewLevelDBStore(dirPath string, cache int, handles int) (*LevelDBStore, error) {
	if cache < 16 {
		cache = 16
	}
	if handles < 16 {
		handles = 16
	}
	//Bloom Filter, can quickly search value
	o := opt.Options{
		OpenFilesCacheCapacity: handles,
		BlockCacheCapacity:     cache / 2 * opt.MiB,
		WriteBuffer:            cache / 4 * opt.MiB, // Two of these are used internally
		Filter:                 filter.NewBloomFilter(10),
	}

	db, err := leveldb.OpenFile(dirPath, &o)
	if _, corrupted := err.(*errors.ErrCorrupted); corrupted {
		db, err = leveldb.RecoverFile(dirPath, nil)
	}
	if err != nil {
		return nil, err
	}

	return &LevelDBStore{
		db:    db,
		batch: nil,
	}, nil
}

func (l *LevelDBStore) Put(key, value []byte) error {
	return l.db.Put(key, value, nil)
}

func (l *LevelDBStore) Get(key []byte) ([]byte, error) {
	return l.db.Get(key, nil)
}

func (l *LevelDBStore) Has(key []byte) (bool, error) {
	return l.db.Has(key, nil)
}

func (l *LevelDBStore) Delete(key []byte) error {
	return l.db.Delete(key, nil)
}

func (l *LevelDBStore) BatchPut(key, value []byte) {
	if l.batch == nil {
		l.batch = new(leveldb.Batch)
	}
	l.batch.Put(key, value)
}

func (l *LevelDBStore) BatchCommit() error {
	if l.batch == nil {
		return nil
	}
	if err := l.db.Write(l.batch, nil); err != nil {
		return err
	}
	l.batch = nil
	return nil
}

func (l *LevelDBStore) SearchAll() (result map[string]string, err error) {
	result = make(map[string]string, 0)
	iterate := l.db.NewIterator(nil, nil)
	for iterate.Next() {
		result[string(iterate.Key())] = string(iterate.Value())
	}
	iterate.Release()
	if err := iterate.Error(); err != nil {
		return nil, err
	}
	return result, nil
}

func (l *LevelDBStore) NewIterator() iterator.Iterator {
	return l.db.NewIterator(nil, nil)
}

// NewIteratorWithPrefix returns a iterator to iterate over subset of database content with a particular prefix.
func (l *LevelDBStore) NewIteratorWithPrefix(prefix []byte) iterator.Iterator {
	return l.db.NewIterator(util.BytesPrefix(prefix), nil)
}

func (l *LevelDBStore) LDB() *leveldb.DB {
	return l.db
}

func (l *LevelDBStore) DeleteAll() error {
	l.batch = new(leveldb.Batch)
	iterate := l.db.NewIterator(nil, nil)
	for iterate.Next() {
		l.batch.Delete(iterate.Key())
	}
	iterate.Release()
	if err := iterate.Error(); err != nil {
		return err
	}
	return l.BatchCommit()
}

func (l *LevelDBStore) NewBatch() Batch {
	return &ldbBatch{db: l.db, b: new(leveldb.Batch)}
}

func (l *LevelDBStore) Close() {
	l.db.Close()
}

type ldbBatch struct {
	db   *leveldb.DB
	b    *leveldb.Batch
	size int
}

func (b *ldbBatch) Put(key, value []byte) error {
	b.b.Put(key, value)
	b.size += len(value)
	return nil
}

func (b *ldbBatch) Write() error {
	return b.db.Write(b.b, nil)
}

func (b *ldbBatch) ValueSize() int {
	return b.size
}

func (b *ldbBatch) Reset() {
	b.b.Reset()
	b.size = 0
}
