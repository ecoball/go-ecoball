package ledger

import (
	"github.com/ecoball/go-ecoball/common"
	"github.com/ecoball/go-ecoball/core/state"
	"github.com/ecoball/go-ecoball/core/types"
)

type Ledger interface {
	NewTxBlock(txs []*types.Transaction, consensusData types.ConsensusData) (*types.Block, error)
	VerifyTxBlock(block *types.Block) error
	SaveTxBlock(block *types.Block) error
	GetTxBlock(hash common.Hash) (*types.Block, error)
	GetTxBlockByHeight(height uint64) (*types.Block, error)
	CheckTransaction(tx *types.Transaction) error
	GetCurrentHeader() *types.Header
	GetCurrentHeight() uint64
	StateDB() *state.State
	ResetStateDB(hash common.Hash) error

	AccountAdd(index common.AccountName, addr common.Address) error
	GetAccount(index common.AccountName) (*state.Account, error)
	AccountGetBalance(indexAcc, indexToken common.AccountName) (uint64, error)
	AccountAddBalance(indexAcc, indexToken common.AccountName, value uint64) error
	AccountSubBalance(indexAcc, indexToken common.AccountName, value uint64) error
	ContractGetInfo(key []byte) ([]byte, error)

	TokenCreate(indexAcc, indexToken common.AccountName, maximum uint64) error
	TokenIsExisted(indexToken common.AccountName) bool
	//SignatureTransaction()
	//GetContractInfo()
	Start()
}
