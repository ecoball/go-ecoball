package nativeservice

import (
	"errors"
	"fmt"
	"github.com/ecoball/go-ecoball/common"
	"github.com/ecoball/go-ecoball/common/config"
	"github.com/ecoball/go-ecoball/common/elog"
	"github.com/ecoball/go-ecoball/core/ledgerimpl/ledger"
	"encoding/json"
	"github.com/ecoball/go-ecoball/core/state"
)

var log = elog.NewLogger("native", config.LogLevel)

type NativeService struct {
	ledger ledger.Ledger
	owner  common.AccountName
	method string
	params []string
}

func NewNativeService(ledger ledger.Ledger, owner common.AccountName, method string, params []string) (*NativeService, error) {
	ns := &NativeService{ledger: ledger, owner: owner, method: method, params: params}
	return ns, nil
}

func (ns *NativeService) Execute() ([]byte, error) {
	switch ns.owner {
	case common.NameToIndex("root"):
		return ns.RootExecute()
	default:
		return nil, errors.New("unknown native contract's owner")
	}
	return nil, nil
}

func (ns *NativeService) RootExecute() ([]byte, error) {
	switch ns.method {
	case "new_account":
		index := common.NameToIndex(ns.params[0])
		addr := common.FormHexString(ns.params[1])
		if _, err := ns.ledger.AccountAdd(index, addr); err != nil {
			return nil, err
		}
	case "set_account":
		index := common.NameToIndex(ns.params[0])
		perm := state.Permission{Keys:make(map[string]state.KeyFactor, 1), Accounts:make(map[string]state.AccFactor, 1)}
		if err := json.Unmarshal([]byte(ns.params[1]), &perm); err != nil {
			fmt.Println(ns.params[1])
			return nil, err
		}
		if err := ns.ledger.AddPermission(index, perm); err != nil {
			return nil, err
		}
	default:
		return nil, errors.New(fmt.Sprintf("unknown method:%s", ns.method))
	}
	return nil, nil
}
