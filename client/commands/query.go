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
	"github.com/ecoball/go-ecoball/client/rpc"
)

var (
	QueryCommands = cli.Command{
		Name:     "query",
		Usage:    "operations for query state",
		Category: "Query",
		Subcommands: []cli.Command{
			{
				Name:   "balance",
				Usage:  "query account's balance",
				Action: queryBalance,
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "address, a",
						Usage: "account address",
					},
				},
			},
		},
	}
)

func queryBalance(c *cli.Context) error {
	//Check the number of flags
	if c.NumFlags() == 0 {
		cli.ShowSubcommandHelp(c)
		return nil
	}

	//account address
	address := c.String("address")
	if address == "" {
		fmt.Println("Invalid account address: ", address)
		os.Exit(1)
	}

	//rpc call
	resp, err := rpc.Call("query", []interface{}{string("balance"), address})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return err
	}

	//result
	switch resp["result"].(type) {
	case map[string]interface{}:

	case string:
		fmt.Println(resp["result"].(string))
	}
	return nil
}
