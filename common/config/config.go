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
	ConsensusPro    = "consensus_algorithm"
	ListPeers       = "peer_list"
	IndexPeers      = "peer_index"
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

worker1_privkey = "0xc3e2cbed03aacc62d8f32045013364ea493f6d24e84f26bcef4edc2e9d260c0e"
worker1_pubkey = "0x04e0c1852b110d1586bf6202abf6e519cc4161d00c3780c04cfde80fd66748cc189b6b0e2771baeb28189ec42a363461357422bf76b1e0724fc63fc97daf52769f"

worker2_privkey = "0x5238ede4f91f6c4f5f1f195cbf674e08cb6a18ae351e474b8927db82d3e5ecf5"
worker2_pubkey = "0x049e78e40b0dcca842b94cb2586d47ecc61888b52dce958b41aa38613c80f6607ee1de23eebb912431eccfe0fea81f8a38792ffecee38c490dde846c646ce1f0ee"

worker3_privkey = "0x105cb8f936eec87d35e42fc0f656ab4b7fc9a007cbf4554f829c44e528df6ce4"
worker3_pubkey = "0x0481bce0ad10bd3d8cdfd089ac5534379149ca5c3cdab28b5063f707d20f3a4a51f192ef7933e91e3fd0a8ea21d8dd735407780937c3c71753b486956fd481349f"

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
	Worker1            account.Account
	Worker2            account.Account
	Worker3            account.Account
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
	Worker1 = account.Account{PrivateKey: common.FromHex(viper.GetString("worker1_privkey")), PublicKey: common.FromHex(viper.GetString("worker1_pubkey")), Alg: 0}
	Worker2 = account.Account{PrivateKey: common.FromHex(viper.GetString("worker2_privkey")), PublicKey: common.FromHex(viper.GetString("worker2_pubkey")), Alg: 0}
	Worker3 = account.Account{PrivateKey: common.FromHex(viper.GetString("worker3_privkey")), PublicKey: common.FromHex(viper.GetString("worker3_pubkey")), Alg: 0}
	PeerList = viper.GetStringSlice(ListPeers)
	PeerIndex = viper.GetStringSlice(IndexPeers)
}
