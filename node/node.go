package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/ecoball/go-ecoball/core/ledgerimpl"
	"github.com/ecoball/go-ecoball/core/store"
	"github.com/ecoball/go-ecoball/net"
	"github.com/ecoball/go-ecoball/net/rpc"
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
	checkPassword()

	fmt.Println("Run Node")
	log.Info("Build Geneses Block")
	_, err := ledgerimpl.NewLedger(store.PathBlock)
	if err != nil {
		log.Fatal(err)
	}

	//start transaction pool
	if _, err = txpool.Start(); err != nil {
		log.Fatal("start txpool error, ", err.Error())
		os.Exit(1)
	}

	net.StartNetWork()
	//TODO, modify temporily
	//consensus.InitConsensusService(l, consensus.CONSENSUS_SOLO, nil)
	//TODO, modify temporily

	//if err != nil {
	//	log.Fatal("Starting net server failed")
	//	os.Exit(1)
	//}

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
