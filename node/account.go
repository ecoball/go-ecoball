// Copyright 2018 The go-ecoball Authors
// This file is part of the go-ecoball library.
//
// The go-ecoball library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ecoball library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ecoball library. If not, see <http://www.gnu.org/licenses/>.

package main

import (
	"github.com/urfave/cli"
)

var (
	Name     string
	Password string
	Address  string
)

func NewNameFlag() cli.Flag {
	return cli.StringFlag{
		Name:        "name",
		Usage:       "wallet file name",
		Value:       "",
		Destination: &Name,
	}
}

func NewPasswordFlag() cli.Flag {
	return cli.StringFlag{
		Name:        "password",
		Usage:       "wallet password",
		Value:       "",
		Destination: &Password,
	}
}

func NewAddressFlag() cli.Flag {
	return cli.StringFlag{
		Name:        "address",
		Usage:       "account address",
		Value:       "",
		Destination: &Password,
	}
}

func checkPassword() {
	/*var (
			nameTime     = 0
			passwordTime = 0
		)

		//empty name
		if "" == Name {
			fmt.Printf("please input wallet file name:")
			fmt.Scan(&Name)
			goto name
		}

	name:
		if 0 != nameTime {
			fmt.Printf("please input wallet file name:")
			fmt.Scan(&Name)
		}

		//file does not exist
		if _, err := os.Stat(Name); err != nil {
			fmt.Fprintf(os.Stderr, "%v", err)
			nameTime++
			if nameTime >= 3 {
				fmt.Fprintln(os.Stderr, "More than three times, maybe you didn't create your wallet, exit...")
				os.Exit(1)
			}
			goto name
		}

		//empty password
		if "" == Password {
			fmt.Printf("please input wallet password:")
			fmt.Scan(&Password)
			goto password
		}

	password:
		if 0 != passwordTime {
			fmt.Printf("please input wallet password:")
			fmt.Scan(&Password)
		}

		//worng password
		wallet := account.Open(Name, []byte(Password))
		if nil == wallet {
			fmt.Fprintln(os.Stderr, "open wallet failed!")
			passwordTime++
			if passwordTime >= 3 {
				fmt.Fprintln(os.Stderr, "More than three times, exit...")
				os.Exit(1)
			}
			goto password
		}

		//get account
		if 0 == len(wallet.KeyData.Accounts) {
			fmt.Fprintln(os.Stderr, "empty account, please create account")
			os.Exit(1)
		}

		var find bool = false
		if Address != "" {
			add := innerCommon.NewAddress(innerCommon.CopyBytes(innerCommon.FromHex(Address)))
			for _, v := range wallet.KeyData.Accounts {
				address := account.AddressFromPubKey(v.PublicKey)
				if address.Equals(&add) {
					common.Account = v
					find = true
				}
			}
		}

		if !find {
			for _, v := range wallet.KeyData.Accounts {
				common.Account = v
				break
			}
		}*/

}
