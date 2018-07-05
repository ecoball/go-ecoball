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

import (
	"gx/ipfs/QmaWsab8a1KQgoxWP3RjK7mBhSi5PB9pR6NwZUrSXvVd1i/go-libp2p-floodsub"
)
type TopicSpaceRouter struct {
	p         *floodsub.PubSub
}

type TopicSpace struct {
	p *floodsub.PubSub
}

func (ts *TopicSpace)joinCheck()  {

}

func (ts *TopicSpace)Join()  {

}

func (ts *TopicSpace)Leave()  {

}

func (ts *TopicSpace)MemberCount() uint32 {
	return 0
}

func (ts *TopicSpace)PutMsg()  {

}

func (ts *TopicSpace)GetMsg()  {

}


