package trie_test

import (
	"encoding/hex"
	"fmt"
	"github.com/ecoball/go-ecoball/common"
	"github.com/ecoball/go-ecoball/core/trie"
	"testing"
)

func TestMerkleTree(t *testing.T) {
	fmt.Println("test merkle tree")

	var hashes []common.Hash
	hx, _ := hex.DecodeString("9f0116a5d819943920cec5d248c922e52dfad475a406d730eb5680a856baf003")
	h := common.NewHash(hx)
	hashes = append(hashes, h)
	merkleRoot, _ := trie.GetMerkleRoot(hashes)
	fmt.Println(merkleRoot.HexString())

	hx, _ = hex.DecodeString("12fbbbca6d41fa262e610e26af488f4ce9a8b7f9dd47025d03b5a33fdc7a0d66")
	h = common.NewHash(hx)
	hashes = append(hashes, h)
	merkleRoot, _ = trie.GetMerkleRoot(hashes)
	fmt.Println(merkleRoot.HexString())

	hx, _ = hex.DecodeString("9fef5218557442a89ee05a736413f0e9a48cd97fab0d560b5709d8c820f6c2e2")
	h = common.NewHash(hx)
	hashes = append(hashes, h)
	merkleRoot, _ = trie.GetMerkleRoot(hashes)
	fmt.Println(merkleRoot.HexString())

}
