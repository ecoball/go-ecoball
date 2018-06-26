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
	AccountGetBalance(addr common.Address, token string) (uint64, error)
	AccountAddBalance(addr common.Address, token string, value uint64) error
	AccountSubBalance(addr common.Address, token string, value uint64) error
	ContractGetInfo(key []byte) ([]byte, error)
	TokenCreate(addr common.Address, token string, maximum uint64) error
	//TokenIsExisted(name string) bool
	//SignatureTransaction()
	//GetContractInfo()
}
