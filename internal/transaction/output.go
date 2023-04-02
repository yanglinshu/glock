package transaction

import (
	"bytes"
	"encoding/gob"
)

// TXOutput represents a transaction output. It contains the value of the output and the public key
// of the recipient. In glock, the public key will be a simple string, rather than a smart contract.
// Note that the value of the output cannot be used partially. If the value is greater than the amount
// needed, the remaining value will be returned to the sender as a new output.
type TXOutput struct {
	Value         int    // Value is the amount of coins in the output
	PublicKeyHash []byte // PublicKeyHash is the hash of the public key of the recipient
}

// NewTXOutput creates and returns a TXOutput.
func NewTXOutput(value int, address string) *TXOutput {
	txo := &TXOutput{value, nil}
	txo.Lock([]byte(address))

	return txo
}

// Lock signs the output.
func (out *TXOutput) Lock(address []byte) {
	pubKeyHash := Base58Decode(address)

	pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-4]
	out.PublicKeyHash = pubKeyHash
}

// IsLockedWithKey checks whether the address is the owner of the output.
func (out *TXOutput) IsLockedWithKey(pubKeyHash []byte) bool {
	return bytes.Equal(out.PublicKeyHash, pubKeyHash)
}

// TXOutputs represents a list of transaction outputs.
type TXOutputs struct {
	Outputs []TXOutput
}

// Serialize serializes the transaction outputs.
func (outs TXOutputs) Serialize() ([]byte, error) {
	var buff bytes.Buffer

	enc := gob.NewEncoder(&buff)
	err := enc.Encode(outs)
	if err != nil {
		return nil, err
	}

	return buff.Bytes(), nil
}

// DeserializeOutputs deserializes the transaction outputs.
func DeserializeOutputs(data []byte) (TXOutputs, error) {
	var outputs TXOutputs

	dec := gob.NewDecoder(bytes.NewReader(data))
	err := dec.Decode(&outputs)
	if err != nil {
		return TXOutputs{}, err
	}

	return outputs, nil
}
