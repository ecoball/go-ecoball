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

package syn

import (
	"net"

	"github.com/ecoball/eballscan/database"
	"github.com/ecoball/go-ecoball/common/elog"
	"github.com/ecoball/go-ecoball/core/ledgerimpl/ledger"
	"github.com/ecoball/go-ecoball/spectator/info"
)

var (
	CoreLedger ledger.Ledger
	log        = elog.NewLogger("spectator", elog.DebugLog)
)

func Dispatch(conn net.Conn, notify info.OneNotify) {
	switch notify.InfoType {
	case info.SynBlock:
		if err := handleSynBlock(conn, notify.Info); nil != err {
			log.Error("handleBlock error: ", err)
		}
	default:

	}
}

func handleSynBlock(conn net.Conn, msg []byte) error {
	var blockHight database.BlockHight
	if err := blockHight.Deserialize(msg); nil != err {
		return err
	}

	hight := uint64(blockHight)

	nowHight := CoreLedger.GetCurrentHeight()
	for hight < nowHight {
		hight++

		block, err := CoreLedger.GetTxBlockByHeight(hight)
		if nil != err {
			log.Error("GetTxBlockByHeight error: ", err)
			continue
		}

		notify, err := info.NewOneNotify(info.InfoBlock, block)
		if nil != err {
			log.Error("NewOneNotify error: ", err)
			continue
		}

		data, err := notify.Serialize()
		if nil != err {
			log.Error("Serialize error: ", err)
			continue
		}

		if _, err := conn.Write(data); nil != err {
			addr := conn.RemoteAddr().String()
			log.Warn(addr, " disconnect")
			break
		}
	}

	return nil
}
