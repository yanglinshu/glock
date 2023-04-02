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

	"github.com/yanglinshu/glock/internal/util"
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

		dataToSign := fmt.Sprintf("%x\n", txCopy)

		// Sign the transaction with the private key
		r, s, err := ecdsa.Sign(rand.Reader, &privKey, []byte(dataToSign))
		if err != nil {
			return err
		}

		// Combine the r and s into a single signature
		signature := append(r.Bytes(), s.Bytes()...)
		tx.Vin[inID].Signature = signature
		txCopy.Vin[inID].PublicKey = nil
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
func (tx *Transaction) Serialize() ([]byte, error) {
	encoded, err := util.GobEncode(tx)
	if err != nil {
		return nil, err
	}

	return encoded, nil
}

// Hash returns the hash of the transaction.
func (tx *Transaction) Hash() ([]byte, error) {
	var hash [32]byte

	txCopy := *tx
	txCopy.ID = []byte{}

	sl, err := txCopy.Serialize()
	if err != nil {
		return nil, err
	}

	hash = sha256.Sum256(sl)

	return hash[:], nil
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

		dataToVerify := fmt.Sprintf("%x\n", txCopy)

		// Verify the signature
		rawPubKey := ecdsa.PublicKey{Curve: curve, X: &x, Y: &y}
		if !ecdsa.Verify(&rawPubKey, []byte(dataToVerify), &r, &s) {
			return false
		}
		txCopy.Vin[inID].PublicKey = nil
	}

	return true
}

// NewCoinbaseTX creates a new coinbase transaction. The transaction will have no inputs, and will
// have an output that will be given to the miner. The value of the output will be the reward for
// mining the block.
func NewCoinbaseTX(to, data string) (*Transaction, error) {
	if data == "" {
		randData := make([]byte, 20)
		_, err := rand.Read(randData)
		if err != nil {
			return nil, err
		}

		data = fmt.Sprintf("%x", randData)
	}

	txin := TXInput{[]byte{}, -1, nil, []byte(data)}
	txout := NewTXOutput(subsidy, to)

	tx := Transaction{nil, []TXInput{txin}, []TXOutput{*txout}}

	var err error = nil
	tx.ID, err = tx.Hash()
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
	encoded, err := util.GobEncode(tx)
	if err != nil {
		return err
	}

	hash := sha256.Sum256(encoded)
	tx.ID = hash[:]
	return nil
}

// DeserializeTransaction deserializes a transaction.
func DeserializeTransaction(data []byte) (Transaction, error) {
	var transaction Transaction

	decoder := gob.NewDecoder(bytes.NewReader(data))
	err := decoder.Decode(&transaction)
	if err != nil {
		return Transaction{}, err
	}

	return transaction, nil
}
