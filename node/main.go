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
	"os"

	"github.com/ecoball/go-ecoball/common/config"
	"github.com/ecoball/go-ecoball/common/elog"
	"github.com/urfave/cli"
)

var log = elog.NewLogger("Node", elog.DebugLog)

func main() {
	app := cli.NewApp()

	//set attribute of EcoBall
	app.Name = "ecoball"
	app.Version = config.EcoVersion
	app.HelpName = "ecoball"
	app.Usage = "Block chain system from QuakerChain Technology"
	app.UsageText = "ecoball is a brand-new, open and compatible multi-chain parallel block chain operating system"
	app.Copyright = "2018 ecoball. All rights reserved"
	app.Author = "EcoBall"
	app.Email = "service@ecoball.org"
	app.HideHelp = true
	app.HideVersion = true

	//befor function
	app.Before = func(*cli.Context) error {
		log.Info("Start aba process...")

		//create config file or load confing file
		config.CreateOrReadConfig()
		return nil
	}

	//commands
	app.Commands = []cli.Command{
		RunCommand,
	}

	//flags
	app.Flags = []cli.Flag{
		NewNameFlag(),
		NewPasswordFlag(),
	}

	//run
	app.Run(os.Args)
}
