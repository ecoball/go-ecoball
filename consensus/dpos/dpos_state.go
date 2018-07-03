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

package dpos

const (
	Second             = int64(1000)
	BlockInterval      = int64(15000)
	GenerationInterval = GenerationSize * BlockInterval * 10
	GenerationSize     = 4
	ConsensusThreshold = GenerationSize*2/3 + 1
	MaxProduceDuration = int64(5250)
	MinProduceDuration = int64(2250)
)

/*type State struct {
	timestamp int64
	leader common.Hash

	//TODO
	bookkeepers []common.Hash
}*/




