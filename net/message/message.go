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

package message

import (
	"io"
	"github.com/ecoball/go-ecoball/net/message/pb"
	inet "gx/ipfs/QmYj8wdn5sZEHX2XMDWGBvcXJNdzVbaVpHmXvhHBVZepen/go-libp2p-net"
	ggio "gx/ipfs/QmZ4Qi3GaRbjcx28Sme5eMH7RQjGkt8wHxt2a65oLaeFEV/gogo-protobuf/io"
	"github.com/ecoball/go-ecoball/common/elog"
)

const (
	APP_MSG_TRN  uint32 = iota
	APP_MSG_BLK
	APP_MSG_GOSSIP_PULL_BLK_REQ
	APP_MSG_GOSSIP_PULL_BLK_ACK
	APP_MSG_GOSSIP_PUSH_BLKS
)
var log = elog.NewLogger("message", elog.DebugLog)
type HandlerFunc func(data []byte) (err error)

type EcoBallNetMsg interface {
	ChainID() uint32
	Type() uint32
	Data() []byte
	Exportable
}


type Exportable interface {
	ToProtoV1() *pb.Message
	ToNetV1(w io.Writer) error
}

type impl struct {
	chainId     uint32
	msgType     uint32
	data        []byte
}


func New(msgType uint32, data []byte) EcoBallNetMsg {
	return newMsg(msgType, data)
}

func newMsg(msgType uint32, data []byte) *impl {
	return &impl{
		chainId: 1,//TODO
		msgType: msgType,
		data: data,
	}
}

func newMessageFromProto(pbm pb.Message) (EcoBallNetMsg, error) {
	m := newMsg(pbm.Type, pbm.Data)
	return m, nil
}

func (m *impl) ChainID() uint32 {
	return m.chainId
}

func (m *impl) Type() uint32 {
	return m.msgType
}

func (m *impl) Data() []byte {
	return m.data
}

func (m *impl) ToProtoV1() *pb.Message {
	pbm := new(pb.Message)
	pbm.ChainId = m.chainId
	pbm.Data = m.data
	pbm.Type = m.msgType
	return pbm
}

func (m *impl) ToNetV1(w io.Writer) error {
	pbw := ggio.NewDelimitedWriter(w)
	return pbw.WriteMsg(m.ToProtoV1())
}

func FromNet(r io.Reader) (EcoBallNetMsg, error) {
	pbr := ggio.NewDelimitedReader(r, inet.MessageSizeMax)
	return FromPBReader(pbr)
}

func FromPBReader(pbr ggio.Reader) (EcoBallNetMsg, error) {
	pb := new(pb.Message)
	if err := pbr.ReadMsg(pb); err != nil {
		return nil, err
	}
	return newMessageFromProto(*pb)
}
