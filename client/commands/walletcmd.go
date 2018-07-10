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
package commands

import (
	"errors"
	"fmt"

	"github.com/ecoball/go-ecoball/account"
	"github.com/ecoball/go-ecoball/common"
	"github.com/urfave/cli"
)

var (
	WalletCommands = cli.Command{
		Name:        "wallet",
		Usage:       "wallet operation",
		Category:    "Wallet",
		Description: "wallet operate",
		ArgsUsage:   "[args]",
		Subcommands: []cli.Command{
			{
				Name:   "create",
				Usage:  "create wallet",
				Action: createWallet,
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "name, n",
						Usage: "wallet name",
					},
					cli.StringFlag{
						Name:  "password, p",
						Usage: "wallet password",
					},
				},
			},
			{
				Name:   "open",
				Usage:  "open wallet",
				Action: openWallet,
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "name, n",
						Usage: "wallet name",
					},
					cli.StringFlag{
						Name:  "password, p",
						Usage: "wallet password",
					},
				},
			},
			{
				Name:   "lock",
				Usage:  "lock wallet",
				Action: lockWallet,
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "password, p",
						Usage: "wallet password",
					},
				},
			},
			{
				Name:   "unlock",
				Usage:  "unlock wallet",
				Action: unlockWallet,
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "password, p",
						Usage: "wallet password",
					},
				},
			},
			{
				Name:   "close",
				Usage:  "close wallet",
				Action: closeWallet,
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "password, p",
						Usage: "wallet password",
					},
				},
			},
			{
				Name:   "createaccount",
				Usage:  "create account",
				Action: createAccount,
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "password, p",
						Usage: "wallet password",
					},
					cli.StringFlag{
						Name:  "account, a",
						Usage: "account name",
					},
				},
			},
			{
				Name:   "list",
				Usage:  "list account",
				Action: listAccount,
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "password, p",
						Usage: "wallet password",
					},
				},
			},
		},
	}
)

func createWallet(c *cli.Context) error {
	//Check the number of flags
	if c.NumFlags() == 0 {
		cli.ShowSubcommandHelp(c)
		return nil
	}

	name := c.String("name")
	if "" == name {
		fmt.Println("Invalid wallet name")
		return errors.New("Invalid wallet name")
	}

	passwd := c.String("password")
	if "" == passwd {
		fmt.Println("Invalid password")
		return errors.New("Invalid password")
	}

	//create wallet file
	if err := account.Create(name, []byte(passwd)); nil != err {
		fmt.Println(err)
		return err
	}

	fmt.Println("create wallet success, wallet file path:", name)
	return nil
}

func openWallet(c *cli.Context) error {
	//Check the number of flags
	if c.NumFlags() == 0 {
		cli.ShowSubcommandHelp(c)
		return nil
	}

	name := c.String("name")
	if "" == name {
		fmt.Println("Invalid wallet name")
		return errors.New("Invalid wallet name")
	}

	passwd := c.String("password")
	if "" == passwd {
		fmt.Println("Invalid password")
		return errors.New("Invalid password")
	}

	//whether the wallet open
	if nil != account.Wallet {
		fmt.Println("The wallet has been opened!")
	}

	//open wallet
	wallet, err := account.Open(name, []byte(passwd))
	if nil != err {
		fmt.Println(err)
		return err
	}

	fmt.Println("open wallet success, wallet file path:", name)
	account.Wallet = wallet
	return nil
}

func lockWallet(c *cli.Context) error {
	//Check the number of flags
	if c.NumFlags() == 0 {
		cli.ShowSubcommandHelp(c)
		return nil
	}

	passwd := c.String("password")
	if "" == passwd {
		fmt.Println("Invalid password")
		return errors.New("Invalid password")
	}

	//whether the wallet open
	if nil == account.Wallet {
		fmt.Println("The wallet has not been opened!")
		return errors.New("The wallet has not been opened!")
	}

	if nil != account.Cipherkeys {
		fmt.Println("the data is wrong!")
		return errors.New("the data is wrong!")
	}

	//lock wallet
	cipherkeysTemp, err := account.Wallet.Lock([]byte(passwd))
	if nil != err {
		fmt.Println(err)
		return err
	}

	//write data
	if err := account.Wallet.StoreWallet(cipherkeysTemp); nil != err {
		fmt.Println(err)
		return err
	}

	fmt.Println("lock wallet success")
	account.Cipherkeys = cipherkeysTemp
	return nil
}

func unlockWallet(c *cli.Context) error {
	//Check the number of flags
	if c.NumFlags() == 0 {
		cli.ShowSubcommandHelp(c)
		return nil
	}

	passwd := c.String("password")
	if "" == passwd {
		fmt.Println("Invalid password")
		return errors.New("Invalid password")
	}

	//whether the wallet open
	if nil == account.Wallet {
		fmt.Println("The wallet has not been opened!")
		return errors.New("The wallet has not been opened!")
	}

	//whether the wallet locked
	if !account.Wallet.CheckLocked() {
		fmt.Println("The wallet has not been locked!")
		return errors.New("The wallet has not been locked!")
	}

	if nil == account.Cipherkeys {
		fmt.Println("the data is wrong!")
		return errors.New("the data is wrong!")
	}

	//unlock wallet
	if err := account.Wallet.Unlock([]byte(passwd), account.Cipherkeys); nil != err {
		fmt.Println(err)
		return err
	}

	fmt.Println("unlock wallet success")
	account.Cipherkeys = nil
	return nil
}

func closeWallet(c *cli.Context) error {
	//Check the number of flags
	if c.NumFlags() == 0 {
		cli.ShowSubcommandHelp(c)
		return nil
	}

	passwd := c.String("password")
	if "" == passwd {
		fmt.Println("Invalid password")
		return errors.New("Invalid password")
	}

	//whether the wallet open
	if nil == account.Wallet {
		fmt.Println("The wallet has not been opened!")
		return errors.New("The wallet has not been opened!")
	}

	//close wallet
	if err := account.Wallet.Close([]byte(passwd)); nil != err {
		fmt.Println(err)
		return err
	}

	fmt.Println("close wallet success")
	account.Cipherkeys = nil
	account.Wallet = nil
	return nil
}

func createAccount(c *cli.Context) error {
	//Check the number of flags
	if c.NumFlags() == 0 {
		cli.ShowSubcommandHelp(c)
		return nil
	}

	passwd := c.String("password")
	if "" == passwd {
		fmt.Println("Invalid password")
		return errors.New("Invalid password")
	}

	//check account name
	accountName := c.String("account")
	if err := common.AccountNameCheck(accountName); nil != err {
		fmt.Println(err)
		return err
	}

	//whether the wallet open
	if nil == account.Wallet {
		fmt.Println("The wallet has not been opened!")
	}

	//whether the wallet locked
	if account.Wallet.CheckLocked() {
		fmt.Println("The wallet has been locked!")
		return errors.New("The wallet has been locked!")
	}

	//create account
	ac, err := account.Wallet.CreateAccount([]byte(passwd), accountName)
	if err != nil {
		fmt.Println(err)
		return err
	}

	fmt.Println("create account suuccess, account name: ", accountName)
	fmt.Println("PrivateKey: ", common.ToHex(ac.PrivateKey[:]))
	fmt.Println("PublicKey: ", common.ToHex(ac.PublicKey[:]))
	return nil
}

func listAccount(c *cli.Context) error {
	//Check the number of flags
	if c.NumFlags() == 0 {
		cli.ShowSubcommandHelp(c)
		return nil
	}

	passwd := c.String("password")
	if "" == passwd {
		fmt.Println("Invalid password")
		return errors.New("Invalid password")
	}

	//whether the wallet open
	if nil == account.Wallet {
		fmt.Println("The wallet has not been opened!")
		return errors.New("The wallet has not been opened!")
	}

	//whether the wallet locked
	if account.Wallet.CheckLocked() {
		fmt.Println("The wallet has been locked!")
		return errors.New("The wallet has been locked!")
	}

	//list account
	account.Wallet.ListAccount()

	return nil
}
