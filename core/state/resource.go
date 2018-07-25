package state

import (
	"errors"
	"fmt"
	"github.com/ecoball/go-ecoball/common"
	"math/big"
)

var cpuAmount = "cpu_amount"
var netAmount = "net_amount"

//var VirtualBlockCpuLimit = 200000000
//var VirtualBlockNetLimit = 1048576000
const BlockCpuLimit = 200000
const BlockNetLimit = 1048576

type Resource struct {
	Ram struct {
		Quota float32 `json:"quota"`
		Used  float32 `json:"used"`
	}
	Net struct {
		Staked    uint64  `json:"staked_aba"`    //total stake delegated from account to self, uint ABA
		Delegated uint64  `json:"delegated_aba"` //total stake delegated to account from others, uint ABA
		Used      float32 `json:"used_kib"`      //uint Kib
		Available float32 `json:"available_kib"` //uint Kib
		Limit     float32 `json:"limit_kib"`     //uint Kib
	}
	Cpu struct {
		Staked    uint64  `json:"staked_aba"`    //total stake delegated from account to self, uint ABA
		Delegated uint64  `json:"delegated_aba"` //total stake delegated to account from others, uint ABA
		Used      float32 `json:"used_ms"`       //uint ms
		Available float32 `json:"available_ms"`  //uint ms
		Limit     float32 `json:"limit_ms"`      //uint ms
	}
}

type Delegate struct {
	Index     common.AccountName `json:"index"`
	CpuStaked uint64             `json:"cpu_aba"`
	NetStaked uint64             `json:"net_aba"`
}

type BlockLimit struct {
	VirtualBlockCpuLimit uint64
	VirtualBlockNetLimit uint64
	BlockCpuLimit        uint64
	BlockNetLimit        uint64
}

func (s *State) SetResourceLimits(from, to common.AccountName, cpuStaked, netStaked uint64) error {
	cpuStakedSum, err := s.GetParam(cpuAmount)
	if err != nil {
		return err
	}
	netStakedSum, err := s.GetParam(netAmount)
	if err != nil {
		return err
	}
	acc, err := s.GetAccountByName(from)
	if err != nil {
		return err
	}
	if from == to {
		if err := acc.SetResourceLimits(true, cpuStaked, netStaked, cpuStakedSum, netStakedSum); err != nil {
			return err
		}
	} else {
		if err := acc.SetDelegateInfo(to, cpuStaked, netStaked); err != nil {
			return err
		}
		accTo, err := s.GetAccountByName(to)
		if err != nil {
			return err
		}
		if err := accTo.SetResourceLimits(false, cpuStaked, netStaked, cpuStakedSum, netStakedSum); err != nil {
			return err
		}
		if err := s.CommitAccount(accTo); err != nil {
			return err
		}
	}

	value := new(big.Int).Add(new(big.Int).SetUint64(uint64(cpuStaked)), new(big.Int).SetUint64(uint64(netStaked)))
	if err := acc.SubBalance(AbaToken, value); err != nil {
		return err
	}
	amount, err := s.GetParam(cpuAmount)
	if err != nil {
		return err
	}
	if err := s.CommitParam(cpuAmount, cpuStaked+amount); err != nil {
		return err
	}
	amount, err = s.GetParam(netAmount)
	if err != nil {
		return err
	}
	if err := s.CommitParam(netAmount, netStaked+amount); err != nil {
		return err
	}

	return s.CommitAccount(acc)
}
func (s *State) SubResourceLimits(index common.AccountName, cpu, net float32) error {
	cpuStakedSum, err := s.GetParam(cpuAmount)
	if err != nil {
		return err
	}
	netStakedSum, err := s.GetParam(netAmount)
	if err != nil {
		return err
	}
	acc, err := s.GetAccountByName(index)
	if err != nil {
		return err
	}
	if err := acc.SubResourceLimits(cpu, net, cpuStakedSum, netStakedSum); err != nil {
		return err
	}
	return s.CommitAccount(acc)
}
func (s *State) CancelDelegate(from, to common.AccountName, cpuStaked, netStaked uint64) error {
	cpuStakedSum, err := s.GetParam(cpuAmount)
	if err != nil {
		return err
	}
	netStakedSum, err := s.GetParam(netAmount)
	if err != nil {
		return err
	}
	acc, err := s.GetAccountByName(from)
	if err != nil {
		return err
	}
	if from != to {
		accTo, err := s.GetAccountByName(to)
		if err != nil {
			return err
		}
		if err := acc.CancelDelegateOther(accTo, cpuStaked, netStaked, cpuStakedSum, netStakedSum); err != nil {
			return err
		}
		if err := s.CommitAccount(accTo); err != nil {
			return err
		}
	} else {
		if err := acc.CancelDelegateSelf(cpuStaked, netStaked, cpuStakedSum, netStakedSum); err != nil {
			return err
		}
	}
	value := new(big.Int).Add(new(big.Int).SetUint64(uint64(cpuStaked)), new(big.Int).SetUint64(uint64(netStaked)))
	if err := acc.AddBalance(AbaToken, value); err != nil {
		return err
	}
	amount, err := s.GetParam(cpuAmount)
	if err != nil {
		return err
	}
	if err := s.CommitParam(cpuAmount, amount-cpuStaked); err != nil {
		return err
	}
	amount, err = s.GetParam(netAmount)
	if err != nil {
		return err
	}
	if err := s.CommitParam(netAmount, amount-cpuStaked); err != nil {
		return err
	}

	return s.CommitAccount(acc)
}

func (a *Account) SetResourceLimits(self bool, cpuStaked, netStaked, cpuStakedSum, netStakedSum uint64) error {
	if self {
		a.Cpu.Staked += cpuStaked
		a.Net.Staked += netStaked
	} else {
		a.Cpu.Delegated += cpuStaked
		a.Net.Delegated += netStaked
	}
	a.Cpu.Limit = float32(a.Cpu.Staked + a.Cpu.Delegated) / float32(cpuStakedSum) * BlockCpuLimit
	a.Cpu.Available = a.Cpu.Limit - a.Cpu.Used
	a.Net.Limit = float32(a.Cpu.Staked + a.Net.Delegated) / float32(netStakedSum) * BlockNetLimit
	a.Net.Available = a.Net.Limit - a.Net.Used

	return nil
}
func (a *Account) CancelDelegateSelf(cpuStaked, netStaked, cpuStakedSum, netStakedSum uint64) error {
	a.Cpu.Staked -= cpuStaked
	a.Cpu.Limit = float32(a.Cpu.Staked + a.Cpu.Delegated) / float32(cpuStakedSum) * BlockCpuLimit
	a.Cpu.Available = a.Cpu.Limit - a.Cpu.Used

	a.Net.Staked -= netStaked
	a.Net.Limit = float32(a.Net.Staked + a.Net.Delegated) / float32(netStakedSum) * BlockNetLimit
	a.Net.Available = a.Net.Limit - a.Net.Used

	return nil
}
func (a *Account) CancelDelegateOther(acc *Account, cpuStaked, netStaked, cpuStakedSum, netStakedSum uint64) error {
	done := false
	for i := 0; i < len(a.Delegates); i++ {
		if a.Delegates[i].Index == acc.Index {
			done = true
			if acc.Cpu.Delegated < cpuStaked {
				return errors.New(fmt.Sprintf("the account:%s cpu amount is not enough", common.IndexToName(acc.Index)))
			}
			if acc.Net.Delegated < netStaked {
				return errors.New(fmt.Sprintf("the account:%s net amount is not enough", common.IndexToName(acc.Index)))
			}
			acc.Cpu.Delegated -= cpuStaked
			acc.Cpu.Limit = float32(acc.Cpu.Staked + acc.Cpu.Delegated) / float32(cpuStakedSum) * BlockCpuLimit
			acc.Cpu.Available = acc.Cpu.Limit - acc.Cpu.Used

			acc.Net.Delegated -= netStaked
			acc.Net.Limit = float32(acc.Net.Staked + acc.Net.Delegated) / float32(netStakedSum) * BlockNetLimit
			acc.Net.Available = acc.Net.Limit - acc.Net.Used

			a.Delegates[i].CpuStaked -= cpuStaked
			a.Delegates[i].NetStaked -= netStaked
			if a.Delegates[i].CpuStaked == 0 && a.Delegates[i].NetStaked == 0 {
				a.Delegates = append(a.Delegates[:i], a.Delegates[i+1:]...)
			}
		}
	}
	if done == false {
		return errors.New(fmt.Sprintf("account:%s is not delegated for %s", common.IndexToName(a.Index), common.IndexToName(acc.Index)))
	}
	return nil
}
func (a *Account) SubResourceLimits(cpu, net float32, cpuStakedSum, netStakedSum uint64) error {
	if a.Cpu.Available < cpu {
		return errors.New(fmt.Sprintf("the account:%s cpu amount is not enough", common.IndexToName(a.Index)))
	}
	if a.Net.Available < net {
		return errors.New(fmt.Sprintf("the account:%s net amount is not enough", common.IndexToName(a.Index)))
	}
	a.Cpu.Used += cpu
	a.Net.Used += net
	a.Cpu.Limit = float32(a.Cpu.Staked + a.Cpu.Delegated) / float32(cpuStakedSum) * BlockCpuLimit
	a.Cpu.Available = a.Cpu.Limit - a.Cpu.Used
	a.Net.Limit = float32(a.Net.Staked + a.Net.Delegated) / float32(netStakedSum) * BlockNetLimit
	a.Net.Available = a.Net.Limit - a.Net.Used
	return nil
}
func (a *Account) SetDelegateInfo(index common.AccountName, cpuStaked, netStaked uint64) error {
	d := Delegate{Index: index, CpuStaked: cpuStaked, NetStaked: netStaked}
	a.Delegates = append(a.Delegates, d)
	return nil
}
