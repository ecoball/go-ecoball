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
package util

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"time"
	"sync"
)

type WaitGroupWrapper struct {
	sync.WaitGroup
}

func (w *WaitGroupWrapper) Wrap(cb func()) {
	w.Add(1)
	go func() {
		cb()
		w.Done()
	}()
}

// Write data to a stream using a timeout.
func WriteWithTimeout(writer io.Writer, data []byte, timeout time.Duration) error {
	result := make(chan error, 1)
	go func(writer io.Writer, data []byte) {
		_, err := writer.Write(data)
		result <- err
	}(writer, data)

	select {
	case err := <-result:
		return err
	case <-time.After(timeout):
		select {
		case result <- errors.New("Timeout!"):
		default:
		}
		err := <-result
		return err
	}
}

// Read data from a stream using a timeout.
func ReadWithTimeout(reader io.Reader, n uint32, timeout time.Duration) ([]byte, error) {
	data := make([]byte, n)
	result := make(chan error, 1)
	go func(reader io.Reader) {
		_, err := io.ReadFull(reader, data)
		result <- err
	}(reader)

	select {
	case err := <-result:
		return data, err
	case <-time.After(timeout):
		select {
		case result <- errors.New("Timeout!"):
		default:
		}
		err := <-result
		return data, err
	}
}

// Read an unsigned 32-bit integer from a stream using a timeout.
func ReadUInt32WithTimeout(reader io.Reader, timeout time.Duration) (uint32, error) {
	var arr [4]byte
	data, err := ReadWithTimeout(reader, 4, timeout)
	if err != nil {
		return 0, err
	}
	copy(arr[:], data)
	n := DecodeBigEndianUInt32(arr)
	return n, nil
}

// Read a signed 64-bit integer from a stream using a timeout.
func ReadInt64WithTimeout(reader io.Reader, timeout time.Duration) (int64, error) {
	var arr [8]byte
	data, err := ReadWithTimeout(reader, 8, timeout)
	if err != nil {
		return 0, err
	}
	copy(arr[:], data)
	n := DecodeBigEndianInt64(arr)
	return n, nil
}

// Encode an unsigned 16-bit integer using big-endian byte order.
func EncodeBigEndianUInt16(n uint16) (data [2]byte) {
	var buf bytes.Buffer
	writer := bufio.NewWriter(&buf)
	binary.Write(writer, binary.BigEndian, &n)
	writer.Flush()
	copy(data[:], buf.Bytes())
	return
}

// Decode an unsigned 16-bit integer using big-endian byte order.
func DecodeBigEndianUInt16(data [2]byte) (n uint16) {
	reader := bytes.NewReader(data[:])
	binary.Read(reader, binary.BigEndian, &n)
	return
}

// Encode an unsigned 32-bit integer using big-endian byte order.
func EncodeBigEndianUInt32(n uint32) (data [4]byte) {
	var buf bytes.Buffer
	writer := bufio.NewWriter(&buf)
	binary.Write(writer, binary.BigEndian, &n)
	writer.Flush()
	copy(data[:], buf.Bytes())
	return
}

// Decode an unsigned 32-bit integer using big-endian byte order.
func DecodeBigEndianUInt32(data [4]byte) (n uint32) {
	reader := bytes.NewReader(data[:])
	binary.Read(reader, binary.BigEndian, &n)
	return
}

// Encode a signed 64-bit integer using big-endian byte order.
func EncodeBigEndianInt64(n int64) (data [8]byte) {
	var buf bytes.Buffer
	writer := bufio.NewWriter(&buf)
	binary.Write(writer, binary.BigEndian, &n)
	writer.Flush()
	copy(data[:], buf.Bytes())
	return
}

// Decode a signed 64-bit integer using big-endian byte order.
func DecodeBigEndianInt64(data [8]byte) (n int64) {
	reader := bytes.NewReader(data[:])
	binary.Read(reader, binary.BigEndian, &n)
	return
}
