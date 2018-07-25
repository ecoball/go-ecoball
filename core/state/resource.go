package state

import (
	"errors"
	"fmt"
	"github.com/ecoball/go-ecoball/common"
	"math/big"
)

var cpuAmount = []byte("cpu_amount")
var netAmount = []byte("net_amount")

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
		Staked    uint64  `json:"staked"`    //total stake delegated from account to self, uint ABA
		Delegated uint64  `json:"delegated"` //total stake delegated to account from others, uint ABA
		Used      float32 `json:"used"`      //uint Mib
		Available float32 `json:"available"` //uint Mib
		Limit     float32 `json:"limit"`     //uint Mib
	}
	Cpu struct {
		Staked    uint64  `json:"staked"`    //total stake delegated from account to self, uint ABA
		Delegated uint64  `json:"delegated"` //total stake delegated to account from others, uint ABA
		Used      float32 `json:"used"`      //uint ms
		Available float32 `json:"available"` //uint ms
		Limit     float32 `json:"limit"`     //uint ms
	}
}

type Delegate struct {
	Index common.AccountName `json:"index"`
	CpuStaked   uint64             `json:"cpu"`
	NetStaked   uint64             `json:"net"`
}

type BlockLimit struct {
	VirtualBlockCpuLimit uint64
	VirtualBlockNetLimit uint64
	BlockCpuLimit        uint64
	BlockNetLimit        uint64
}

func (s *State) SetResourceLimits(from, to common.AccountName, cpu, net uint64) error {
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
		return errors.New(fmt.Sprintf("the account:%s no enough balance", common.IndexToName(acc.Index)))
	}
	if err := acc.SubBalance(AbaToken, value); err != nil {
		return err
	}
	if err := s.AddResourceAmount(new(big.Int).SetUint64(uint64(cpu)), new(big.Int).SetUint64(uint64(net))); err != nil {
		return err
	}
	return s.CommitAccount(acc)
}
func (s *State) SubResourceLimits(index common.AccountName, cpu, net uint64) error {
	acc, err := s.GetAccountByName(index)
	if err != nil {
		return err
	}
	if err := acc.SubResourceLimits(cpu, net); err != nil {
		return err
	}
	return s.CommitAccount(acc)
}
func (s *State) CancelDelegate(from, to common.AccountName, cpu, net uint64) error {
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
		return errors.New(fmt.Sprintf("the cpu amount[%d] < 0", c))
	}
	data, err := value.GobEncode()
	if err != nil {
		return err
	}
	if err := s.trie.TryUpdate(cpuAmount, data); err != nil {
		return err
	}
	value = new(big.Int).Sub(n, net)
	if value.Sign() < 0 {
		return errors.New(fmt.Sprintf("the net amount[%d] < 0", n))
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

func (a *Account) SetResourceLimits(self bool, cpu, net uint64) error {
	if !self {
		if cpu != 0 {
			a.Cpu.Limit += float32(cpu)
			a.Cpu.Delegated += cpu
			a.Cpu.Available += float32(cpu)
		}
		if net != 0 {
			a.Net.Limit += float32(net)
			a.Net.Delegated += net
			a.Net.Available += float32(net)
		}
	} else {
		if cpu != 0 {
			a.Cpu.Staked += cpu
			a.Cpu.Limit += float32(cpu)
			a.Cpu.Available += float32(cpu)
		}
		if net != 0 {
			a.Net.Staked += net
			a.Net.Limit += float32(net)
			a.Net.Available += float32(net)
		}
	}
	return nil
}
func (a *Account) CancelDelegateSelf(cpu, net uint64) error {
	if cpu != 0 {
		a.Cpu.Staked -= cpu
		a.Cpu.Limit -= float32(cpu)
		a.Cpu.Available -= float32(cpu)
	}
	if net != 0 {
		a.Net.Staked -= net
		a.Net.Limit -= float32(net)
		a.Net.Available -= float32(net)
	}
	return nil
}
func (a *Account) CancelDelegateOther(acc *Account, cpu, net uint64) error {
	done := false
	for i := 0; i < len(a.Delegates); i++ {
		if a.Delegates[i].Index == acc.Index {
			done = true
			if acc.Cpu.Delegated < cpu {
				return errors.New(fmt.Sprintf("the account:%s cpu amount is not enough", common.IndexToName(acc.Index)))
			}
			if acc.Net.Delegated < net {
				return errors.New(fmt.Sprintf("the account:%s net amount is not enough", common.IndexToName(acc.Index)))
			}
			acc.Cpu.Limit -= float32(cpu)
			acc.Cpu.Delegated -= cpu
			acc.Cpu.Available = acc.Cpu.Limit - acc.Cpu.Used
			acc.Net.Limit -= float32(net)
			acc.Net.Delegated -= net
			acc.Net.Available = acc.Net.Limit - acc.Net.Used

			a.Cpu.Staked -= cpu
			a.Net.Staked -= net
			a.Delegates[i].CpuStaked -= cpu
			a.Delegates[i].NetStaked -= net
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
func (a *Account) SubResourceLimits(cpu, net uint64) error {
	if a.Cpu.Available < float32(cpu) {
		return errors.New(fmt.Sprintf("the account:%s cpu amount is not enough", common.IndexToName(a.Index)))
	}
	if a.Net.Available < float32(net) {
		return errors.New(fmt.Sprintf("the account:%s net amount is not enough", common.IndexToName(a.Index)))
	}
	a.Cpu.Available -= float32(cpu)
	a.Cpu.Used += float32(cpu)
	a.Net.Available -= float32(net)
	a.Net.Used += float32(net)
	return nil
}
func (a *Account) SetDelegateInfo(index common.AccountName, cpu, net uint64) error {
	d := Delegate{Index: index, CpuStaked: cpu, NetStaked: net}
	a.Delegates = append(a.Delegates, d)
	a.Cpu.Staked += cpu
	a.Net.Staked += net
	return nil
}
