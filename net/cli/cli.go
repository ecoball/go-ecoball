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
package cli

import (
	"fmt"
	"os"

	"github.com/urfave/cli"
	"github.com/ecoball/go-ecoball/client/rpc"
	"github.com/ecoball/go-ecoball/common/elog"
)

var url = "http://" + "127.0.0.1" + ":" + "16700"

var log = elog.NewLogger("cli", elog.DebugLog)

var (
	P2pCommand = cli.Command{
		Action:      cli.ShowSubcommandHelp,
		Name:        "p2p",
		Usage:       "Manage p2p",
		ArgsUsage:   " ",
		Description: `Manage p2p`,
		Subcommands: []cli.Command{
			{
				Action:      listMyId,
				Name:        "id",
				Usage:       "show my id",
				ArgsUsage:   " ",
				Flags:       nil,
				Description: `show my id`,
			},
			{
				Action:      listPeers,
				Name:        "peer",
				Usage:       "show my peers",
				ArgsUsage:   " ",
				Flags:       nil,
				Description: `show my peers`,
			},
		},
	}
)

var (
	NetworkCommand = cli.Command{
		Action:      cli.ShowSubcommandHelp,
		Name:        "network",
		Usage:       "Manage network",
		ArgsUsage:   " ",
		Description: `Manage network`,
		Subcommands: []cli.Command{
			P2pCommand,
		},
	}
)

func listMyId(ctx *cli.Context) error {

	rsp, err := rpc.Call("netlistmyid", []interface{}{})
	if err != nil {
		//TODO
	}

	//r := make(map[string]interface{})
	//err = json.Unmarshal(rsp,&r)
	//if err != nil {
	//TODO
	//}

	switch rsp["result"].(type) {
	case map[string]interface{}:

	case string:
		fmt.Println(rsp["result"].(string))
		os.Exit(1)
	}

	return nil
}

func listPeers(ctx *cli.Context) error {

	rsp, err := rpc.Call("netlistmypeer", []interface{}{})
	if err != nil {
		//TODO
	}

	//r := make(map[string]interface{})
	//err = json.Unmarshal(rsp,&r)
	//if err != nil {
	//TODO
	//}
	fmt.Printf("peer:%+v\n", rsp)
	switch res := rsp["result"].(type) {
	case []string:
		for _, pid := range res {
			fmt.Print(pid)
		}
		fmt.Println(res)
		os.Exit(1)
	default:
		fmt.Printf("peer:%+v\n", res)
	}

	return nil
}
