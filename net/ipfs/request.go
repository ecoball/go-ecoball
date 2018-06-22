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

package ipfs

import (
	"context"
	cmds "gx/ipfs/QmSKYWC84fqkKB54Te5JMcov2MBVzucXaRGxFqByzzCbHe/go-ipfs-cmds"
	cli "gx/ipfs/QmSKYWC84fqkKB54Te5JMcov2MBVzucXaRGxFqByzzCbHe/go-ipfs-cmds/cli"
	commands "github.com/ipfs/go-ipfs/core/commands"
	"time"
	"github.com/ecoball/go-ecoball/common/elog"
)

var log = elog.NewLogger("ipfs", elog.DebugLog)

var Root = &cmds.Command{
	Options:  commands.Root.Options,
	Helptext: commands.Root.Helptext,
}

var commandsClientCmd = commands.CommandsCmd(Root)

var localCommands = map[string]*cmds.Command{
	"commands": commandsClientCmd,
}

func init() {
	Root.Subcommands = localCommands
	for k, v := range commands.Root.Subcommands {
		if _, found := Root.Subcommands[k]; !found {
			Root.Subcommands[k] = v
		}
	}
}

func NewRequest(args []string) (*cmds.Request, error) {
	Root.Subcommands = localCommands
	for k, v := range commands.Root.Subcommands {
		if _, found := Root.Subcommands[k]; !found {
			Root.Subcommands[k] = v
		}
	}
	cctx := context.Background()
	req, err := cli.Parse(cctx, args, nil, Root)
	return req, err
}

func NewRequestWithTimeout(args []string, timeout time.Duration) (*cmds.Request, error) {
	Root.Subcommands = localCommands
	for k, v := range commands.Root.Subcommands {
		if _, found := Root.Subcommands[k]; !found {
			Root.Subcommands[k] = v
		}
	}
	cctx, _ := context.WithTimeout(context.Background(), timeout)
	req, err := cli.Parse(cctx, args, nil, Root)
	return req, err
}
