package state

import (
	"errors"
	"fmt"
	"github.com/ecoball/go-ecoball/common"
)

type account struct {
	Actor      common.AccountName `json:"actor"`
	Weight     uint32             `json:"weight"`
	Permission string             `json:"permission"`
}

type address struct {
	Actor  common.Address `json:"actor"`
	Weight uint32         `json:"weight"`
}

type Permission struct {
	PermName  string             `json:"perm_name"`
	Parent    string             `json:"parent"`
	Threshold uint32             `json:"threshold"`
	Keys      map[string]address `json:"keys, omitempty"`
	Accounts  map[string]account `json:"accounts, omitempty"`
}

func (p *Permission) CheckPermission(state *State, signatures []common.Signature) error {
	Keys := make(map[common.Address][]byte, 1)
	Accounts := make(map[string][]byte, 1)
	for _, s := range signatures {
		addr := common.AddressFromPubKey(s.PubKey)
		acc, err := state.GetAccountByAddr(addr)
		if err == nil {
			Accounts[acc.Index.String()] = s.SigData
		} else {
			log.Warn("permission", p.PermName, "error:", err)
		}
		Keys[addr] = s.SigData
	}
	var weightKey uint32
	for addr := range Keys {
		if key, ok := p.Keys[addr.HexString()]; ok {
			weightKey += key.Weight
		}
		if weightKey >= p.Threshold {
			return nil
		}
	}
	var weightAcc uint32
	for acc := range Accounts {
		if a, ok := p.Accounts[acc]; ok {
			weightAcc += a.Weight
		}
		if weightAcc >= p.Threshold {
			return nil
		}
	}

	return errors.New(fmt.Sprintf("weight is not enough, keys weight:%d, accounts weight:%d", weightKey, weightAcc))
}
