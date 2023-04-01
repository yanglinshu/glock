package transaction

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"fmt"
)

// subsidy is the amount of coins given to the miner as a reward for mining a block.
const subsidy = 10

// Transaction is a struct that contains the ID, inputs and outputs of a transaction. The Id is a
// unique identifier for the transaction. The inputs must be the outputs of previous transactions.
// The outputs will be the new outputs of the transaction.
// When a miner creates a new block, it will include a coinbase transaction. This transaction will
// have no inputs, and will have an output that will be given to the miner. The value of the output
// will be the reward for mining the block.
type Transaction struct {
	ID   []byte     // ID is the hash of the transaction
	Vin  []TXInput  // Vin is the inputs of the transaction
	Vout []TXOutput // Vout is the outputs of the transaction
}

// TXInput represents a transaction input. It contains the ID of the transaction that contains the
// output, the index of the output in the transaction, and the signature of the input. The signature
// is used to verify that the owner of the output is the one spending it.
type TXInput struct {
	Txid      []byte // Txid is the ID of the transaction that contains the output
	Vout      int    // Vout is the index of the output in the transaction
	ScriptSig string // ScriptSig is the signature of the input
}

// TXOutput represents a transaction output. It contains the value of the output and the public key
// of the recipient. In glock, the public key will be a simple string, rather than a smart contract.
// Note that the value of the output cannot be used partially. If the value is greater than the amount
// needed, the remaining value will be returned to the sender as a new output.
type TXOutput struct {
	Value     int    // Value is the amount of coins in the output
	ScriptSig string // ScriptSig is the signature of the output
}

// NewCoinbaseTX creates a new coinbase transaction. The transaction will have no inputs, and will
// have an output that will be given to the miner. The value of the output will be the reward for
// mining the block.
func NewCoinbaseTX(to, data string) (*Transaction, error) {
	if data == "" {
		data = fmt.Sprintf("Reward to '%s'", to)
	}

	txin := TXInput{[]byte{}, -1, data}
	txout := TXOutput{subsidy, to}
	tx := Transaction{nil, []TXInput{txin}, []TXOutput{txout}}

	err := tx.SetID()
	if err != nil {
		return nil, err
	}

	return &tx, nil
}

// IsCoinbase checks whether the transaction is a coinbase transaction.
func (tx *Transaction) IsCoinbase() bool {
	return len(tx.Vin) == 1 && len(tx.Vin[0].Txid) == 0 && tx.Vin[0].Vout == -1
}

// SetID sets the ID of the transaction to the hash of the transaction.
func (tx *Transaction) SetID() error {
	var encoded bytes.Buffer
	var hash [32]byte

	enc := gob.NewEncoder(&encoded)
	err := enc.Encode(tx)
	if err != nil {
		return err
	}

	hash = sha256.Sum256(encoded.Bytes())
	tx.ID = hash[:]
	return nil
}

// CanUnlockOutputWith checks whether the address provided can unlock the output with the provided
// ID.
func (in *TXInput) CanUnlockOutputWith(unlockingData string) bool {
	return in.ScriptSig == unlockingData
}

// CanBeUnlockedWith checks whether the address provided can unlock the output.
func (out *TXOutput) CanBeUnlockedWith(unlockingData string) bool {
	return out.ScriptSig == unlockingData
}
