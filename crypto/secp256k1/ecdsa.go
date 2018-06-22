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
package secp256k1

import (
	"crypto/ecdsa"
	"crypto/elliptic"

	"encoding/hex"
	"errors"
	"math/big"

	"crypto/rand"

	"github.com/ecoball/go-ecoball/crypto/secp256k1/bitelliptic"
)

func S256() elliptic.Curve {
	return bitelliptic.S256()
}

//私钥生成
func NewECDSAPrivateKey() (*ecdsa.PrivateKey, error) {
	var priv *ecdsa.PrivateKey

	for {
		privKey, err := ecdsa.GenerateKey(S256(), rand.Reader)
		if err != nil {
			return nil, err
		}
		privData, err := FromECDSA(privKey)
		if err != nil {
			return nil, err
		}
		priv = privKey
		if SeckeyVerify(privData) {
			break
		}
	}
	return priv, nil
}
//导出私钥
func FromECDSA(priv *ecdsa.PrivateKey) ([]byte, error) {
	if priv == nil {
		return nil, errors.New("ecdsa: no private key")
	}
	return paddedBigBytes(priv.D, priv.Params().BitSize/8), nil
}
//导出公钥
func FromECDSAPub(pub *ecdsa.PublicKey) ([]byte, error) {
	if pub == nil || pub.X == nil || pub.Y == nil {
		return nil, errors.New("ecdsa: no public key")
	}
	return elliptic.Marshal(S256(), pub.X, pub.Y), nil
}


func HexToECDSA(hexkey string) (*ecdsa.PrivateKey, error) {
	b, err := hex.DecodeString(hexkey)
	if err != nil {
		return nil, err
	}
	return ToECDSA(b)
}

//
func ToECDSA(d []byte) (*ecdsa.PrivateKey, error) {
	priv := new(ecdsa.PrivateKey)
	priv.PublicKey.Curve = S256()
	priv.D = new(big.Int).SetBytes(d)
	priv.PublicKey.X, priv.PublicKey.Y = priv.PublicKey.Curve.ScalarBaseMult(d)
	return priv, nil
}

//
func ToECDSAPub(pub []byte) (*ecdsa.PublicKey, error) {
	if len(pub) == 0 {
		return nil, errors.New("ecdsa: no public key")
	}
	x, y := elliptic.Unmarshal(S256(), pub)
	return &ecdsa.PublicKey{Curve: S256(), X: x, Y: y}, nil
}

func zeroKey(k *ecdsa.PrivateKey) {
	b := k.D.Bits()
	for i := range b {
		b[i] = 0
	}
}

func paddedBigBytes(bigint *big.Int, n int) []byte {
	if bigint.BitLen()/8 >= n {
		return bigint.Bytes()
	}
	ret := make([]byte, n)
	readBits(bigint, ret)
	return ret
}

const (
	wordBits = 32 << (uint64(^big.Word(0)) >> 63)
	wordBytes = wordBits / 8
)

func readBits(bigint *big.Int, buf []byte) {
	i := len(buf)
	for _, d := range bigint.Bits() {
		for j := 0; j < wordBytes && i > 0; j++ {
			i--
			buf[i] = byte(d)
			d >>= 8
		}
	}
}
