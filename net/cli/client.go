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
	"os"

	"github.com/ecoball/go-ecoball/common/config"
	"github.com/urfave/cli"
)

func main() {

	app := cli.NewApp()
	app.Name = "ecocli"
	app.Version = config.EcoVersion
	app.HelpName = "ecocli"
	app.Usage = "command line tool for ecoball"
	app.UsageText = "ecoclient [global options] command [command options] [args]"
	app.HideHelp = false
	app.HideVersion = false

	app.Commands = []cli.Command{
		NetworkCommand,
	}

	app.Run(os.Args)
}
