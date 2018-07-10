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
		Name:     "wallet",
		Usage:    "wallet operation",
		Category: "Wallet",
		Subcommands: []cli.Command{
			{
				Name:   "createwallet",
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
				Name:   "createaccount",
				Usage:  "create account",
				Action: createAccount,
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "name, n",
						Usage: "wallet name",
					},
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
				Name:   "listaccount",
				Usage:  "list account",
				Action: listAccount,
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
		},
	}
)

/*func walletAction(c *cli.Context) error {
	if c.NumFlags() == 0 {
		cli.ShowSubcommandHelp(c)
		return nil
	}
	name := c.String("name")
	create := c.Bool("create")
	list := c.Bool("list")
	passwd := c.String("password")
	if name == "" {
		fmt.Println("Invalid wallet name.")
		os.Exit(1)
	}
	if passwd == "" {
		fmt.Println("Invalid password.")
		os.Exit(1)
	}
	var wallet *account.WalletImpl

	if create {
		wallet = account.Create(name, []byte(passwd))
	} else {
		wallet = account.Open(name, []byte(passwd))
	}
	if wallet == nil {
		fmt.Println("Failed to open wallet: ", name)
		os.Exit(1)
	}

	createac := c.Bool("createaccount")

	if wallet.Accounts == nil {
		fmt.Printf("error\n")
	}
	if createac {
		wallet.CreateAccount([]byte(passwd))
	}
	if list {
		wallet.ListAccount()
	}

	return nil
}*/

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
	wallet := account.Create(name, []byte(passwd))
	if nil == wallet {
		fmt.Println("create wallet failed!")
		return errors.New("create wallet failed!")
	} else {
		fmt.Println("create wallet success, wallet file path:", name)
	}

	return nil
}

func createAccount(c *cli.Context) error {
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

	accountName := c.String("account")
	if err := common.AccountNameCheck(accountName); nil != err {
		return err
	}

	//create account
	wallet := account.Open(name, []byte(passwd))
	if nil == wallet {
		fmt.Println("Failed to open wallet: ", name)
		return errors.New("Failed to open wallet: " + name)
	}

	if _, err := wallet.CreateAccount([]byte(passwd), accountName); err != nil {
		fmt.Println(err)
		return err
	} /*else {
		fmt.Println("private key of ", accountName, " is "+ToHex(ac.PrivateKey[:]))
		fmt.Println("public key of ", accountName, " is "+ToHex(ac.PublicKey[:]))
	}*/

	return nil
}

func listAccount(c *cli.Context) error {
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

	//list account
	wallet := account.Open(name, []byte(passwd))
	if nil == wallet {
		fmt.Println("Failed to open wallet: ", name)
		return errors.New("Failed to open wallet: " + name)
	}

	wallet.ListAccount()

	return nil
}
