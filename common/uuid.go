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
package common

import (
	"encoding/binary"
	"errors"
	"fmt"
	"regexp"
	"strings"
)

type AccountName uint64
var base32Alphabet = []byte(".12345abcdefghijklmnopqrstuvwxyz")

/**
 *  @brief get a account index name
 */
func (a AccountName) String() string {
	return IndexToName(a)
}
/**
 *  @brief convert a string name to uint64 type
 *  @param name - the account name string
 */
func NameToIndex(name string) AccountName {
	var index uint64
	var i uint32
	sLen := uint32(len(name))
	for ; i <= 12; i++ {
		var c uint64
		if i < sLen {
			c = uint64(charToSymbol(name[i]))
		}
		if i < 12 {
			c &= 0x1f
			c <<= 64 - 5*(i+1)
		} else {
			c &= 0x0f
		}
		index |= c
	}
	return AccountName(index)
}
/**
 *  @brief convert a uint64 name to string type
 *  @param index - the account name index
 */
func IndexToName(index AccountName) string {
	a := []byte{'.', '.', '.', '.', '.', '.', '.', '.', '.', '.', '.', '.', '.'}
	tmp := uint64(index)
	i := uint32(0)
	for ; i <= 12; i++ {
		bit := 0x1f
		if i == 0 {
			bit = 0x0f
		}
		c := base32Alphabet[tmp&uint64(bit)]
		a[12-i] = c

		shift := uint(5)
		if i == 0 {
			shift = 4
		}
		tmp >>= shift
	}

	return strings.TrimRight(string(a), ".")
}
/**
 *  @brief convert a uint64 name to byte type for mpt trie's key
 *  @param index - the account name index
 */
func IndexToBytes(index AccountName) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, uint64(index))
	return b
}
/**
 *  @brief convert name's byte to uint64 type
 *  @param data - the account name bytes
 */
func IndexSetBytes(data []byte) AccountName {
	index := binary.BigEndian.Uint64(data)
	return AccountName(index)
}
/**
 *  @brief check a name string validity, must be 1-5, a-z and ., the len of name must be 1-12
 *  @param name - the account name
 */
func AccountNameCheck(name string) error {
	reg := `^[.1-5a-z]{1,12}$`
	rgx := regexp.MustCompile(reg)
	if !rgx.MatchString(name) {
		e := fmt.Sprintf("Invalid name\n" +
			" Name should be less than 13 characters and only contains the following symbol .12345abcdefghijklmnopqrstuvwxyz")
		return errors.New(e)
	}
	return nil
}

func TokenNameCheck(name string) error {
	reg := `^[A-Z]{1,12}$`
	rgx := regexp.MustCompile(reg)
	if !rgx.MatchString(name) {
		e := fmt.Sprintf("Invalid name\n" +
			" Name should be less than 13 characters and only contains the following symbol 12345abcdefghijklmnopqrstuvwxyz")
		return errors.New(e)
	}
	return nil
}

func charToSymbol(c byte) byte {
	if c >= 'a' && c <= 'z' {
		return c - 'a' + 6
	}
	if c >= '1' && c <= '5' {
		return c - '1' + 1
	}
	return 0
}
