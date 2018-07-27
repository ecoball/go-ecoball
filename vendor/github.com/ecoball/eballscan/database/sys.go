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

package database

import (
	"encoding/json"
	"net"

	"github.com/ecoball/go-ecoball/spectator/info"
)

type BlockHight int

func (this *BlockHight) Serialize() ([]byte, error) {
	return json.Marshal(*this)
}

func (this *BlockHight) Deserialize(data []byte) error {
	return json.Unmarshal(data, this)
}

func SynBlocks(conn net.Conn) {
	hight := BlockHight(MaxHight)
	oneNotify, err := info.NewOneNotify(info.SynBlock, &hight)
	if nil != err {
		log.Error("SynBlocks newOneNotify error: ", err)
		return
	}

	info, err := oneNotify.Serialize()
	if nil != err {
		log.Error("SynBlocks Serialize error: ", err)
		return
	}

	if _, err := conn.Write(info); nil != err {
		log.Error("SynBlocks Write error: ", err)
	}
}
