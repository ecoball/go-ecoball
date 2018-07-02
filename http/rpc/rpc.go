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

package rpc

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"sync"

	"github.com/ecoball/go-ecoball/common/config"
	"github.com/ecoball/go-ecoball/common/elog"
	"github.com/ecoball/go-ecoball/http/commands"
	"github.com/ecoball/go-ecoball/http/common"
	nrpc "github.com/ecoball/go-ecoball/net/rpc"
)

func init() {
	httpServer.method2Handle = make(map[string]func([]interface{}) *common.Response)
	rpcLog = elog.NewLogger("http", elog.NoticeLog)
}

var (
	rpcLog     elog.Logger
	httpServer HttpRpcServer
)

type HttpRpcServer struct {
	sync.RWMutex
	method2Handle map[string]func([]interface{}) *common.Response
}

func (this *HttpRpcServer) AddHandleFunc(pattern string, handler func([]interface{}) *common.Response) {
	this.Lock()
	defer this.Unlock()
	this.method2Handle[pattern] = handler
}

func Handle(w http.ResponseWriter, r *http.Request) {
	httpServer.RLock()
	defer httpServer.RUnlock()

	if r.Method == "OPTIONS" {
		w.Header().Add("Access-Control-Allow-Headers", "Content-Type")
		w.Header().Set("content-type", "application/json;charset=utf-8")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		return
	}
	//JSON RPC commands should be POSTs
	if r.Method != "POST" {
		rpcLog.Warn("HTTP JSON RPC Handle - Method!=\"POST\"")
		return
	}

	//check if there is Request Body to read
	if r.Body == nil {
		rpcLog.Warn("HTTP JSON RPC Handle - Request body is nil")
		return
	}

	//read the body of the request
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		rpcLog.Error("HTTP JSON RPC Handle - ioutil.ReadAll: ", err)
		return
	}

	request := make(map[string]interface{})
	err = json.Unmarshal(body, &request)
	if err != nil {
		rpcLog.Error("HTTP JSON RPC Handle - json.Unmarshal: ", err)
		return
	}
	if request["method"] == nil {
		rpcLog.Error("HTTP JSON RPC Handle - method not found: ")
		return
	}

	//get the corresponding function
	function, ok := httpServer.method2Handle[request["method"].(string)]
	if ok {
		rpcLog.Info("new http rpc call for method: ", request["method"])
		response := function(request["params"].([]interface{}))
		data, err := response.Serialize()
		if err != nil {
			rpcLog.Error("HTTP JSON RPC Handle - json.Marshal: ", err)
			return
		}
		w.Header().Add("Access-Control-Allow-Headers", "Content-Type")
		w.Header().Set("content-type", "application/json;charset=utf-8")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Write(data)
	} else {
		//if the function does not exist
		rpcLog.Warn("HTTP JSON RPC Handle - No function to call for ", request["method"])
		data, err := json.Marshal(map[string]interface{}{
			"errorCode": int64(-32601),
			"desc":      "The called method was not found on the server",
			"result":    nil,
		})
		if err != nil {
			rpcLog.Error("HTTP JSON RPC Handle - json.Marshal: ", err)
			return
		}
		w.Header().Add("Access-Control-Allow-Headers", "Content-Type")
		w.Header().Set("content-type", "application/json;charset=utf-8")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Write(data)
	}
}

func StartRPCServer() {
	http.HandleFunc("/", Handle)

	//add handle
	httpServer.AddHandleFunc("transfer", commands.Transfer)

	//query
	httpServer.AddHandleFunc("query", commands.Query)

	//set contract
	httpServer.AddHandleFunc("setContract", commands.SetContract)

	//invoke contract
	httpServer.AddHandleFunc("invokeContract", commands.InvokeContract)

	httpServer.AddHandleFunc("netlistmyid", nrpc.CliServerListMyId)
	httpServer.AddHandleFunc("netlistmypeer", nrpc.CliServerListMyPeers)

	//listen port
	err := http.ListenAndServe(":"+config.HttpLocalPort, nil)
	if err != nil {
		rpcLog.Fatal("ListenAndServe: ", err.Error())
	}
}
