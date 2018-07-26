package state

import (
	"errors"
	"fmt"
	"github.com/ecoball/go-ecoball/common"
	"math/big"
)

type Token struct {
	Name    string   `json:"index"`
	Balance *big.Int `json:"balance"`
}

func (s *State) AccountGetBalance(index common.AccountName, token string) (*big.Int, error) {
	acc, err := s.GetAccountByName(index)
	if err != nil {
		return nil, err
	}

	return acc.Balance(token)
}
func (s *State) AccountSubBalance(index common.AccountName, token string, value *big.Int) error {
	acc, err := s.GetAccountByName(index)
	if err != nil {
		return err
	}

	balance, err := acc.Balance(token)
	if err != nil {
		return err
	}
	if balance.Cmp(value) == -1 {
		return errors.New("no enough balance")
	}
	if err := acc.SubBalance(AbaToken, value); err != nil {
		return err
	}
	if err := s.CommitAccount(acc); err != nil {
		return err
	}
	return nil
}
func (s *State) AccountAddBalance(index common.AccountName, token string, value *big.Int) error {
	acc, err := s.GetAccountByName(index)
	if err != nil {
		return err
	}
	if err := acc.AddBalance(AbaToken, value); err != nil {
		return err
	}
	if err := s.CommitAccount(acc); err != nil {
		return err
	}

	return nil
}
func (s *State) CreateToken(token string, value *big.Int) error {
	//add token into trie
	data, err := value.GobEncode()
	if err != nil {
		return err
	}
	if err := s.trie.TryUpdate([]byte(token), data); err != nil {
		return err
	}
	return nil
}
func (s *State) GetToken(token string) (*big.Int, error) {
	if data, err := s.trie.TryGet([]byte(token)); err != nil {
		return nil, err
	} else {
		value := new(big.Int)
		if err := value.GobDecode(data); err != nil {
			return nil, err
		}
		return value, nil
	}
}
func (s *State) TokenExisted(name string) bool {
	data, err := s.trie.TryGet([]byte(name))
	if err != nil {
		log.Error(err)
		return false
	}
	return string(data) == name
}

/**
 *  @brief create a new token in account
 *  @param index - the unique id of token name created by common.NameToIndex()
 */
func (a *Account) AddToken(name string) error {
	log.Info("add token:", name)
	ac := Token{Name: name, Balance: new(big.Int).SetUint64(0)}
	a.Tokens[name] = ac
	return nil
}

/**
 *  @brief check the token for existence, return true if existed
 *  @param index - the unique id of token name created by common.NameToIndex()
 */
func (a *Account) TokenExisted(token string) bool {
	_, ok := a.Tokens[token]
	if ok {
		return true
	}
	return false
}

/**
 *  @brief add balance into account
 *  @param index - the unique id of token name created by common.NameToIndex()
 *  @param amount - value of token
 */
func (a *Account) AddBalance(name string, amount *big.Int) error {
	log.Info("add token", name, "balance:", amount, a.Index)
	if amount.Sign() == 0 {
		return errors.New("amount is zero")
	}
	ac, ok := a.Tokens[name]
	if !ok {
		if err := a.AddToken(name); err != nil {
			return err
		}
		ac, _ = a.Tokens[name]
	}
	ac.SetBalance(new(big.Int).Add(ac.GetBalance(), amount))
	a.Tokens[name] = ac
	return nil
}

/**
 *  @brief sub balance into account
 *  @param index - the unique id of token name created by common.NameToIndex()
 *  @param amount - value of token
 */
func (a *Account) SubBalance(token string, amount *big.Int) error {
	if amount.Sign() == 0 {
		return errors.New("amount is zero")
	}
	t, ok := a.Tokens[token]
	if !ok {
		return errors.New("not sufficient funds")
	}
	balance := t.GetBalance()
	value := new(big.Int).Sub(balance, amount)
	if value.Sign() < 0 {
		return errors.New("the balance is not enough")
	}
	t.SetBalance(value)
	a.Tokens[token] = t
	return nil
}

/**
 *  @brief get the balance of account
 *  @param index - the unique id of token name created by common.NameToIndex()
 *  @return big.int - value of token
 */
func (a *Account) Balance(token string) (*big.Int, error) {
	t, ok := a.Tokens[token]
	if !ok {
		return nil, errors.New(fmt.Sprintf("can't find token account:%s, in account:%s", token, common.IndexToName(a.Index)))
	}
	return t.GetBalance(), nil
}

/**
 *  @brief set balance of account
 *  @param amount - value of token
 */
func (t *Token) SetBalance(amount *big.Int) {
	//TODO:将变动记录存到日志文件
	t.setBalance(amount)
}
func (t *Token) setBalance(amount *big.Int) {
	t.Balance = amount
}
func (t *Token) GetBalance() *big.Int {
	return t.Balance
}
