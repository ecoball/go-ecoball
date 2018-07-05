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
package sharding


type RegReq struct {

}

type RegRsp struct {
	addrInfo string   //
	upTime   uint32
}


type CommitteeInfoNotify struct {
	members []string
	leader  string
}

type CommitteeInfoAck struct {

}

type ShardPreReq struct {

}

type ShardPreRsp struct {
	memberCount uint32
	preLeader   string
	preBackup   string
}

type ShardingInfo struct {
	shardCount uint32
	newLeader  string
	neweBackup string
}

type JoinTopicSpaceReq struct {
	id string
}

type JoinTopicSpaceRsp struct {
	ret string
}




