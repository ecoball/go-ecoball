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
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"sync"

	"github.com/ecoball/go-ecoball/client/common"
	inner "github.com/ecoball/go-ecoball/common"
	"github.com/ecoball/go-ecoball/crypto/aes"
)

var cipherkeys []byte //需要存储的钱包文件的密钥数据

const (
	unlock byte = 0 //钱包未锁
	locked byte = 1 //钱包已锁
)

type KeyData struct {
	Checksum [64]byte                      `json:"Checksum"`
	Accounts map[inner.AccountName]Account `json:"Accounts"`
}

type WalletImpl struct {
	mu   sync.Mutex
	path string
	KeyData
	lockflag byte
}

/**
创建钱包
*/
func Create(path string, password []byte) *WalletImpl {

	if common.FileExisted(path) {
		fmt.Println("The file already exists")
		return nil
	}
	newWallet := &WalletImpl{
		path:     path,
		lockflag: unlock,
		KeyData: KeyData{
			Checksum: sha512.Sum512(password),
			Accounts: make(map[inner.AccountName]Account),
		},
	}
	cipherkeys = newWallet.lock(password)
	newWallet.storeWallet(cipherkeys)

	newWallet.unlock(password)
	return newWallet
}

/**
打开钱包
*/
func Open(path string, password []byte) *WalletImpl {
	newWallet := &WalletImpl{
		path:     path,
		lockflag: unlock,
		KeyData: KeyData{
			Accounts: make(map[inner.AccountName]Account),
		},
	}
	cipherkeys = newWallet.loadWallet()
	if err := newWallet.unlock(password); nil != err {
		return nil
	}
	return newWallet
}

/**
关闭钱包
*/
func Close(path string, password []byte) error {
	newWallet := &WalletImpl{
		path:     path,
		lockflag: unlock,
	}
	newWallet.lock(password)
	newWallet.storeWallet(cipherkeys)
	return nil
}

/**
方法：内存数据存储到钱包文件中
*/
func (wi *WalletImpl) storeWallet(data []byte) error {
	var err error
	var file *os.File
	file, err = os.OpenFile(wi.path, os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		return err
		log.Printf("write file err")
	}
	defer file.Close()
	n, err := file.Write(data)
	if n != len(data) || err != nil {
		log.Printf("write err")
		return err
	}
	return nil
}

/**
方法：将钱包文件的数据导入到内存中
*/
func (wi *WalletImpl) loadWallet() []byte {
	var err error
	var file *os.File

	file, err = os.OpenFile(wi.path, os.O_RDONLY, 0666)
	if err != nil {
		log.Println("open file fail")
	}
	defer file.Close()
	data, err := ioutil.ReadAll(file)
	if err != nil {
		log.Printf("read err:%s\n", err.Error())
	}
	return data
}

/**
方法：将密钥数据加密
*/
func (wi *WalletImpl) lock(password []byte) []byte {
	var err error

	if wi.lockflag != unlock {
		return nil
	}
	aesKey := wi.Checksum[0:32]
	iv := wi.Checksum[32:48]
	data, _ := json.Marshal(wi.KeyData)
	cipherkey, err := aes.AesEncrypt(data, aesKey, iv)
	if err != nil {
		return nil
	}
	for i := 0; i < len(wi.Checksum); i++ {
		wi.Checksum[i] = 0
	}
	for k, _ := range wi.Accounts {
		delete(wi.Accounts, k)
	}
	wi.lockflag = locked
	return cipherkey
}

/**
方法：将密钥数据解密
*/
func (wi *WalletImpl) unlock(password []byte) error {
	var err error

	checksum := sha512.Sum512(password)
	aesKey := checksum[0:32]
	iv := checksum[32:48]
	aeskeys, err := aes.AesDecrypt(cipherkeys, aesKey, iv)
	err = json.Unmarshal(aeskeys, &wi.KeyData)
	if err != nil {
		log.Println("unmarshal error")
		return err
	}
	if wi.Checksum != checksum {
		log.Println("password error")
		return err
	}
	wi.lockflag = unlock
	return nil
}

/**
创建账号
*/
func (wi *WalletImpl) CreateAccount(password []byte, name string) (Account, error) {
	ac, err := NewAccount(0)
	if err != nil {
		return Account{}, err
	}
	addr := inner.NameToIndex(name)
	wi.mu.Lock()
	wi.Accounts[addr] = ac
	wi.mu.Unlock()
	cipherkeys = wi.lock(password)
	wi.storeWallet(cipherkeys)
	wi.unlock(password)
	return ac, nil
}

/**
列出所有账号
*/
func (wi *WalletImpl) ListAccount() {
	data, _ := json.MarshalIndent(wi.KeyData, "", "	")
	fmt.Printf("data:%s\n", data)
}
