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
	"fmt"
	"os"

	"github.com/urfave/cli"
	"github.com/ecoball/go-ecoball/account"
)

var (
	WalletCommands = cli.Command{
		Name:     "wallet",
		Usage:    "wallet operation",
		Category: "Wallet",
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "name",
				Usage: "wallet name",
			},
			cli.StringFlag{
				Name:  "password, p",
				Usage: "wallet password",
			},
			cli.BoolFlag{
				Name:  "create",
				Usage: "create wallet",
			},
			cli.BoolFlag{
				Name:  "list",
				Usage: "list wallet information",
			},
			cli.BoolFlag{
				Name:  "changepassword",
				Usage: "change wallet password",
			},
			cli.BoolFlag{
				Name:  "balance",
				Usage: "get balance",
			},
			cli.BoolFlag{
				Name:  "createaccount",
				Usage: "create account",
			},
		},
		Action: walletAction,
	}
)

func walletAction(c *cli.Context) error {
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
}
