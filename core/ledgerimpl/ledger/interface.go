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

	AccountAdd(index common.AccountName, addr common.Address) (*state.Account, error)
	SetContract(index common.AccountName, t types.VmType, des, code []byte) error
	GetContract(index common.AccountName) (*types.DeployInfo, error)
	AccountGet(index common.AccountName) (*state.Account, error)
	AddPermission(index common.AccountName, perm state.Permission) error
	FindPermission(index common.AccountName, name string) (string, error)
	CheckPermission(index common.AccountName, name string, sig []common.Signature) error
	AccountGetBalance(index common.AccountName, token string) (uint64, error)
	AccountAddBalance(index common.AccountName, token string, value uint64) error
	AccountSubBalance(index common.AccountName, token string, value uint64) error

	SetResourceLimits(index common.AccountName, cpu, net float32) error
	StoreGet(index common.AccountName, key []byte) ([]byte, error)
	StoreSet(index common.AccountName, key, value []byte) error

	TokenCreate(index common.AccountName, token string, maximum uint64) error
	TokenIsExisted(token string) bool
	Start()

	GetChainTx() ChainInterface
}
