// Copyright 2018 The eballscan Authors
// This file is part of the eballscan.
//
// The eballscan is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The eballscan is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the eballscan. If not, see <http://www.gnu.org/licenses/>.

package onlooker

import (
	"net"

	"github.com/ecoball/eballscan/database"
	"github.com/ecoball/eballscan/syn"
	"github.com/ecoball/go-ecoball/common"
	"github.com/ecoball/go-ecoball/common/elog"
	"github.com/ecoball/go-ecoball/core/types"
	"github.com/ecoball/go-ecoball/spectator/info"
)

var (
	log = elog.NewLogger("onlooker", elog.DebugLog)
)

func Bystander() {
	conn, err := net.Dial("tcp", "127.0.0.1:9000")
	if err != nil {
		log.Error("explorer server net.Dial error: ", err)
		return
	}

	syn.SynBlocks(conn)

	buf := make([]byte, 1024*10)
	for {
		n, err := conn.Read(buf)
		if err != nil {
			log.Error("explorer server conn.Read error: ", err)
			break
		}

		notify := info.OneNotify{info.InfoNil, []byte{}}
		if err := notify.Deserialize(buf[:n]); nil != err {
			log.Error("explorer server notify.Deserialize error: ", err)
			continue
		}
		go dispatch(notify)
	}
}

func dispatch(notify info.OneNotify) {
	switch notify.InfoType {
	case info.InfoBlock:
		if err := handleBlock(notify.Info); nil != err {
			log.Error("handleBlock error: ", err)
		}
	default:

	}
}

func handleBlock(info []byte) error {
	oneBlock := types.Block{}
	if err := oneBlock.Deserialize(info); nil != err {
		return err
	}

	return database.AddBlock(int(oneBlock.Height), int(oneBlock.CountTxs), common.ToHex(oneBlock.Hash.Bytes()), common.ToHex(oneBlock.PrevHash.Bytes()),
		common.ToHex(oneBlock.MerkleHash.Bytes()), common.ToHex(oneBlock.StateHash.Bytes()))
}
