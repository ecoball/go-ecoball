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

package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/ecoball/go-ecoball/node/node"
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "run node",
	Long:  ``,
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		node.RunNode(viper.GetViper())
	},
}

func init() {
	RootCmd.AddCommand(runCmd)
	RootCmd.Flags().StringVarP(&node.Name, "name", "n", "", "wallet file name")
	RootCmd.Flags().StringVarP(&node.Password, "password", "p", "", "wallet password")
}
