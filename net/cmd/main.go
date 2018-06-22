package main

import (
	"github.com/ecoball/go-ecoball/net"
	"time"
	"encoding/hex"
	"math/big"
	"github.com/ecoball/go-ecoball/common/elog"
	"github.com/ecoball/go-ecoball/common"
	abaactor "github.com/ecoball/go-ecoball/common/event"
	"github.com/ecoball/go-ecoball/core/types"
	//ipfs "github.com/ecoball/go-ecoball/net/ipfs"
	//"fmt"
	//"fmt"
)
var workPath = "./store"
var log = elog.NewLogger("main", elog.DebugLog)

func send() {
	go func() {
		timerChan := time.NewTicker(500 * time.Microsecond).C
		for {
			select {
			case <-timerChan:
				fromUser, _ := hex.DecodeString("01b1a6569a557eafcccc71e0d02461fd4b601aea")
				toUser, _ := hex.DecodeString("01ca5cdd56d99a0023166b337ffc7fd0d2c42330")
				from := common.NewAddress(fromUser)
				to := common.NewAddress(toUser)
				value := big.NewInt(100)
				timeStamp := time.Now().Unix()
				//log.Debug(timeStamp)
				tx, _ := types.NewTransfer(from, to, value, 0, timeStamp)
				abaactor.Send(0, abaactor.ActorP2P, tx)
			}
		}
	}()
}

func main()  {
	net.StartNetWork()
	time.Sleep(5 * time.Second)
	//just for test
	send()
	//hash, _ := ipfs.AddDirectory("./store/blocks/")
	//fmt.Printf("hash : %s\n", hash)

	mainExitCh := make(chan int)

	<-mainExitCh
}
