package blockchain

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"time"

	"github.com/yanglinshu/glock/internal/transaction"
)

// Block represents a block in the blockchain. It contains the header and the transactions.
type Block struct {
	Timestamp     int64                      // Time of creation of the block
	Transactions  []*transaction.Transaction // Transactions in the block
	PrevBlockHash []byte                     // Hash of the previous block
	Hash          []byte                     // Hash of the current block
	Nonce         int                        // Nonce is the number of times the hash of the block is calculated
}

// NewBlock creates and returns a pointer to a Block.
func NewBlock(transactions []*transaction.Transaction, prevBlockHash []byte) *Block {
	block := &Block{time.Now().Unix(), transactions, prevBlockHash, []byte{}, 0}
	pow := NewProofOfWork(block)
	nonce, hash := pow.Run()

	block.Hash = hash[:]
	block.Nonce = nonce

	return block
}

// NewGenesisBlock creates and returns a pointer to a genesis block.
func NewGenesisBlock(coinbase *transaction.Transaction) *Block {
	return NewBlock([]*transaction.Transaction{coinbase}, []byte{})
}

// Serialize serializes the block into a byte slice using the Gob encoding.
func (b *Block) Serialize() ([]byte, error) {
	var result bytes.Buffer
	encoder := gob.NewEncoder(&result)

	err := encoder.Encode(b)

	if err != nil {
		return nil, err
	}

	return result.Bytes(), nil
}

// HashTransactions returns the hash of the transactions in the block.
// In Bitcoin, the transactions are hashed in the Merkle tree, allowing for efficient verification
// of the transactions in the block.
func (b *Block) HashTransactions() []byte {
	var txHashes [][]byte
	var txHash [32]byte

	for _, tx := range b.Transactions {
		txHashes = append(txHashes, tx.ID)
	}

	txHash = sha256.Sum256(bytes.Join(txHashes, []byte{}))

	return txHash[:]
}

// DeserializeBlock deserializes a byte slice into a block using the Gob encoding.
func DeserializeBlock(d []byte) (*Block, error) {
	var block Block

	decoder := gob.NewDecoder(bytes.NewReader(d))
	err := decoder.Decode(&block)

	if err != nil {
		return nil, err
	}

	return &block, nil
}
