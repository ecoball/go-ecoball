package nativeservice

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ecoball/go-ecoball/common"
	"github.com/ecoball/go-ecoball/common/config"
	"github.com/ecoball/go-ecoball/common/elog"
	"github.com/ecoball/go-ecoball/core/state"
	"strconv"
)

var log = elog.NewLogger("native", config.LogLevel)

type NativeService struct {
	state  *state.State
	owner  common.AccountName
	method string
	params []string
}

func NewNativeService(s *state.State, owner common.AccountName, method string, params []string) (*NativeService, error) {
	ns := &NativeService{state: s, owner: owner, method: method, params: params}
	return ns, nil
}

func (ns *NativeService) Execute() ([]byte, error) {
	switch ns.owner {
	case common.NameToIndex("root"):
		return ns.RootExecute()
	case common.NameToIndex("worker1"):
		return ns.SystemExecute(ns.owner)
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
		if _, err := ns.state.AddAccount(index, addr); err != nil {
			return nil, err
		}
	case "set_account":
		index := common.NameToIndex(ns.params[0])
		perm := state.Permission{Keys: make(map[string]state.KeyFactor, 1), Accounts: make(map[string]state.AccFactor, 1)}
		if err := json.Unmarshal([]byte(ns.params[1]), &perm); err != nil {
			fmt.Println(ns.params[1])
			return nil, err
		}
		if err := ns.state.AddPermission(index, perm); err != nil {
			return nil, err
		}
	default:
		return nil, errors.New(fmt.Sprintf("unknown method:%s", ns.method))
	}
	return nil, nil
}

func (ns *NativeService) SystemExecute(index common.AccountName) ([]byte, error) {
	switch ns.method {
	case "pledge":
		from := common.NameToIndex(ns.params[0])
		to := common.NameToIndex(ns.params[1])
		cpu, err := strconv.ParseUint(ns.params[2], 10, 64)
		if err != nil {
			return nil, err
		}
		net, err := strconv.ParseUint(ns.params[3], 10, 64)
		if err != nil {
			return nil, err
		}

		if err := ns.state.SetResourceLimits(from, to, cpu, net); err != nil {
			return nil, err
		}
		return nil, nil
	case "cancel_pledge":
		from := common.NameToIndex(ns.params[0])
		to := common.NameToIndex(ns.params[1])
		cpu, err := strconv.ParseUint(ns.params[2], 10, 64)
		if err != nil {
			return nil, err
		}
		net, err := strconv.ParseUint(ns.params[3], 10, 64)
		if err != nil {
			return nil, err
		}
		log.Debug(from, to, cpu, net)
		if err := ns.state.CancelDelegate(from, to, cpu, net); err != nil {
			return nil, err
		}
	default:
		return nil, errors.New(fmt.Sprintf("unknown method:%s", ns.method))
	}
	return nil, nil
}
