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
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/ecoball/go-ecoball/common/config"
	"github.com/ecoball/go-ecoball/consensus/solo"
	"github.com/ecoball/go-ecoball/core/ledgerimpl"
	"github.com/ecoball/go-ecoball/core/store"
	"github.com/ecoball/go-ecoball/http/rpc"
	"github.com/ecoball/go-ecoball/net"
	"github.com/ecoball/go-ecoball/txpool"
	"github.com/ecoball/go-ecoball/webserver"
	"github.com/urfave/cli"
)

var (
	RunCommand = cli.Command{
		Name:   "run",
		Usage:  "run node",
		Action: runNode,
	}
)

func runNode(c *cli.Context) error {
	//get account
	//checkPassword()

	fmt.Println("Run Node")
	log.Info("Build Geneses Block")
	l, err := ledgerimpl.NewLedger(store.PathBlock)
	if err != nil {
		log.Fatal(err)
	}
	log.Info("consensus", config.ConsensusAlgorithm)
	//start consensus
	switch config.ConsensusAlgorithm {
	case "SOLO":
		c, _ := solo.NewSoloConsensusServer(l)
		c.Start()
	case "DPOS":
		log.Info("Start DPOS consensus")

		l.Start()
	default:
		log.Fatal("unsupported consensus algorithm:", config.ConsensusAlgorithm)
	}
	//start transaction pool
	if _, err := txpool.Start(); err != nil {
		log.Fatal("start txpool error, ", err.Error())
		os.Exit(1)
	}

	net.StartNetWork(l)
	//start http server
	go rpc.StartRPCServer()

	//start web server
	go webserver.StartWebServer()

	//wait single to exit
	wait()

	return nil
}

func wait() {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	defer signal.Stop(interrupt)
	sig := <-interrupt
	log.Info("ecoball received signal:", sig)
}
