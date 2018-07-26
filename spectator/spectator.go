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

package spectator

import (
	"net"

	"github.com/ecoball/go-ecoball/common/elog"
	"github.com/ecoball/go-ecoball/core/ledgerimpl/ledger"
	"github.com/ecoball/go-ecoball/spectator/info"
	"github.com/ecoball/go-ecoball/spectator/syn"
)

var (
	log = elog.NewLogger("spectator", elog.DebugLog)
)

func Bystander(l ledger.Ledger) {
	syn.CoreLedger = l
	listener, err := net.Listen("tcp", "127.0.0.1:9000")
	if nil != err {
		log.Error("explorer server net.Listen error: ", err)
		return
	}
	defer listener.Close()

	buf := make([]byte, 1024)
	for {
		conn, err := listener.Accept()
		if nil != err {
			log.Error("explorer server net.Accept error: ", err)
			return
		}

		n, err := conn.Read(buf)
		if err != nil {
			log.Error("explorer server conn.Read error: ", err)
			continue
		}

		notify := info.OneNotify{info.InfoNil, []byte{}}
		if err := notify.Deserialize(buf[:n]); nil != err {
			log.Error("explorer server notify.Deserialize error: ", err)
			continue
		}
		go syn.Dispatch(conn, notify)

		info.Onlookers.Add(conn)
	}
}
