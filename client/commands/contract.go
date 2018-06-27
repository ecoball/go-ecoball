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
	"io/ioutil"
	"os"

	"github.com/ecoball/go-ecoball/client/rpc"
	"github.com/urfave/cli"
)

var (
	ContractCommands = cli.Command{
		Name:        "contract",
		Usage:       "contract operate",
		Category:    "Contract",
		Description: "you could deploy or execute contract",
		ArgsUsage:   "[args]",
		Subcommands: []cli.Command{
			{
				Name:   "deploy",
				Usage:  "deploy contract",
				Action: setContract,
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "path, p",
						Usage: "contract file path",
					},
					cli.StringFlag{
						Name:  "name, n",
						Usage: "contract name",
					},
					cli.StringFlag{
						Name:  "description, d",
						Usage: "contract description",
					},
					cli.StringFlag{
						Name:  "author, a",
						Usage: "contract author",
					},
					cli.StringFlag{
						Name:  "email, e",
						Usage: "author email",
					},
				},
			},
			{
				Name:   "invoke",
				Usage:  "invoke contract",
				Action: invokeContract,
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "address, a",
						Usage: "contract address",
					},
					cli.StringFlag{
						Name:  "method, m",
						Usage: "contract method",
					},
					cli.StringFlag{
						Name:  "param, p",
						Usage: "method parameters",
					},
				},
			},
		},
	}
)

func setContract(c *cli.Context) error {

	//Check the number of flags
	if c.NumFlags() == 0 {
		cli.ShowSubcommandHelp(c)
		return nil
	}

	//contract file path
	fileName := c.String("path")
	if fileName == "" {
		fmt.Println("Invalid file path: ", fileName)
		return errors.New("Invalid contrace file path")
	}

	//file data
	file, err := os.OpenFile(fileName, os.O_RDONLY, 0666)
	if err != nil {
		fmt.Println("open file failed")
		return errors.New("open file failed: " + fileName)
	}

	defer file.Close()
	data, err := ioutil.ReadAll(file)
	if err != nil {
		fmt.Println("read contract filr err: ", err.Error())
		return err
	}

	//contract name
	contractName := c.String("name")
	if contractName == "" {
		fmt.Println("Invalid contract name: ", contractName)
		return errors.New("Invalid contract name")
	}

	//contract description
	description := c.String("description")
	if description == "" {
		fmt.Println("Invalid contract description: ", description)
		return errors.New("Invalid contract description")
	}

	//contract author
	author := c.String("author")
	if author == "" {
		fmt.Println("Invalid contract author: ", author)
		return errors.New("Invalid contract author")
	}

	//author email
	email := c.String("email")
	if email == "" {
		fmt.Println("Invalid author email: ", email)
		return errors.New("Invalid author email")
	}

	//rpc call
	resp, err := rpc.Call("setContract", []interface{}{string(data), contractName, description, author, email})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return err
	}

	//result
	switch resp["result"].(type) {
	case map[string]interface{}:

	case string:
		fmt.Println(resp["result"].(string))
		os.Exit(1)
	}

	return nil
}

func invokeContract(c *cli.Context) error {
	//Check the number of flags
	if c.NumFlags() == 0 {
		cli.ShowSubcommandHelp(c)
		return nil
	}

	//contract address
	contractAddress := c.String("address")
	if contractAddress == "" {
		fmt.Println("Invalid contract address: ", contractAddress)
		return errors.New("Invalid contract address")
	}

	//contract name
	contractMethod := c.String("method")
	if contractMethod == "" {
		fmt.Println("Invalid contract method: ", contractMethod)
		return errors.New("Invalid contract method")
	}

	//contract parameter
	contractParam := c.String("param")

	//rpc call
	resp, err := rpc.Call("invokeContract", []interface{}{contractAddress, contractMethod, contractParam})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return err
	}

	//result
	switch resp["result"].(type) {
	case map[string]interface{}:

	case string:
		fmt.Println(resp["result"].(string))
		os.Exit(1)
	}

	return nil
}
