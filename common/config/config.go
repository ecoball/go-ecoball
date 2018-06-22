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
	"io/ioutil"
	"os"
	"path"

	"github.com/spf13/viper"

	"fmt"

	"github.com/ecoball/go-ecoball/common/utils"
)

//EcoBall version
var EcoVersion string

//listen port
var HttpLocalPort string = "20337"

const (
	HttpPort   = "http_port"
	Version    = "version"
	LogLevel   = "log_level"
	LogPath    = "log_path"
	PublicKey  = "pub_key"
	PrivateKey = "pri_key"
)

const (
	StringBlock    = "/Block"
	StringHeader   = "/Header"
	StringTxs      = "/Txs"
	StringContract = "/Contract"
	StringState    = "/State"
)

//set default value
var configDefault = `#toml configuration for aba
http_port = "20678"			# 
version = "1.0"				#
log_path = ""				# 
log_level = 5				#
pub_key = "1234567890"		#
pri_key = ""				#
`

type Config struct {
	FilePath string
}

func SetConfig() error {
	c := new(Config)
	c.FilePath = "./"
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

func CreateOrReadConfig() {
	if err := SetConfig(); err != nil {
		fmt.Errorf("%s\n", err)
		os.Exit(-1)
	}

	HttpLocalPort = viper.GetString(HttpPort)
	EcoVersion = viper.GetString(Version)
}
