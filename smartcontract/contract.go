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
	"fmt"
	"github.com/ecoball/go-ecoball/core/state"
	"github.com/ecoball/go-ecoball/core/types"
	"github.com/ecoball/go-ecoball/smartcontract/nativeservice"
	"github.com/ecoball/go-ecoball/smartcontract/wasmservice"
)

type ContractService interface {
	Execute() ([]byte, error)
}

func NewContractService(s *state.State, tx *types.Transaction) (ContractService, error) {
	if s == nil || tx == nil {
		return nil, errors.New("the contract service's ledger interface or tx is nil")
	}
	contract, err := s.GetContract(tx.Addr)
	if err != nil {
		return nil, err
	}
	invoke, ok := tx.Payload.GetObject().(types.InvokeInfo)
	if !ok {
		return nil, errors.New("transaction type error[invoke]")
	}
	fmt.Println("method:", string(invoke.Method))
	fmt.Println("param:", invoke.Param)
	switch contract.TypeVm {
	case types.VmNative:
		service, err := nativeservice.NewNativeService(s, tx.Addr, string(invoke.Method), invoke.Param)
		if err != nil {
			return nil, err
		}
		return service, nil
	case types.VmWasm:
		service, err := wasmservice.NewWasmService(s, tx, contract, &invoke)
		if err != nil {
			return nil, err
		}
		return service, nil
	default:
		return nil, errors.New("unknown virtual machine")
	}
}
