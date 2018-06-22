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
	"time"
	//"net/http"
	//erpc "github.com/ecoball/go-ecoball/http/rpc"
	"github.com/ecoball/go-ecoball/common/elog"
	eactor "github.com/ecoball/go-ecoball/common/event"
	"github.com/ecoball/go-ecoball/http/common"
)

var log = elog.NewLogger("NetRpc", elog.DebugLog)

var rpcPort = 16700

//This just for test -:)
func StartRPCServer() error {

	//http.HandleFunc("/", erpc.Handle)

	//erpc.HandleFunc("netlistmyid", CliServerListMyId)
	//erpc.HandleFunc("netlistmypeer", CliServerListMyPeers)
	//addr := "127.0.0.1:16700"
	//err := http.ListenAndServe(addr, nil)
	//if err != nil {
	//TODO
	//}

	return nil
}

//server do
func networkListMyId() (string, error) {
	req := new(ListMyIdReq)
	result, err := eactor.SendSync(eactor.ActorP2P, req, 5*time.Second)
	if err != nil {
		log.Error(err.Error())
		return "", err
	}
	rsp, ok := result.(*ListMyIdRsp)
	if !ok {
		log.Error(err.Error())
		return "", err
	}
	return rsp.Id, nil
}

func CliServerListMyId(params []interface{}) *common.Response {
	id, err := networkListMyId()
	if err != nil {
		return common.NewResponse(common.INTERNAL_ERROR, false)
	}
	log.Debug("id ", id)
	return common.NewResponse(common.SUCCESS, id)
}

func networkListMyPeers() ([]string, error) {
	req := new(ListPeersReq)
	result, err := eactor.SendSync(eactor.ActorP2P, req, 5*time.Second)
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}
	rsp, ok := result.(*ListPeersRsp)
	if !ok {
		log.Error(err.Error())
		return nil, err
	}
	return rsp.Peer, nil
}

func CliServerListMyPeers(params []interface{}) *common.Response {
	peers, err := networkListMyPeers()
	if err != nil {
		return common.NewResponse(common.INTERNAL_ERROR, false)
	}
	log.Debug(peers)
	return common.NewResponse(common.SUCCESS, peers)
}
