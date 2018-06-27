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

package smartcontract

import (
	"errors"
	"github.com/ecoball/go-ecoball/core/ledgerimpl/ledger"
	"github.com/ecoball/go-ecoball/core/types"
	"github.com/ecoball/go-ecoball/smartcontract/wasmservice"
)

type Service interface {
	Execute() []byte
}

type ContractService struct {
	ledger  ledger.Ledger
	tx      *types.Transaction
	Service Service
}

func NewContractService(ledger ledger.Ledger, tx *types.Transaction) (*ContractService, error) {
	if ledger == nil {
		return nil, errors.New("the contract service's ledger interface is nil")
	}
	return &ContractService{ledger: ledger, tx: tx}, nil
}

func (c *ContractService) ExecuteContract(vmType types.VmType, method string, code []byte, arg []uint64) (ret []byte, err error) {
	if c.ledger == nil {
		return nil, errors.New("the contract service's ledger interface is nil")
	}
	switch vmType {
	case types.VmWasm:
		c.Service, err = wasmservice.NewWasmService(c.ledger, method, code, arg)
		if err != nil {
			return nil, err
		}
	}
	return c.Service.Execute(), nil
}

func (c *ContractService) ParseArguments(param string) ([]uint64, error) {
	return nil, nil
}
