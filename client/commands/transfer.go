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
	"os"

	"github.com/ecoball/go-ecoball/client/rpc"

	"github.com/urfave/cli"
)

var (
	TransferCommands = cli.Command{
		Name:        "transfer",
		Usage:       "user ABA transfer",
		Category:    "Transfer",
		Description: "With ecoclient transfer, you could transfer ABA to others",
		ArgsUsage:   "[args]",
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "from, f",
				Usage: "sender address",
			},
			cli.StringFlag{
				Name:  "to, t",
				Usage: "revicer address",
			},
			cli.Int64Flag{
				Name:  "value, v",
				Usage: "ABA amount",
			},
		},
		Action: transferAction,
		OnUsageError: func(c *cli.Context, err error, isSubcommand bool) error {
			return cli.NewExitError("", 1)
		},
	}
)

func transferAction(c *cli.Context) error {
	//Check the number of flags
	if c.NumFlags() == 0 {
		cli.ShowSubcommandHelp(c)
		return nil
	}

	from := c.String("from")
	if from == "" {
		fmt.Println("Invalid sender address: ", from)
		return errors.New("Invalid sender address")
	}

	to := c.String("to")
	if to == "" {
		fmt.Println("Invalid revicer address: ", to)
		return errors.New("Invalid revicer address")
	}

	value := c.Int64("value")
	if value <= 0 {
		fmt.Println("Invalid aba amount: ", value)
		return errors.New("Invalid aba amount")
	}

	resp, err := rpc.Call("transfer", []interface{}{from, to, value})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return err
	}

	//result
	return rpc.EchoResult(resp)
}
