package state

import (
	"math/big"
	"github.com/ecoball/go-ecoball/common"
	"errors"
	"fmt"
)

var cpuAmount = []byte("cpu_amount")
var netAmount = []byte("net_amount")
//var VirtualBlockCpuLimit = 200000000
//var VirtualBlockNetLimit = 1048576000
var BlockCpuLimit = 200000
var BlockNetLimit = 1048576

type BlockLimit struct {
	VirtualBlockCpuLimit uint64
	VirtualBlockNetLimit uint64
	BlockCpuLimit uint64
	BlockNetLimit uint64
}

func (s *State) SetResourceLimits(from, to common.AccountName, cpu, net float32) error {
	acc, err := s.GetAccountByName(from)
	if err != nil {
		return err
	}
	if from == to {
		if err := acc.SetResourceLimits(true, cpu, net); err != nil {
			return err
		}
	} else {
		if err := acc.SetDelegateInfo(to, cpu, net); err != nil {
			return err
		}
		accTo, err := s.GetAccountByName(to)
		if err != nil {
			return err
		}
		if err := accTo.SetResourceLimits(false, cpu, net); err != nil {
			return err
		}
		if err := s.CommitAccount(accTo); err != nil {
			return err
		}
	}
	balance, err := acc.Balance(AbaToken)
	if err != nil {
		return err
	}
	value := new(big.Int).Add(new(big.Int).SetUint64(uint64(cpu)), new(big.Int).SetUint64(uint64(net)))
	if balance.Cmp(value) == -1 {
		return errors.New("no enough balance")
	}
	if err := acc.SubBalance(AbaToken, value); err != nil {
		return err
	}
	if err := s.AddResourceAmount(new(big.Int).SetUint64(uint64(cpu)), new(big.Int).SetUint64(uint64(net))); err != nil {
		return err
	}
	return s.CommitAccount(acc)
}
func (s *State) SubResourceLimits(index common.AccountName, cpu, net float32) error {
	acc, err := s.GetAccountByName(index)
	if err != nil {
		return err
	}
	if err := acc.SubResourceLimits(cpu, net); err != nil {
		return err
	}
	return s.CommitAccount(acc)
}
func (s *State) CancelDelegate(from, to common.AccountName, cpu, net float32) error {
	acc, err := s.GetAccountByName(from)
	if err != nil {
		return err
	}
	if from != to {
		accTo, err := s.GetAccountByName(to)
		if err != nil {
			return err
		}
		if err := acc.CancelDelegateOther(accTo, cpu, net); err != nil {
			return err
		}
		if err := s.CommitAccount(accTo); err != nil {
			return err
		}
	} else {
		if err := acc.CancelDelegateSelf(cpu, net); err != nil {
			return err
		}
	}
	value := new(big.Int).Add(new(big.Int).SetUint64(uint64(cpu)), new(big.Int).SetUint64(uint64(net)))
	if err := acc.AddBalance(AbaToken, value); err != nil {
		return err
	}
	if err := s.SubResourceAmount(new(big.Int).SetUint64(uint64(cpu)), new(big.Int).SetUint64(uint64(net))); err != nil {
		return err
	}
	return s.CommitAccount(acc)
}
func (s *State) AddResourceAmount(cpu, net *big.Int) error {
	c, n, err := s.GetResourceAmount()
	if err != nil {
		return err
	}
	value := new(big.Int).Add(cpu, c)
	data, err := value.GobEncode()
	if err != nil {
		return err
	}
	if err := s.trie.TryUpdate(cpuAmount, data); err != nil {
		return err
	}
	log.Debug("cpu amount:", value)
	value = new(big.Int).Add(net, n)
	data, err = value.GobEncode()
	if err != nil {
		return err
	}
	if err := s.trie.TryUpdate(netAmount, data); err != nil {
		return err
	}
	log.Debug("net amount:", value)
	return nil
}
func (s *State) SubResourceAmount(cpu, net *big.Int) error {
	c, n, err := s.GetResourceAmount()
	if err != nil {
		return err
	}
	value := new(big.Int).Sub(c, cpu)
	if value.Sign() < 0 {
		return errors.New("the cpu amount < 0")
	}
	data, err := value.GobEncode()
	if err != nil {
		return err
	}
	if err := s.trie.TryUpdate(cpuAmount, data); err != nil {
		return err
	}
	value = new(big.Int).Sub(net, n)
	if value.Sign() < 0 {
		return errors.New("the net amount < 0")
	}
	data, err = value.GobEncode()
	if err != nil {
		return err
	}
	if err := s.trie.TryUpdate(netAmount, data); err != nil {
		return err
	}
	return nil
}
func (s *State) GetResourceAmount() (*big.Int, *big.Int, error) {
	data, _ := s.trie.TryGet(cpuAmount)
	cpu := new(big.Int)
	if err := cpu.GobDecode(data); err != nil {
		return nil, nil, err
	}
	data, _ = s.trie.TryGet(netAmount)
	net := new(big.Int)
	if err := net.GobDecode(data); err != nil {
		return nil, nil, err
	}
	return cpu, net, nil
}

func (a *Account) SetResourceLimits(self bool, cpu, net float32) error {
	if !self {
		if cpu != 0 {
			a.Cpu.Limit += cpu
			a.Cpu.Delegated += cpu
			a.Cpu.Available += cpu
		}
		if net != 0 {
			a.Net.Limit += net
			a.Net.Delegated += net
			a.Net.Available += net
		}
	} else {
		if cpu != 0 {
			a.Cpu.Staked += cpu
			a.Cpu.Limit += cpu
			a.Cpu.Available += cpu
		}
		if net != 0 {
			a.Net.Staked += net
			a.Net.Limit += net
			a.Net.Available += net
		}
	}
	return nil
}
func (a *Account) CancelDelegateSelf(cpu, net float32) error {
	if cpu != 0 {
		a.Cpu.Staked -= cpu
		a.Cpu.Limit -= cpu
		a.Cpu.Available -= cpu
	}
	if net != 0 {
		a.Net.Staked -= net
		a.Net.Limit -= net
		a.Net.Available -= net
	}
	return nil
}
func (a *Account) CancelDelegateOther(acc *Account, cpu, net float32) error {
	done := false
	for i := 0; i < len(a.Delegates); i++ {
		if a.Delegates[i].Index == acc.Index {
			done = true
			if acc.Cpu.Delegated < cpu {
				return errors.New("cpu amount is not enough")
			}
			if acc.Net.Delegated < net {
				return errors.New("net amount is not enough")
			}
			acc.Cpu.Limit -= cpu
			acc.Cpu.Delegated -= cpu
			acc.Cpu.Available = acc.Cpu.Limit - acc.Cpu.Used
			acc.Net.Limit -= net
			acc.Net.Delegated -= net
			acc.Net.Available = acc.Net.Limit - acc.Net.Used

			a.Cpu.Staked -= cpu
			a.Net.Staked -= net
			a.Delegates[i].Cpu -= cpu
			a.Delegates[i].Net -= net
			if a.Delegates[i].Cpu == 0 && a.Delegates[i].Net == 0 {
				a.Delegates = append(a.Delegates[:i], a.Delegates[i+1:]...)
			}
		}
	}
	if done == false {
		return errors.New(fmt.Sprintf("account:%s is not delegated for %s", common.IndexToName(a.Index), common.IndexToName(acc.Index)))
	}
	return nil
}
func (a *Account) SubResourceLimits(cpu, net float32) error {
	if a.Cpu.Available < cpu {
		return errors.New("cpu is not enough")
	}
	if a.Net.Available < net {
		return errors.New("net is not enough")
	}
	a.Cpu.Available -= cpu
	a.Cpu.Used += cpu
	a.Net.Available -= net
	a.Net.Used += net
	return nil
}
func (a *Account) SetDelegateInfo(index common.AccountName, cpu, net float32) error {
	d := Delegate{Index: index, Cpu: cpu, Net: net}
	a.Delegates = append(a.Delegates, d)
	a.Cpu.Staked += cpu
	a.Net.Staked += net
	return nil
}

func (a *Account) PledgeCpu(token string, value *big.Int) error {
	if err := a.SubBalance(token, value); err != nil {
		return err
	}
	a.Cpu.Available = 100
	a.Cpu.Limit = 100
	a.Cpu.Used = 0
	return nil
}
func (a *Account) CancelPledgeCpu(token string, value *big.Int) error {
	if err := a.AddBalance(token, value); err != nil {
		return err
	}
	a.Cpu.Available = 0
	a.Cpu.Limit = 100
	a.Cpu.Used = 100
	return nil
}