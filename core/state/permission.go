package state

import (
	"errors"
	"fmt"
	"github.com/ecoball/go-ecoball/common"
	"encoding/json"
)

var Owner = "owner"
var Active = "active"

type AccFactor struct {
	Actor      common.AccountName `json:"actor"`
	Weight     uint32             `json:"weight"`
	Permission string             `json:"permission"`
}

type KeyFactor struct {
	Actor  common.Address `json:"actor"`
	Weight uint32         `json:"weight"`
}

type Permission struct {
	PermName  string               `json:"perm_name"`
	Parent    string               `json:"parent"`
	Threshold uint32               `json:"threshold"`
	Keys      map[string]KeyFactor `json:"keys, omitempty"`
	Accounts  map[string]AccFactor `json:"accounts, omitempty"`
}

/**
 *  @brief create a new permission object
 *  @param name - the permission's name
 *  @param parent - the parent name of this permission, if the permission's name is 'owner', then the parent is null
 *  @param threshold - the threshold of this permission, when the weight greater than or equal to threshold, permission will only take effect
 *  @param addr - the public keys list
 *  @param acc - the accounts list
 */
func NewPermission(name, parent string, threshold uint32, addr []KeyFactor, acc []AccFactor) Permission {
	Keys := make(map[string]KeyFactor, 1)
	for _, a := range addr {
		Keys[a.Actor.HexString()] = a
	}
	Accounts := make(map[string]AccFactor, 1)
	for _, a := range acc {
		Accounts[a.Actor.String()] = a
	}
	return Permission{
		PermName:  name,
		Parent:    parent,
		Threshold: threshold,
		Keys:      Keys,
		Accounts:  Accounts,
	}
}

/**
 *  @brief check that the signatures meets the permission requirement
 *  @param state - the mpt trie, used to search account
 *  @param signatures - the transaction's signatures list
 */
func (p *Permission) CheckPermission(state *State, signatures []common.Signature) error {
	Keys := make(map[common.Address][]byte, 1)
	Accounts := make(map[string][]byte, 1)
	for _, s := range signatures {
		addr := common.AddressFromPubKey(s.PubKey)
		acc, err := state.GetAccountByAddr(addr)
		if err == nil {
			Accounts[acc.Index.String()] = s.SigData
		} else {
			log.Warn("permission", p.PermName, "error:", err) //allow mixed with invalid account, just have enough weight
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
			if next, err := state.GetAccountByName(a.Actor); err != nil {
				return err
			} else {
				perm := next.Permissions[a.Permission]
				if err := perm.CheckPermission(state, signatures); err != nil {
					return err
				}
			}
		}
		if weightAcc >= p.Threshold {
			return nil
		}
	}

	return errors.New(fmt.Sprintf("weight is not enough, keys weight:%d, accounts weight:%d", weightKey, weightAcc))
}

/**
 *  @brief add a permission object into account, then update to mpt trie
 *  @param perm - the permission object
 */
func (s *State) AddPermission(index common.AccountName, perm Permission) error {
	acc, err := s.GetAccountByName(index)
	if err != nil {
		return err
	}
	acc.AddPermission(perm)
	return s.CommitAccount(acc)
}

/**
 *  @brief check the permission's validity, this method will not modified mpt trie
 *  @param index - the account index
 *  @param state - the world state tree
 *  @param name - the permission names
 *  @param signatures - the signatures list
 */
func (s *State) CheckPermission(index common.AccountName, name string, signatures []common.Signature) error {
	acc, err := s.GetAccountByName(index)
	if err != nil {
		return err
	}
	return acc.CheckPermission(s, name, signatures)
}

/**
 *  @brief search the permission by name, return json array string
 *  @param index - the account index
 *  @param name - the permission names
 */
func (s *State) FindPermission(index common.AccountName, name string) (string, error) {
	acc, err := s.GetAccountByName(index)
	if err != nil {
		return "", err
	}
	if str, err := acc.FindPermission(name); err != nil {
		return "", err
	} else {
		return "[" + str + "]", nil
	}
}

/**
 *  @brief set the permission into account, if the permission existed, will be to overwrite
 *  @param name - the permission name
 */
func (a *Account) AddPermission(perm Permission) {
	a.Permissions[perm.PermName] = perm
}

/**
 *  @brief check that the signatures meets the permission requirement
 *  @param state - the mpt trie, used to search account
 *  @param name - the permission name
 *  @param signatures - the transaction's signatures list
 */
func (a *Account) CheckPermission(state *State, name string, signatures []common.Signature) error {
	if perm, ok := a.Permissions[name]; !ok {
		return errors.New(fmt.Sprintf("can't find this permission in account:%s", name))
	} else {
		if "" != perm.Parent {
			if err := a.CheckPermission(state, perm.Parent, signatures); err == nil {
				return nil
			}
		}
		if err := perm.CheckPermission(state, signatures); err != nil {
			return errors.New(fmt.Sprintf("account:%s %s", common.IndexToName(a.Index), err.Error()))
		}
	}
	return nil
}

/**
 *  @brief get the permission information by name, return json string
 *  @param name - the permission name
 */
func (a *Account) FindPermission(name string) (str string, err error) {
	perm, ok := a.Permissions[name]
	if !ok {
		return "", errors.New(fmt.Sprintf("can't find this permission:%s", name))
	}
	b, err := json.Marshal(perm)
	if err != nil {
		return "", err
	}
	str += string(b)
	if "" != perm.Parent {
		if s, err := a.FindPermission(perm.Parent); err == nil {
			str += "," + s
		}
	}
	return string(str), nil
}
