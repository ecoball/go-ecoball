package trie

import (
	"github.com/ecoball/go-ecoball/common"
	"github.com/ecoball/go-ecoball/common/elog"
	"github.com/ecoball/go-ecoball/core/store"
	"sync"
	"time"
)

const secureKeyLength = 11 + 32

var secureKeyPrefix = []byte("secure-key-")
var log = elog.NewLogger("trie", elog.InfoLog)

type DatabaseReader interface {
	Get(key []byte) (value []byte, err error)
	Has(key []byte) (bool, error)
}

type Database struct {
	diskDB store.Database // Persistent storage for matured trie nodes

	nodes     map[common.Hash]*cachedNode // Data and references relationships of a node
	preImages map[common.Hash][]byte      // PreImages of nodes from the secure trie
	secKeyBuf [secureKeyLength]byte       // Ephemeral buffer for calculating preImage keys

	gcTime  time.Duration      // Time spent on garbage collection since last commit
	gcNodes uint64             // Nodes garbage collected since last commit
	gcSize  common.StorageSize // Data storage garbage collected since last commit

	nodesSize     common.StorageSize // Storage size of the nodes cache
	preImagesSize common.StorageSize // Storage size of the preImages cache

	lock sync.RWMutex
}

type cachedNode struct {
	blob     []byte              // Cached data block of the trie node
	parents  int                 // Number of live nodes referencing this one
	children map[common.Hash]int // Children referenced by this nodes
}

func NewDatabase(diskDB store.Database) *Database {
	return &Database{
		diskDB: diskDB,
		nodes: map[common.Hash]*cachedNode{
			{}: {children: make(map[common.Hash]int)},
		},
		preImages: make(map[common.Hash][]byte),
	}
}

func (db *Database) DiskDB() DatabaseReader {
	return db.diskDB
}

func (db *Database) Insert(hash common.Hash, blob []byte) {
	db.lock.Lock()
	defer db.lock.Unlock()

	db.insert(hash, blob)
}

func (db *Database) insert(hash common.Hash, blob []byte) {
	if _, ok := db.nodes[hash]; ok {
		return
	}
	db.nodes[hash] = &cachedNode{
		blob:     common.CopyBytes(blob),
		children: make(map[common.Hash]int),
	}
	db.nodesSize += common.StorageSize(common.HashLen + len(blob))
}

func (db *Database) insertPreImage(hash common.Hash, preimage []byte) {
	if _, ok := db.preImages[hash]; ok {
		return
	}
	db.preImages[hash] = common.CopyBytes(preimage)
	db.preImagesSize += common.StorageSize(common.HashLen + len(preimage))
}

func (db *Database) Node(hash common.Hash) ([]byte, error) {
	// Retrieve the node from cache if available
	db.lock.RLock()
	node := db.nodes[hash]
	db.lock.RUnlock()

	if node != nil {
		return node.blob, nil
	}
	// Content unavailable in memory, attempt to retrieve from disk
	return db.diskDB.Get(hash[:])
}

func (db *Database) preImage(hash common.Hash) ([]byte, error) {
	// Retrieve the node from cache if available
	db.lock.RLock()
	preImage := db.preImages[hash]
	db.lock.RUnlock()

	if preImage != nil {
		return preImage, nil
	}
	// Content unavailable in memory, attempt to retrieve from disk
	return db.diskDB.Get(db.secureKey(hash[:]))
}

func (db *Database) secureKey(key []byte) []byte {
	buf := append(db.secKeyBuf[:0], secureKeyPrefix...)
	buf = append(buf, key...)
	return buf
}

func (db *Database) Nodes() []common.Hash {
	db.lock.RLock()
	defer db.lock.RUnlock()

	var hashes = make([]common.Hash, 0, len(db.nodes))
	for hash := range db.nodes {
		if hash != (common.Hash{}) { // Special case for "root" references/nodes
			hashes = append(hashes, hash)
		}
	}
	return hashes
}

func (db *Database) Reference(child common.Hash, parent common.Hash) {
	db.lock.RLock()
	defer db.lock.RUnlock()

	db.reference(child, parent)
}

func (db *Database) reference(child common.Hash, parent common.Hash) {
	// If the node does not exist, it's a node pulled from disk, skip
	node, ok := db.nodes[child]
	if !ok {
		return
	}
	// If the reference already exists, only duplicate for roots
	if _, ok = db.nodes[parent].children[child]; ok && parent != (common.Hash{}) {
		return
	}
	node.parents++
	db.nodes[parent].children[child]++
}

func (db *Database) Dereference(child common.Hash, parent common.Hash) {
	db.lock.Lock()
	defer db.lock.Unlock()

	nodes, storage, start := len(db.nodes), db.nodesSize, time.Now()
	db.dereference(child, parent)

	db.gcNodes += uint64(nodes - len(db.nodes))
	db.gcSize += storage - db.nodesSize
	db.gcTime += time.Since(start)

	log.Notice("DeReferenced trie from memory database", "nodes", nodes-len(db.nodes), "size", storage-db.nodesSize, "time", time.Since(start),
		"gcNodes", db.gcNodes, "gcSize", db.gcSize, "gcTime", db.gcTime, "liveNodes", len(db.nodes), "liveSize", db.nodesSize)
}

// dereference is the private locked version of Dereference.
func (db *Database) dereference(child common.Hash, parent common.Hash) {
	// Dereference the parent-child
	node := db.nodes[parent]

	node.children[child]--
	if node.children[child] == 0 {
		delete(node.children, child)
	}
	// If the node does not exist, it's a previously committed node.
	node, ok := db.nodes[child]
	if !ok {
		return
	}
	// If there are no more references to the child, delete it and cascade
	node.parents--
	if node.parents == 0 {
		for hash := range node.children {
			db.dereference(hash, child)
		}
		delete(db.nodes, child)
		db.nodesSize -= common.StorageSize(common.HashLen + len(node.blob))
	}
}

func (db *Database) Commit(node common.Hash, report bool) error {
	db.lock.RLock()

	start := time.Now()
	batch := db.diskDB.NewBatch()

	for hash, preImage := range db.preImages {
		if err := batch.Put(db.secureKey(hash[:]), preImage); err != nil {
			log.Error("Failed to commit preImage from trie database", "err", err)
			db.lock.RUnlock()
			return err
		}
		if batch.ValueSize() > store.IdealBatchSize {
			if err := batch.Write(); err != nil {
				return err
			}
			batch.Reset()
		}
	}
	nodes, storage := len(db.nodes), db.nodesSize+db.preImagesSize
	if err := db.commit(node, batch); err != nil {
		log.Error("Failed to commit trie from trie database", "err", err)
		db.lock.RUnlock()
		return err
	}
	if err := batch.Write(); err != nil {
		log.Error("Failed to write trie to disk", "err", err)
		db.lock.RUnlock()
		return err
	}
	db.lock.RUnlock()

	db.lock.Lock()
	defer db.lock.Unlock()

	db.preImages = make(map[common.Hash][]byte)
	db.preImagesSize = 0

	db.unCache(node)

	logger := log.Info
	if !report {
		logger = log.Debug
	}
	logger("Persisted trie from memory database", "nodes", nodes-len(db.nodes), "size", storage-db.nodesSize, "time", time.Since(start),
		"gcNodes", db.gcNodes, "gcSize", db.gcSize, "gcTime", db.gcTime, "liveNodes", len(db.nodes), "liveSize", db.nodesSize)

	db.gcNodes, db.gcSize, db.gcTime = 0, 0, 0

	return nil
}

func (db *Database) commit(hash common.Hash, batch store.Batch) error {
	node, ok := db.nodes[hash]
	if !ok {
		return nil
	}
	for child := range node.children {
		if err := db.commit(child, batch); err != nil {
			return err
		}
	}
	if err := batch.Put(hash[:], node.blob); err != nil {
		return err
	}
	if batch.ValueSize() >= store.IdealBatchSize {
		if err := batch.Write(); err != nil {
			return err
		}
		batch.Reset()
	}
	return nil
}

func (db *Database) unCache(hash common.Hash) {
	node, ok := db.nodes[hash]
	if !ok {
		return
	}
	for child := range node.children {
		db.unCache(child)
	}
	delete(db.nodes, hash)
	db.nodesSize -= common.StorageSize(common.HashLen + len(node.blob))
}

func (db *Database) Size() common.StorageSize {
	db.lock.RLock()
	defer db.lock.RUnlock()
	return db.nodesSize + db.preImagesSize
}
