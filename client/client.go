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

package main

import (
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"sort"
	"strings"
	"syscall"

	"github.com/ecoball/go-ecoball/client/commands"
	"github.com/ecoball/go-ecoball/client/common"
	"github.com/ecoball/go-ecoball/common/config"
	ncli "github.com/ecoball/go-ecoball/net/cli"
	"github.com/peterh/liner"
	"github.com/urfave/cli"
)

var (
	historyFilePath = filepath.Join(os.TempDir(), ".ecoclient_history")
	commandName     = []string{"contract", "transfer", "wallet", "query", "attach"}
)

func newClientApp() *cli.App {
	app := cli.NewApp()

	//set attribute of client
	app.Name = "ecoclient"
	app.Version = config.EcoVersion
	app.HelpName = "ecoclient"
	app.Usage = "command line tool for ecoball"
	app.UsageText = "ecoclient [global options] command [command options] [args]"
	app.Copyright = "2018 ecoball. All rights reserved"
	app.Author = "ecoball"
	app.Email = "service@ecoball.org"
	app.HideHelp = true
	app.HideVersion = true

	//commands
	app.Commands = []cli.Command{
		commands.ContractCommands,
		commands.TransferCommands,
		commands.WalletCommands,
		commands.QueryCommands,
		commands.AttachCommands,
		commands.CreateCommands,
		ncli.P2pCommand,
	}

	sort.Sort(cli.CommandsByName(app.Commands))

	return app
}

func main() {
	app := newClientApp()

	//console
	app.After = func(c *cli.Context) error {
		newConsole()
		return nil
	}

	//run
	app.Run(os.Args)
}

func newConsole() {
	state := liner.NewLiner()
	defer func() {
		state.Close()
		if err := recover(); err != nil {
			fmt.Println("panic occur：", err)
		}
	}()

	//set attribute of console
	state.SetCtrlCAborts(true)
	state.SetTabCompletionStyle(liner.TabPrints)
	state.SetMultiLineMode(true)
	state.SetCompleter(func(line string) (c []string) {
		for _, n := range commandName {
			if strings.HasPrefix(n, strings.ToLower(line)) {
				c = append(c, n)
			}
		}
		return
	})

	//read history info
	var historyFile *os.File
	var err error
	if common.FileExisted(historyFilePath) {
		historyFile, err = os.Open(historyFilePath)
		_, err = state.ReadHistory(historyFile)
	} else {
		historyFile, err = os.Create(historyFilePath)
	}

	if nil != err {
		fmt.Println(err)
		os.Exit(1)
	}

	defer func() {
		state.WriteHistory(historyFile)
		historyFile.Close()
	}()

	//new console
	scheduler := make(chan string)

	go func() {
		for {
			info := <-scheduler
			line, errLine := state.Prompt(info)
			if errLine == nil {
				state.AppendHistory(line)
				scheduler <- line
			} else if errLine == liner.ErrPromptAborted {
				fmt.Println("Aborted")
				close(scheduler)
				return
			} else {
				fmt.Println("Error reading line: ", errLine)
				close(scheduler)
				return
			}
		}
	}()

	//single abort
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

	scheduler <- "ecoclient: \\>"
	for {
		select {
		case <-sig:
			fmt.Println("exit signal")
			return
		case line, ok := <-scheduler:
			if ok {
				if "exit" == line {
					return
				} else {
					handleLine(line)
				}
				scheduler <- "ecoclient: \\>"
			}
		}
	}
}

func handleLine(line string) error {
	args := []string{os.Args[0]}
	lines := strings.Fields(line)
	args = append(args, lines...)

	//run
	newClientApp().Run(args)
	return nil
}
