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

package info

import (
	"encoding/json"
	"fmt"
	"net"
	"sync"

	"github.com/ecoball/go-ecoball/common/elog"
)

var (
	Onlookers = onlooker{connects: make([]net.Conn, 0, 10)}
	log       = elog.NewLogger("info", elog.DebugLog)
)

type NotifyInfo interface {
	Serialize() ([]byte, error)
	Deserialize(data []byte) error
}

type NotifyType int

const (
	InfoNil NotifyType = iota
	InfoBlock
	SynBlock
)

type OneNotify struct {
	InfoType NotifyType
	Info     []byte
}

func NewOneNotify(oneType NotifyType, message NotifyInfo) (*OneNotify, error) {
	oneMessage, err := message.Serialize()
	if nil != err {
		return nil, err
	}
	return &OneNotify{oneType, oneMessage}, nil
}

func (this *OneNotify) Serialize() ([]byte, error) {
	return json.Marshal(*this)
}

func (this *OneNotify) Deserialize(data []byte) error {
	return json.Unmarshal(data, this)
}

type onlooker struct {
	connects []net.Conn
	sync.Mutex
}

func (this *onlooker) Add(conn net.Conn) {
	this.Lock()
	defer this.Unlock()

	for _, v := range this.connects {
		if conn == v {
			return
		}
	}

	this.connects = append(this.connects, conn)
}

func (this *onlooker) notify(info []byte) {
	this.Lock()
	defer this.Unlock()

	for k, v := range this.connects {
		if _, err := v.Write(info); nil != err {
			addr := v.RemoteAddr().String()
			log.Warn(addr, " disconnect")
			this.connects = append(this.connects[:k], this.connects[k+1:]...)
		}

		fmt.Println(string(info))
	}
}

func Notify(infoType NotifyType, message NotifyInfo) error {
	info, err := NewOneNotify(infoType, message)
	if nil != err {
		return err
	}

	data, err := info.Serialize()
	if nil != err {
		return nil
	}

	Onlookers.notify(data)
	return nil
}
