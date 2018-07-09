package common

import (
	"math/big"
	"encoding/binary"
	"github.com/btcsuite/btcutil/base58"
	"regexp"
	"fmt"
	"errors"
)

func NameToIndex(name string) uint64 {
	base := base58.Encode([]byte(name))
	index := new(big.Int).SetBytes([]byte(base)).Uint64()
	return index
}

func IndexToName(index uint64) string {
	u := new(big.Int).SetUint64(index)
	name := base58.Decode(string(u.Bytes()))
	return string(name)
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
	reg := `^[1-5a-z]{1,12}$`
	rgx := regexp.MustCompile(reg)
	if !rgx.MatchString(name) {
		e := fmt.Sprintf("Invalid name\n" +
			" Name should be less than 13 characters and only contains the following symbol 12345abcdefghijklmnopqrstuvwxyz")
		return errors.New(e)
	}
	return nil
}
