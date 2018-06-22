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

package node

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/ecoball/go-ecoball/http/rpc"
	"github.com/ecoball/go-ecoball/txpool"

	"github.com/spf13/viper"
	"github.com/ecoball/go-ecoball/common/elog"
	"github.com/ecoball/go-ecoball/core/ledgerimpl"
	"github.com/ecoball/go-ecoball/core/store"
	"github.com/ecoball/go-ecoball/net"
	"github.com/ecoball/go-ecoball/webserver"
)

var (
	log      = elog.NewLogger("Node", elog.DebugLog)
	Name     string
	Password string
)

func RunNode(config *viper.Viper) {
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
}

func wait() {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	defer signal.Stop(interrupt)
	sig := <-interrupt
	fmt.Println("ecoball received signal:", sig)
}

func checkPassword() {
	/*	var (
				nameTime     = 0
				passwordTime = 0
			)

			//empty name
			if "" == Name {
				fmt.Printf("please input wallet file name:")
				fmt.Scan(&Name)
				goto name
			}

		name:
			if 0 != nameTime {
				fmt.Printf("please input wallet file name:")
				fmt.Scan(&Name)
			}

			//file does not exist
			if _, err := os.Stat(Name); err != nil {
				fmt.Fprintf(os.Stderr, "%v", err)
				nameTime++
				if nameTime >= 3 {
					fmt.Fprintln(os.Stderr, "More than three times, maybe you didn't create your wallet, exit...")
					os.Exit(1)
				}
				goto name
			}

			//empty password
			if "" == Password {
				fmt.Printf("please input wallet password:")
				fmt.Scan(&Password)
				goto password
			}

		password:
			if 0 != passwordTime {
				fmt.Printf("please input wallet password:")
				fmt.Scan(&Password)
			}

			//worng password
			wallet := account.Open(Name, []byte(Password))
			if nil == wallet {
				fmt.Fprintln(os.Stderr, "open wallet failed!")
				passwordTime++
				if passwordTime >= 3 {
					fmt.Fprintln(os.Stderr, "More than three times, exit...")
					os.Exit(1)
				}
				goto password
			}

			//get account
			if 0 == len(wallet.KeyData.Accounts) {
				fmt.Fprintln(os.Stderr, "empty account, please create account")
				os.Exit(1)
			}

			for _, v := range wallet.KeyData.Accounts {
				common.Account = v
			}*/
}
