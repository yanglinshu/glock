package transaction

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"
)

// subsidy is the amount of coins given to the miner as a reward for mining a block.
const subsidy = 10

// TXInput represents a transaction input. It contains the ID of the transaction that contains the
// output, the index of the output in the transaction, and the signature of the input. The signature
// is used to verify that the owner of the output is the one spending it.
type TXInput struct {
	Txid      []byte // Txid is the ID of the transaction that contains the output
	Vout      int    // Vout is the index of the output in the transaction
	Signature []byte // Signature is the signature of the input
	PublicKey []byte // PublicKey is the public key of the owner of the output
}

// UsesKey checks whether the address is the owner of the output.
func (in *TXInput) UsesKey(pubKeyHash []byte) (bool, error) {
	lockingHash, err := HashPubKey(in.PublicKey)
	if err != nil {
		return false, err
	}

	return bytes.Equal(lockingHash, pubKeyHash), nil
}

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

// Sign signs each input of the transaction.
func (tx *Transaction) Sign(privKey ecdsa.PrivateKey, prevTXs map[string]Transaction) error {
	if tx.IsCoinbase() {
		return nil
	}

	// Sign the inputs, one at a time
	txCopy := tx.TrimmedCopy()
	for inID, vin := range txCopy.Vin {
		prevTx := prevTXs[hex.EncodeToString(vin.Txid)]
		txCopy.Vin[inID].Signature = nil
		txCopy.Vin[inID].PublicKey = prevTx.Vout[vin.Vout].PublicKeyHash
		txCopy.ID = txCopy.Hash()
		txCopy.Vin[inID].PublicKey = nil

		// Sign the transaction with the private key
		r, s, err := ecdsa.Sign(rand.Reader, &privKey, txCopy.ID)
		if err != nil {
			return err
		}

		// Combine the r and s into a single signature
		signature := append(r.Bytes(), s.Bytes()...)
		tx.Vin[inID].Signature = signature
	}

	return nil
}

// TrimmedCopy creates a trimmed copy of the transaction. The copy will have no signature, and the
// public key of the inputs will be replaced by the hash of the public key.
func (tx *Transaction) TrimmedCopy() Transaction {
	var inputs []TXInput
	var outputs []TXOutput

	for _, vin := range tx.Vin {
		inputs = append(inputs, TXInput{vin.Txid, vin.Vout, nil, nil})
	}

	for _, vout := range tx.Vout {
		outputs = append(outputs, TXOutput{vout.Value, vout.PublicKeyHash})
	}

	txCopy := Transaction{tx.ID, inputs, outputs}

	return txCopy
}

// Serialize serializes the transaction using the gob package.
func (tx *Transaction) Serialize() []byte {
	var encoded bytes.Buffer

	enc := gob.NewEncoder(&encoded)
	err := enc.Encode(tx)
	if err != nil {
		panic(err)
	}

	return encoded.Bytes()
}

// Hash returns the hash of the transaction.
func (tx *Transaction) Hash() []byte {
	var hash [32]byte

	txCopy := *tx
	txCopy.ID = []byte{}

	hash = sha256.Sum256(txCopy.Serialize())

	return hash[:]
}

// String returns a human-readable representation of the transaction.
func (tx *Transaction) String() string {
	var lines []string

	lines = append(lines, fmt.Sprintf("--- Transaction %x:", tx.ID))

	for i, input := range tx.Vin {
		lines = append(lines, fmt.Sprintf("     Input %d:", i))
		lines = append(lines, fmt.Sprintf("       TXID:      %x", input.Txid))
		lines = append(lines, fmt.Sprintf("       Out:       %d", input.Vout))
		lines = append(lines, fmt.Sprintf("       Signature: %x", input.Signature))
		lines = append(lines, fmt.Sprintf("       PubKey:    %x", input.PublicKey))
	}

	for i, output := range tx.Vout {
		lines = append(lines, fmt.Sprintf("     Output %d:", i))
		lines = append(lines, fmt.Sprintf("       Value:  %d", output.Value))
		lines = append(lines, fmt.Sprintf("       Script: %x", output.PublicKeyHash))
	}

	return strings.Join(lines, "\n")
}

// Verify verifies the signatures of the transaction.
func (tx *Transaction) Verify(prevTXs map[string]Transaction) bool {
	txCopy := tx.TrimmedCopy()
	curve := elliptic.P256()

	for inID, vin := range tx.Vin {
		// Get the public key from the previous transaction
		prevTx := prevTXs[hex.EncodeToString(vin.Txid)]
		txCopy.Vin[inID].Signature = nil
		txCopy.Vin[inID].PublicKey = prevTx.Vout[vin.Vout].PublicKeyHash
		txCopy.ID = txCopy.Hash()
		txCopy.Vin[inID].PublicKey = nil

		// Extract the real signature and the real public key from the transaction
		r := big.Int{}
		s := big.Int{}
		sigLen := len(vin.Signature)
		r.SetBytes(vin.Signature[:(sigLen / 2)])
		s.SetBytes(vin.Signature[(sigLen / 2):])

		x := big.Int{}
		y := big.Int{}
		keyLen := len(vin.PublicKey)
		x.SetBytes(vin.PublicKey[:(keyLen / 2)])
		y.SetBytes(vin.PublicKey[(keyLen / 2):])

		// Verify the signature
		rawPubKey := ecdsa.PublicKey{Curve: curve, X: &x, Y: &y}
		if !ecdsa.Verify(&rawPubKey, txCopy.ID, &r, &s) {
			return false
		}
	}

	return true
}

// NewCoinbaseTX creates a new coinbase transaction. The transaction will have no inputs, and will
// have an output that will be given to the miner. The value of the output will be the reward for
// mining the block.
func NewCoinbaseTX(to, data string) *Transaction {
	if data == "" {
		data = fmt.Sprintf("Reward to '%s'", to)
	}

	txin := TXInput{[]byte{}, -1, nil, []byte(data)}
	txout := NewTXOutput(subsidy, to)

	tx := Transaction{nil, []TXInput{txin}, []TXOutput{*txout}}
	tx.ID = tx.Hash()

	return &tx
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
