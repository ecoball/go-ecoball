package ledger

import (
	"github.com/ecoball/go-ecoball/common"
	"github.com/ecoball/go-ecoball/core/types"
)

type Ledger interface {
	NewTxBlock(txs []*types.Transaction, consensusData types.ConsensusData) (*types.Block, error)
	SaveTxBlock(block *types.Block) error
	GetTxBlock(hash common.Hash) (*types.Block, error)
	CheckTransaction(tx *types.Transaction) error
	GetCurrentHeader() *types.Header
	GetCurrentHeight() uint64
	GetAccountBalance(addr common.Address) (uint64, error)
	AddAccountBalance(addr common.Address, value uint64) error
	SubAccountBalance(addr common.Address, value uint64) error
	GetContractInfo(key []byte) ([]byte, error)
	//SignatureTransaction()
	//GetContractInfo()
}
