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

package ipfs

import (
	"github.com/ipfs/go-ipfs/path"
	"gx/ipfs/QmSKYWC84fqkKB54Te5JMcov2MBVzucXaRGxFqByzzCbHe/go-ipfs-cmds"
	"bytes"
	"github.com/ipfs/go-ipfs/core"
	"encoding/json"
	"fmt"
	"github.com/ipfs/go-ipfs/core/commands"
	cmd "github.com/ipfs/go-ipfs/commands"
)

type IpnsRspEntry struct {
	Value commands.IpnsEntry
}

// Publish a signed IPNS record to our Peer ID
func Publish(hash string) (string, error) {
	args := []string{"name", "publish", hash}

	result, err := publish(args)
	if err != nil {
		return "", err
	}

	fmt.Printf("Published %s: %s\n", result.Name, result.Value)

	return result.Name, nil
}

// Publish another IPFS record at /ipns/<peerID>:<altRoot>
// Valid lifetime units are "ns", "us" (or "µs"), "ms", "s", "m", "h"
func PublishAltRoot(altRoot string, value path.Path, lifeTime string) error {
	var keyOpt string
	if altRoot == IpfsNode.Identity.String() {
		keyOpt = "--key=self"
	} else {
		keyOpt = "--key=" + altRoot
	}
	lifeTimeOpt := "--lifetime=" + lifeTime
	args := []string{"name", "publish", keyOpt, lifeTimeOpt, value.String()}
	result, err := publish(args)
	if err != nil {
		return err
	}

	fmt.Printf("Published %s: %s\n", result.Name, result.Value)

	return nil
}

func publish(args []string) (*commands.IpnsEntry, error) {
	req, err := NewRequest(args)
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}
	req.Options["encoding"] = cmds.JSON
	req.Command.Type = commands.IpnsEntry{}
	buf := bytes.NewBuffer(nil)
	wc := writecloser{Writer: buf, Closer: nopCloser{}}
	rsp := cmds.NewWriterResponseEmitter(wc, req, cmds.Encoders[cmds.JSON])
	var env cmd.Context
	env.ConstructNode = func() (*core.IpfsNode, error) {
		return IpfsNode, nil
	}

	Root.Call(req, rsp, &env)

	var result IpnsRspEntry
	err = json.Unmarshal(buf.Bytes(), &result)

	if err != nil {
		log.Error(err.Error())
		return nil, err
	}

	return &result.Value, nil
}
