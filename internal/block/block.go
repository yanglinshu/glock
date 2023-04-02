package block

import (
	"bytes"
	"encoding/gob"
	"time"

	"github.com/yanglinshu/glock/internal/transaction"
	"github.com/yanglinshu/glock/internal/util"
)

// Block represents a block in the blockchain. It contains the header and the transactions.
type Block struct {
	Timestamp     int64                      // Time of creation of the block
	Transactions  []*transaction.Transaction // Transactions in the block
	PrevBlockHash []byte                     // Hash of the previous block
	Hash          []byte                     // Hash of the current block
	Nonce         int                        // Nonce is the number of times the hash of the block is calculated
	Height        int                        // Height of the block in the blockchain
}

// NewBlock creates and returns a pointer to a Block.
func NewBlock(transactions []*transaction.Transaction, prevBlockHash []byte, height int) *Block {
	block := &Block{time.Now().Unix(), transactions, prevBlockHash, []byte{}, 0, height}
	pow := NewProofOfWork(block)
	nonce, hash := pow.Run()

	block.Hash = hash[:]
	block.Nonce = nonce

	return block
}

// NewGenesisBlock creates and returns a pointer to a genesis block.
func NewGenesisBlock(coinbase *transaction.Transaction) *Block {
	return NewBlock([]*transaction.Transaction{coinbase}, []byte{}, 0)
}

// Serialize serializes the block into a byte slice using the Gob encoding.
func (b *Block) Serialize() ([]byte, error) {
	result, err := util.GobEncode(b)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// HashTransactions returns the hash of the transactions in the block.
// In Bitcoin, the transactions are hashed in the Merkle tree, allowing for efficient verification
// of the transactions in the block.
func (b *Block) HashTransactions() []byte {
	var txHashes [][]byte

	for _, tx := range b.Transactions {
		txHashes = append(txHashes, tx.ID)
	}
	mTree := NewMerkleTree(txHashes)

	return mTree.RootNode.Data
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
