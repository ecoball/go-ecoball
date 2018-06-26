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

	"github.com/ecoball/go-ecoball/common/utils"
)

const (
	StringBlock     = "/Block"
	StringHeader    = "/Header"
	StringTxs       = "/Txs"
	StringContract  = "/Contract"
	StringState     = "/State"
	StringConsensus = "/Consensus"
)

//TODO
const (
	ConsensusAlgorithm = "DPOS"
)

const (
	configDefault = `#toml configuration for aba
http_port = "20678"			 
version = "1.0"	
log_dir = "./Log/"
output_to_terminal = "true"		
log_level = 1				 								
`
)

var (
	HttpLocalPort    string
	EcoVersion       string
	LogDir           string
	OutputToTerminal bool
	LogLevel         int
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
		filePath = path.Join(c.FilePath, "aba.toml")
	} else {
		dirPath = c.FilePath
		filePath = path.Join(c.FilePath, "aba.toml")
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
	viper.SetConfigName("aba")
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
}
