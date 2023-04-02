package util

import (
	"bytes"
	"encoding/gob"
	"fmt"
)

// IntToHex converts an integer to a hexadecimal byte array.
func IntToHex(n int64) []byte {
	return []byte(fmt.Sprintf("%x", n))
}

// ReverseBytes reverses a byte array
func ReverseBytes(data []byte) {
	for i, j := 0, len(data)-1; i < j; i, j = i+1, j-1 {
		data[i], data[j] = data[j], data[i]
	}
}

// gobEncode encodes a data structure into a byte array.
func GobEncode(data any) ([]byte, error) {
	var buff bytes.Buffer

	enc := gob.NewEncoder(&buff)
	err := enc.Encode(data)
	if err != nil {
		return nil, err
	}

	return buff.Bytes(), nil
}
