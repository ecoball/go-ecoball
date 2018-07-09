package common

import (
	"encoding/binary"
	"regexp"
	"fmt"
	"errors"
	"strings"
)

func NameToIndex(name string) (index uint64) {
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
	return
}
var base32Alphabet = []byte(".12345abcdefghijklmnopqrstuvwxyz")
func IndexToName(index uint64) string {
	a := []byte{'.', '.', '.', '.', '.', '.', '.', '.', '.', '.', '.', '.', '.'}
	tmp := index
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


func IndexToBytes(index uint64) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, uint64(index))
	return b
}

func IndexSetBytes(data []byte) uint64 {
	index := binary.BigEndian.Uint64(data)
	return index
}

func AccountNameCheck(name string) error {
	reg := `^[.1-5a-z]{1,12}$`
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
