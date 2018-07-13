// Copyright 2018 The go-ecoball Authors
// This file is part of the go-ecoball library.
//
// The go-ecoball library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ecoball library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ecoball library. If not, see <http://www.gnu.org/licenses/>.

package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"

	"github.com/spf13/viper"

	"github.com/ecoball/go-ecoball/account"
	"github.com/ecoball/go-ecoball/common"
	"github.com/ecoball/go-ecoball/common/utils"
)

// peer list
var PeerList []string
var PeerIndex []string

const (
	StringBlock     = "/Block"
	StringHeader    = "/Header"
	StringTxs       = "/Txs"
	StringContract  = "/Contract"
	StringState     = "/State"
	StringConsensus = "/Consensus"
	ConsensusPro = "consensus_algorithm"
	ListPeers  = "peer_list"
	IndexPeers = "peer_index"
)

var configDefault = `#toml configuration for EcoBall system
http_port = "20678"          # client http port
version = "1.0"              # system version
log_dir = "/tmp/Log/"        # log file location
output_to_terminal = "true"	 	
log_level = 1                # debug level	
consensus_algorithm = "SOLO" # can set as SOLO, DPOS, ababft
root_privkey = "0x33a0330cd18912c215c9b1125fab59e9a5ebfb62f0223bbea0c6c5f95e30b1c6"
root_pubkey = "0x0463613734b23e5dd247b7147b63369bf8f5332f894e600f7357f3cfd56886f75544fd095eb94dac8401e4986de5ea620f5a774feb71243e95b4dd6b83ca49910c"
peer_list = [ "120202c924ed1a67fd1719020ce599d723d09d48362376836e04b0be72dfe825e24d810000", 
              "120202935fb8d28b70706de6014a937402a30ae74a56987ed951abbe1ac9eeda56f0160000" ]
peer_index = [ "1", "2" ]
`

var (
	HttpLocalPort      string
	EcoVersion         string
	LogDir             string
	OutputToTerminal   bool
	LogLevel           int
	ConsensusAlgorithm string
	RootPrivkey        string
	RootPubkey         string
	Root               account.Account
)

type Config struct {
	FilePath string
}

func SetConfig() error {
	c := new(Config)
	if err := c.CreateConfigFile(); err != nil {
		return err
	}
	return c.InitConfig()
}

func (c *Config) CreateConfigFile() error {
	if "" == c.FilePath {
		c.FilePath, _ = defaultPath()
	}
	var dirPath string
	var filePath string
	if "" == path.Ext(c.FilePath) {
		dirPath = c.FilePath
		filePath = path.Join(c.FilePath, "ecoball.toml")
	} else {
		dirPath = c.FilePath
		filePath = path.Join(c.FilePath, "ecoball.toml")
		//filePath = path.Dir(c.FilePath)
	}
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		if err := os.MkdirAll(dirPath, 0700); err != nil {
			fmt.Println("could not create directory:", dirPath, err)
			return err
		}
	}
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		if err := ioutil.WriteFile(filePath, []byte(configDefault), 0644); err != nil {
			fmt.Println("write file err:", err)
			return err
		}
	}
	return nil
}

func defaultPath() (string, error) {
	return utils.DirHome()
}

func (c *Config) InitConfig() error {
	viper.SetConfigName("ecoball")
	viper.AddConfigPath(c.FilePath)
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	} else {
		fmt.Println("can't load config file:", err)
		return err
	}
	return nil
}

func init() {
	if err := SetConfig(); err != nil {
		fmt.Println("init config failed: ", err)
		os.Exit(-1)
	}

	initVariable()
}

func initVariable() {
	HttpLocalPort = viper.GetString("http_port")
	EcoVersion = viper.GetString("version")
	LogDir = viper.GetString("log_dir")
	OutputToTerminal = viper.GetBool("output_to_terminal")
	LogLevel = viper.GetInt("log_level")
	ConsensusAlgorithm = viper.GetString("consensus_algorithm")
	RootPrivkey = viper.GetString("root_privkey")
	RootPubkey = viper.GetString("root_pubkey")
	Root = account.Account{PrivateKey: common.FromHex(RootPrivkey), PublicKey: common.FromHex(RootPubkey), Alg: 0}
	PeerList = viper.GetStringSlice(ListPeers)
	PeerIndex = viper.GetStringSlice(IndexPeers)
}
