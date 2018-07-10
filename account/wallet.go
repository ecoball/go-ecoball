// Copyright 2018 The go-ecoball Authors
// This file is part of the go-ecoball.
//
// The go-ecoball is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ecoball is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ecoball. If not, see <http://www.gnu.org/licenses/>.
package account

import (
	"crypto/sha512"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/ecoball/go-ecoball/client/common"
	inner "github.com/ecoball/go-ecoball/common"
	"github.com/ecoball/go-ecoball/crypto/aes"
)

const (
	unlock byte = 0 //钱包未锁
	locked byte = 1 //钱包已锁
)

type KeyData struct {
	Checksum [64]byte                      `json:"Checksum"`
	Accounts map[inner.AccountName]Account `json:"Accounts"`
}

type WalletImpl struct {
	path string
	KeyData
	lockflag byte
}

var Wallet *WalletImpl //存储当前打开的钱包
var Cipherkeys []byte  //存储加密后的数据

/**
创建钱包
*/
func Create(path string, password []byte) error {
	//whether the wallet file exists
	if common.FileExisted(path) {
		return errors.New("The file already exists")
	}

	newWallet := &WalletImpl{
		path:     path,
		lockflag: unlock,
		KeyData: KeyData{
			Checksum: sha512.Sum512(password),
			Accounts: make(map[inner.AccountName]Account),
		},
	}

	//lock wallet
	cipherkeysTemp, err := newWallet.Lock(password)
	if nil != err {
		return err
	}

	//write data
	if err := newWallet.StoreWallet(cipherkeysTemp); nil != err {
		return err
	}

	//unlock wallet
	if err := newWallet.Unlock(password, cipherkeysTemp); nil != err {
		return err
	}

	return nil
}

/**
打开钱包
*/
func Open(path string, password []byte) (*WalletImpl, error) {
	newWallet := &WalletImpl{
		path:     path,
		lockflag: unlock,
		KeyData: KeyData{
			Accounts: make(map[inner.AccountName]Account),
		},
	}

	//load data
	cipherkeysTemp, err := newWallet.loadWallet()
	if nil != err {
		return nil, err
	}

	//unlock wallet
	if err := newWallet.Unlock(password, cipherkeysTemp); nil != err {
		return nil, err
	}

	return newWallet, nil
}

/**
关闭钱包
*/
func (wi *WalletImpl) Close(password []byte) error {
	//lock wallet
	cipherkeysTemp, err := wi.Lock(password)
	if nil != err {
		return err
	}

	//write data
	if err := wi.StoreWallet(cipherkeysTemp); nil != err {
		return err
	}

	return nil
}

/**
方法：内存数据存储到钱包文件中
*/
func (wi *WalletImpl) StoreWallet(data []byte) error {
	//open file
	file, err := os.OpenFile(wi.path, os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		return err
	}
	defer file.Close()

	//write data
	n, err := file.Write(data)
	if n != len(data) || err != nil {
		return err
	}

	return nil
}

/**
方法：将钱包文件的数据导入到内存中
*/
func (wi *WalletImpl) loadWallet() ([]byte, error) {
	//open file
	file, err := os.OpenFile(wi.path, os.O_RDONLY, 0666)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	//read data
	data, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}

	return data, nil
}

/**
方法：将密钥数据加密
*/
func (wi *WalletImpl) Lock(password []byte) ([]byte, error) {
	//whether the wallet is locked
	if wi.lockflag != unlock {
		return nil, errors.New("the wallet has been locked!!")
	}

	//whether the password is correct
	if (sha512.Sum512(password)) != wi.Checksum {
		return nil, errors.New("wrong password!!")
	}

	//marshal keyData
	data, err := json.Marshal(wi.KeyData)
	if nil != err {
		return nil, err
	}

	//encrypt data
	aesKey := wi.Checksum[0:32]
	iv := wi.Checksum[32:48]
	cipherkeyTemp, err := aes.AesEncrypt(data, aesKey, iv)
	if err != nil {
		return nil, err
	}

	//erase data
	for i := 0; i < len(wi.Checksum); i++ {
		wi.Checksum[i] = 0
	}
	for k, _ := range wi.Accounts {
		delete(wi.Accounts, k)
	}
	wi.lockflag = locked

	return cipherkeyTemp, nil
}

/**
方法：将密钥数据解密
*/
func (wi *WalletImpl) Unlock(password []byte, cipherkeysTemp []byte) error {
	//Decrypt data
	checksum := sha512.Sum512(password)
	aesKey := checksum[0:32]
	iv := checksum[32:48]
	aeskeys, err := aes.AesDecrypt(cipherkeysTemp, aesKey, iv)
	if nil != err {
		return err
	}

	//unmarshal data
	wallet := *wi
	if err := json.Unmarshal(aeskeys, &wi.KeyData); nil != err {
		*wi = wallet
		return err
	}

	//check password
	if wi.Checksum != checksum {
		*wi = wallet
		return errors.New("password error")
	}
	wi.lockflag = unlock

	return nil
}

/**
创建账号
*/
func (wi *WalletImpl) CreateAccount(password []byte, name string) (Account, error) {
	//create account
	ac, err := NewAccount(0)
	if err != nil {
		return Account{}, err
	}
	addr := inner.NameToIndex(name)
	wi.Accounts[addr] = ac

	//lock wallet
	cipherkeysTemp, err := wi.Lock(password)
	if nil != err {
		return Account{}, err
	}

	//write data
	if err := wi.StoreWallet(cipherkeysTemp); nil != err {
		return Account{}, err
	}

	//unlock wallet
	if err := wi.Unlock(password, cipherkeysTemp); nil != err {
		return Account{}, err
	}

	return ac, nil
}

/**
列出所有账号
*/
func (wi *WalletImpl) ListAccount() {
	for k, v := range wi.KeyData.Accounts {
		fmt.Println("account name: ", inner.IndexToName(k))
		fmt.Println("PrivateKey: ", inner.ToHex(v.PrivateKey[:]))
		fmt.Println("PublicKey: ", inner.ToHex(v.PublicKey[:]))
	}
}

/**
判断是否为锁定状态
**/
func (wi *WalletImpl) CheckLocked() bool {
	return wi.lockflag == locked
}
