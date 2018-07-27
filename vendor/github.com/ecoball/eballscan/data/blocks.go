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

package data

import (
	"fmt"
	"sync"

	"github.com/ecoball/go-ecoball/common/elog"
)

var (
	Blocks = Block{BlocksInfo: make(map[int]BlockInfo, 0)}
	log    = elog.NewLogger("data", elog.DebugLog)
)

type BlockInfo struct {
	Hash       string
	PrevHash   string
	MerkleHash string
	StateHash  string
	CountTxs   int
}

type Block struct {
	BlocksInfo map[int]BlockInfo

	sync.RWMutex
}

func (this *Block) Add(hight int, info BlockInfo) {
	this.Lock()
	defer this.Unlock()

	if _, ok := this.BlocksInfo[hight]; ok {
		return
	}
	this.BlocksInfo[hight] = info
}

func PrintBlock() string {
	Blocks.RLock()
	defer Blocks.RUnlock()

	result := "hight\thash\tprevHash\tmerkleHash\tstateHash\tcountTxs\n"
	for k, v := range Blocks.BlocksInfo {
		result += fmt.Sprintf("%d\t%s\t%s\t%s\t%s\t%d\n", k, v.Hash, v.PrevHash, v.MerkleHash, v.StateHash, v.CountTxs)
	}

	return result
}
