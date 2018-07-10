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
	"crypto/sha256"
	"errors"

	"github.com/ecoball/go-ecoball/common"
	"github.com/ecoball/go-ecoball/crypto/secp256k1"
	"golang.org/x/crypto/ripemd160"
)

type Algorithm uint8 //算法类型

type Account struct {
	PublicKey  []byte    `json:"Publickey"`
	PrivateKey []byte    `json:"Privatekey"`
	Alg        Algorithm `json:"Alg"`
}

/**
创建账号
*/
func NewAccount(alg Algorithm) (Account, error) {
	pri, err := secp256k1.NewECDSAPrivateKey()
	if err != nil {
		return Account{}, errors.New("NewECDSAPrivateKey error: " + err.Error())
	}
	pridata, err := secp256k1.FromECDSA(pri)
	if err != nil {
		return Account{}, errors.New("FromECDSAPrivateKey error: " + err.Error())
	}
	pub, err := secp256k1.FromECDSAPub(&pri.PublicKey)
	if err != nil {
		return Account{}, errors.New("new account error: " + err.Error())
	}
	account := Account{
		PrivateKey: pridata,
		PublicKey:  pub,
		Alg:        alg,
	}
	return account, nil
}

/**
ECDSA算法签名
*/
func (s *Account) Sign(data []byte) ([]byte, error) {
	if s.PrivateKey == nil {
		return nil, errors.New("no private key")
	}
	signature, err := secp256k1.Sign(data, s.PrivateKey)
	if err != nil {
		return nil, err
	}
	return signature, nil
}

//ECDSA算法验签
func (s *Account) Verify(data []byte, signature []byte) (bool, error) {
	if s.PublicKey == nil {
		return false, errors.New("no public key")
	}
	return secp256k1.Verify(data, signature, s.PublicKey)
}

/**
公钥生成地址
*/
func AddressFromPubKey(pubkey []byte) common.Address {
	var addr common.Address
	temp := sha256.Sum256(pubkey)
	md := ripemd160.New()
	md.Write(temp[:])
	md.Sum(addr[:0])
	addr[0] = 0x01
	return addr
}
