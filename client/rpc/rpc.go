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

package rpc

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/ecoball/go-ecoball/client/common"
	innerCommon "github.com/ecoball/go-ecoball/http/common"
)

// RPC call
func Call(method string, params []interface{}) (map[string]interface{}, error) {

	data, err := json.Marshal(map[string]interface{}{
		"method": method,
		"params": params,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Marshal JSON request: %v\n", err)
		return nil, err
	}

	resp, err := http.Post(common.RpcAddress(), "application/json", strings.NewReader(string(data)))
	if err != nil {
		fmt.Fprintf(os.Stderr, "POST request: %v\n", err)
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Fprintf(os.Stderr, "GET response: %v\n", err)
		return nil, err
	}

	result := make(map[string]interface{})
	err = json.Unmarshal(body, &result)
	if err != nil {
		fmt.Println("Unmarshal JSON failed")
		return nil, err
	}

	return result, nil
}

//rpc result
func EchoResult(resp map[string]interface{}) error {
	var (
		errorCode int64
		desc      string
		invalid   bool = false
	)

	if v, ok := resp["errorCode"].(float64); ok {
		errorCode = int64(v)
	} else {
		invalid = true
	}

	if v, ok := resp["desc"].(string); ok {
		desc = v
	} else {
		invalid = true
	}

	if invalid {
		fmt.Println("errorCode or desc of respone is wrong!")
		return errors.New("errorCode or desc of respone is wrong!")
	} else if errorCode != int64(innerCommon.SUCCESS) {
		fmt.Println("failed: ", desc)
		return errors.New(desc)
	}

	//success
	switch resp["result"].(type) {
	case map[string]interface{}:

	case string:
		fmt.Println(resp["result"].(string))
	}

	return nil
}
