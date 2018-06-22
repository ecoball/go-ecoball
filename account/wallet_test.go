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
	"testing"

	"fmt"
	"crypto/sha256"
)

func TestSign(t *testing.T) {

	path := "test.dat"
	password := []byte{1,2,3,4}
	txn := []byte{1,2,3,4,5,67,8,9,0,5}

	txnhash := sha256.Sum256(txn)
	var wallet *WalletImpl

	wallet = Create(path, password)
	wallet = Open(path, password)
	acc,_ := wallet.CreateAccount(password)
	wallet.ListAccount()
	address := AddressFromPubKey(acc.PublicKey)
	addr := address.ToBase58()

	signed,_:= acc.Sign(txnhash[:])
	Close(path, password)

	wallet = Open(path, password)
    account := wallet.Accounts[addr]

	ret,err :=account.Verify(txnhash[:],signed)
	if ret != true{
		fmt.Printf("verify err\n")
	}
    if err != nil{
    	fmt.Printf("err:%s\n",err.Error())
	}

}
