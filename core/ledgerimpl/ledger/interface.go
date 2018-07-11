package ledger

import (
	"github.com/ecoball/go-ecoball/common"
	"github.com/ecoball/go-ecoball/core/state"
	"github.com/ecoball/go-ecoball/core/types"
)

type Ledger interface {
	GetTxBlock(hash common.Hash) (*types.Block, error)
	NewTxBlock(txs []*types.Transaction, consensusData types.ConsensusData) (*types.Block, error)
	VerifyTxBlock(block *types.Block) error
	SaveTxBlock(block *types.Block) error
	GetTxBlockByHeight(height uint64) (*types.Block, error)
	CheckTransaction(tx *types.Transaction) error
	GetCurrentHeader() *types.Header
	GetCurrentHeight() uint64
	StateDB() *state.State
	ResetStateDB(hash common.Hash) error

	AccountGet(index common.AccountName) (*state.Account, error)
	AccountAdd(index common.AccountName, addr common.Address) (*state.Account, error)
	AccountGetBalance(index common.AccountName, token string) (uint64, error)
	AccountAddBalance(index common.AccountName, token string, value uint64) error
	AccountSubBalance(index common.AccountName, token string, value uint64) error

	TokenCreate(index common.AccountName, token string, maximum uint64) error
	TokenIsExisted(token string) bool
	Start()
}
