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
	"errors"
	cmds "gx/ipfs/QmSKYWC84fqkKB54Te5JMcov2MBVzucXaRGxFqByzzCbHe/go-ipfs-cmds"
	"fmt"
	cmd "github.com/ipfs/go-ipfs/commands"
	"github.com/ipfs/go-ipfs/core"
	"github.com/ipfs/go-ipfs/core/coreunix"
	"bytes"
	"io"
	"encoding/json"
	"path"
	"crypto/sha256"
	"encoding/hex"
	"io/ioutil"
	"os"
)

var addErr = errors.New(`Add directory failed`)

type writecloser struct {
	io.Writer
	io.Closer
}
type nopCloser struct{}

//TODO move to node
// The path to the openbazaar repo in the file system
var RepoPath string

func SetRepoPath(path string)  {
	RepoPath = path
}


func (c nopCloser) Close() error { return nil }

func IpfsAdd(data []byte) (string, error) {
	h := sha256.Sum256(data)
	tmpPath := path.Join(RepoPath, hex.EncodeToString(h[:])+".data")
	err := ioutil.WriteFile(tmpPath, data, os.ModePerm)
	if err != nil {
		return "", nil
	}
	hash, err := AddFile(tmpPath)
	if err != nil {
		return "", nil
	}
	err = os.Remove(tmpPath)
	if err != nil {
		return "", nil
	}
	//TODO publish to our peer
	//PublishAltRoot()
	return hash, nil
}

func AddFile(filePath string) (string, error) {
	args := []string{"add", filePath}
	req, err := NewRequest(args)
	if err != nil {
		log.Error(err.Error())
		return "", nil
	}
	req.Options["encoding"] = cmds.JSON
	req.Command.Type = coreunix.AddedObject{}
	buf := bytes.NewBuffer(nil)
	wc := writecloser{Writer: buf, Closer: nopCloser{}}
	rsp := cmds.NewWriterResponseEmitter(wc, req, cmds.Encoders[cmds.JSON])
	var env cmd.Context
	env.ConstructNode = func() (*core.IpfsNode, error) {
		return IpfsNode, nil
	}
	Root.Call(req, rsp, &env)
	var result coreunix.AddedObject
	err = json.Unmarshal(buf.Bytes(), &result)
	if err != nil {
		log.Error(err.Error())
		return "", nil
	}
	fmt.Printf("added %s %s\n", result.Name, result.Hash)
	return result.Hash, nil
}

func AddDirectory(fpath string) (string, error) {
	_, root := path.Split(fpath)
	args := []string{"add", "-r", fpath}
	req, err := NewRequest(args)
	if err != nil {
		log.Error(err.Error())
		return "", nil
	}
	req.Options["encoding"] = cmds.JSON
	req.Command.Type = []coreunix.AddedObject{}
	buf := bytes.NewBuffer(nil)
	wc := writecloser{Writer: buf, Closer: nopCloser{}}
	rsp := cmds.NewWriterResponseEmitter(wc, req, cmds.Encoders[cmds.JSON])
	var env cmd.Context
	env.ConstructNode = func() (*core.IpfsNode, error) {
		return IpfsNode, nil
	}
	Root.Call(req, rsp, &env)
	var result []coreunix.AddedObject
	err = json.Unmarshal(buf.Bytes(), result)
	if err != nil {
		log.Error(err.Error())
		return "", nil
	}
	var rootHash string
	for _, v := range result {
		if v.Name == root {
			rootHash = v.Hash
			break
		}
	}
	fmt.Printf("added %s %s\n", root, rootHash)
	return rootHash, nil
}