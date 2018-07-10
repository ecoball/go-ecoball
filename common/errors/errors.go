/*
Copyright QuakerChain. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/
package errors

import "fmt"

type ErrCode int

const (
	ErrNoCode               ErrCode = -2
	ErrNoError              ErrCode = 0
	ErrUnknown              ErrCode = -1
	ErrDuplicatedTx         ErrCode = 45002
	ErrDuplicateInput       ErrCode = 45003
	ErrAssetPrecision       ErrCode = 45004
	ErrTransactionBalance   ErrCode = 45005
	ErrAttributeProgram     ErrCode = 45006
	ErrTransactionContracts ErrCode = 45007
	ErrTransactionPayload   ErrCode = 45008
	ErrDoubleSpend          ErrCode = 45009
	ErrTxHashDuplicate      ErrCode = 45010
	ErrStateUpdaterVaild    ErrCode = 45011
	ErrSummaryAsset         ErrCode = 45012
	ErrXmitFail             ErrCode = 45013
	ErrNoAccount            ErrCode = 45014
	ErrRetryExhausted       ErrCode = 45015
	ErrTxPoolFull           ErrCode = 45016
)

func (err ErrCode) ErrorInfo() string {
	switch err {
	case ErrNoCode:
		return "no error code"
	case ErrNoError:
		return "not an error"
	case ErrUnknown:
		return "unknown error"
	case ErrDuplicatedTx:
		return "duplicated transaction detected"
	case ErrDuplicateInput:
		return "duplicated transaction input detected"
	case ErrAssetPrecision:
		return "invalid asset precision"
	case ErrTransactionBalance:
		return "transaction balance unmatched"
	case ErrAttributeProgram:
		return "attribute program error"
	case ErrTransactionContracts:
		return "invalid transaction smartcontract"
	case ErrTransactionPayload:
		return "invalid transaction payload"
	case ErrDoubleSpend:
		return "double spent transaction detected"
	case ErrTxHashDuplicate:
		return "duplicated transaction hash detected"
	case ErrStateUpdaterVaild:
		return "invalid state updater"
	case ErrSummaryAsset:
		return "invalid summary asset"
	case ErrXmitFail:
		return "transmit error"
	case ErrRetryExhausted:
		return "retry exhausted"
	case ErrTxPoolFull:
		return "tx pool full"
	}

	return fmt.Sprintf("Unknown error? Error code = %d", err)
}

func (err ErrCode) Error() string {
	return err.ErrorInfo()
}

func (err ErrCode) Value() int {
	return int(err)
}
